package gochatutils

import (
//	"log"
	"net"
//	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type User struct {
	EventBus chan Event
	Address  string
	Name     string
}

type UserState int

const (
	AskName UserState = iota + 1
	WaitForNameInput
	WaitForNameApproval
	WaitToJoinChatRoom
	NotInChatRoom
	InChatRoom
)

var maxMessageSize = 1000

func isControl(c byte) bool {
	return !(c >= 32 && c != 127)
}

func getCommands(input string) []string {
	input = strings.TrimSpace(input)
	commands := strings.Split(input, " ")
	var result []string
	for i := 0; i < len(commands); i++ {
		commands[i] = strings.TrimSpace(commands[i])
		if len(commands[i]) > 0 {
			result = append(result, commands[i])
		}
	}
	return result
}

func handleConnection(con net.Conn, eventBus chan Event) {
	//defer log.Println("wtf happened????")
	data := make([]byte, maxMessageSize)
	dataNoCtrlChar := make([]byte, maxMessageSize)
	address := con.RemoteAddr().String()
	for {
		n, err := con.Read(data)
		if err != nil {
			// added go
			go SendEvent(eventBus, CreateConnectionClosedEvent(), address, address, err.Error())
			return
		}
		j := 0
		for i := 0; i < n; i++ {
			if !isControl(data[i]) {
				dataNoCtrlChar[j] = data[i]
				j++
			}
		}
		if j > 0 {
			SendEvent(eventBus, CreateInputEvent(dataNoCtrlChar[:j]), address, address, "received", strconv.Itoa(j), "bytes")
			//need to sleep to let other goroutines run and limit amount of data that any one user can send
			time.Sleep(30 * time.Millisecond)
		}
	}
}

func RunUser(connection net.Conn, userEventBus chan Event, systemEventBus chan Event) {
	//defer log.Println("wtf happened????")
	go handleConnection(connection, userEventBus)

	var nop NoOpEvent
	var userName, chatRoomName string
	var chatRoomEventBus chan Event
	address := connection.RemoteAddr().String()

	userState := AskName

	SendEvent(userEventBus, nop, address, address)

	connection.Write([]byte("Welcome to the Golang chat server\n"))

	for {
		event := <-userEventBus
		//log.Println(address, "received event", reflect.TypeOf(event).String(), "in state", userState)

		if evt, ok := event.(OutputEvent); ok {
			from := evt.From
			if from != "*" {
				from += ":"
			}
			connection.Write([]byte(from + " "))
			connection.Write(evt.Data)
			connection.Write([]byte("\n"))
			continue
		} else if _, ok := event.(ConnectionClosedEvent); ok {
			if chatRoomEventBus != nil {
				// added go
				go SendEvent(systemEventBus, CreateLeaveChatRoomEvent(userName, chatRoomName), address, "*")
			}
			// added go
			go SendEvent(systemEventBus, CreateUserQuitEvent(connection.RemoteAddr().String()), address, "*")
			return
		}

		switch userState {

		case AskName:
			connection.Write([]byte("Login Name?\n"))
			userState = WaitForNameInput
			break

		case WaitForNameInput:
			if evt, ok := event.(InputEvent); ok {
				newUserName := strings.TrimSpace(string(evt.Data))
				if len(newUserName) > 0 {
					resultEvent := CreateAssignUserNameEvent(newUserName, connection.RemoteAddr().String())
					//added go
					go SendEvent(systemEventBus, resultEvent, address, "*")
					userState = WaitForNameApproval
				} else {
					connection.Write([]byte("User name cannot be blank or all white space!\n"))
					//added go
					SendEvent(userEventBus, nop, address, address)
					userState = AskName
				}
			}
			break

		case WaitForNameApproval:
			if evt, ok := event.(AssignUserNameEvent); ok {
				if evt.UserName == "" {
					connection.Write([]byte("Sorry, name taken.\n"))
					// added go
					SendEvent(userEventBus, nop, address, address)
					userState = AskName
				} else {
					userName = evt.UserName
					connection.Write([]byte("Welcome " + userName + "!\n"))
					userState = NotInChatRoom
				}
			}
			break

		case NotInChatRoom:
			if evt, ok := event.(InputEvent); ok {
				command := getCommands(string(evt.Data))
				if len(command) == 0 {
					continue
				}
				switch command[0] {
				case "/rooms":
					// added go
					go SendEvent(systemEventBus, CreateGetChatRoomsEvent(userName), address, "*")
					break
				case "/join":
					if len(command) > 1 {
						// added go
						go SendEvent(systemEventBus, CreateJoinChatRoomEvent(userName, command[1]), address, "*")
						userState = WaitToJoinChatRoom
					} else {
						connection.Write([]byte("Chat room name cannot be empty!\n"))
					}
					break
				case "/quit":
					if chatRoomEventBus != nil {
						// added go
						go SendEvent(systemEventBus, CreateLeaveChatRoomEvent(userName, chatRoomName), address, "*")
					}
					// added go
					go SendEvent(systemEventBus, CreateUserQuitEvent(userName), address, "*")
					connection.Write([]byte("BYE\n"))
					connection.Close()
					break
				default:
					connection.Write([]byte("Unknown command!\n"))
				}
			} else if evt, ok := event.(AvailableChatRoomsEvent); ok {
				connection.Write([]byte("Active rooms are:\n"))
				for i := 0; i < len(evt.RoomNames); i++ {
					connection.Write([]byte("* " + evt.RoomNames[i] + " (" + strconv.Itoa(evt.UserQuantity[i]) + ")\n"))
				}
				connection.Write([]byte("end of list.\n"))
			}
			break

		case WaitToJoinChatRoom:
			if evt, ok := event.(JoinedChatRoomEvent); ok {
				connection.Write([]byte("Entering room:" + evt.ChatRoom + "\n"))
				evt.UserList = append(evt.UserList, userName)
				sort.Strings(evt.UserList)
				for i := 0; i < len(evt.UserList); i++ {
					output := "* " + evt.UserList[i]
					if userName == evt.UserList[i] {
						output += " (** this is you)"
					}
					connection.Write([]byte(output + "\n"))
				}
				connection.Write([]byte("end of list.\n"))
				chatRoomName = evt.ChatRoom
				chatRoomEventBus = evt.ChatRoomEventBus
				userState = InChatRoom
			}
			break

		case InChatRoom:
			if evt, ok := event.(InputEvent); ok {
				command := getCommands(string(evt.Data))
				if len(command) == 0 {
					// still need to report something and not crash
					command = make([]string, 1)
				}
				switch command[0] {
				case "/leave":
					// added go
					go SendEvent(systemEventBus, CreateLeaveChatRoomEvent(userName, chatRoomName), address, "*")
					chatRoomEventBus = nil
					chatRoomName = ""
					userState = NotInChatRoom
					break
				default:
					go SendEvent(chatRoomEventBus, CreateOutputEvent(userName, evt.Data), address, chatRoomName)
					break
				}
			}
			break
		}
	}
}
