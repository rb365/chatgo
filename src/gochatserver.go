package main

import (
	. "gochatutils"
	"log"
	"net"
//	"reflect"
)

// 1. write client : should be very simple, connect, send, receive
// 2. write tests : create 100000 clients that have state machines and do random things like chat(large and small messages)/leave/join/quit/connect
// 3. regression test! deploy to the digital ocean and run crazy test from local machine maybe even time the average response time
// 4. comments
// 7. robustness

const port string = ":16180"
const userEventBusSize = 1000
const chatRoomEventBusSize = 1000000
const systemEventBusSize = 1000000

func main() {
	eventBus := make(chan Event, systemEventBusSize)
	go runAdministrator(eventBus)
	listenForNewConnections(eventBus)
}

func listenForNewConnections(eventBus chan Event) {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalln("Error happened connecting to port", port, err.Error())
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error happened accepting connection on port", port, err.Error())
			continue
		}
		// added go
		go SendEvent(eventBus, CreateNewConnectionEvent(conn), conn.RemoteAddr().String(), "*")
	}
}

func runAdministrator(eventBus chan Event) {
	//defer log.Println("admin wtf happened????")
	addressToUser := make(map[string]User)
	nameToUser := make(map[string]User)
	nameToChatRoom := make(map[string]ChatRoom)
	for {
		event := <-eventBus
		//log.Println("system received event", reflect.TypeOf(event).String())

		switch event.(type) {

		case NewConnectionEvent:
			evt := event.(NewConnectionEvent)
			user := new(User)
			user.Address = evt.Connection.RemoteAddr().String()
			user.EventBus = make(chan Event, userEventBusSize)
			addressToUser[user.Address] = *user
			go RunUser(evt.Connection, user.EventBus, eventBus)
			break

		case AssignUserNameEvent:
			evt := event.(AssignUserNameEvent)
			user := addressToUser[evt.Address]
			if _, ok := nameToUser[evt.UserName]; !ok {
				user.Name = evt.UserName
				addressToUser[evt.Address] = user
				nameToUser[evt.UserName] = user
			} else {
				evt.UserName = ""
			}
			//added go
			go SendEvent(user.EventBus, evt, "*", user.Address)
			break

		case JoinChatRoomEvent:
			evt := event.(JoinChatRoomEvent)
			user := nameToUser[evt.UserName]
			chatRoom, ok := nameToChatRoom[evt.ChatRoom]
			if !ok {
				chatRoom.EventBus = make(chan Event, chatRoomEventBusSize)
				chatRoom.UserList = make(map[string]User)
				chatRoom.Name = evt.ChatRoom
				nameToChatRoom[chatRoom.Name] = chatRoom
				go RunRoom(chatRoom.Name, chatRoom.EventBus, eventBus)
			}
			// added go
			go SendEvent(chatRoom.EventBus, CreateAddUserEvent(user), "*", chatRoom.Name)
			chatRoom.UserList[user.Name] = user
			break

		case GetChatRoomsEvent:
			evt := event.(GetChatRoomsEvent)
			var roomNames []string
			var userQuantity []int
			for name, chatRoom := range nameToChatRoom {
				roomNames = append(roomNames, name)
				userQuantity = append(userQuantity, len(chatRoom.UserList))
			}
			resultEvent := CreateAvailableChatRoomsEvent(roomNames, userQuantity)
			//added go
			go SendEvent(nameToUser[evt.UserName].EventBus, resultEvent, "*", evt.UserName)
			break

		case LeaveChatRoomEvent:
			evt := event.(LeaveChatRoomEvent)
			chatRoom := nameToChatRoom[evt.ChatRoom]
			//added go
			go SendEvent(chatRoom.EventBus, event, "*", chatRoom.Name)
			delete(chatRoom.UserList, evt.UserName)
			if len(chatRoom.UserList) == 0 {
				//added go
				go SendEvent(chatRoom.EventBus, CreateKillChatRoomEvent(), "*", chatRoom.Name)
				delete(nameToChatRoom, chatRoom.Name)
				//close(chatRoom.EventBus)
			}
			break

		case UserQuitEvent:
			evt := event.(UserQuitEvent)
			user := addressToUser[evt.Address]
			//log.Println("user quit ", user.Name, user.Address)
			if user.Name != "" {
				delete(nameToUser, user.Name)
			}
			delete(addressToUser, user.Address)
			//close(user.EventBus)
			break

		default:
			break
		}
	}
}
