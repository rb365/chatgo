package gochatutils

import (
	//"log"
	//"reflect"
)

type ChatRoom struct {
	EventBus chan Event
	Name     string
	UserList map[string]User
}

func RunRoom(roomName string, eventBus chan Event, systemEventBus chan Event) {
	userList := make(map[string]User)

	for {
		event := <-eventBus
		//log.Println(roomName, "received event", reflect.TypeOf(event).String())

		switch event.(type) {

		case OutputEvent:
			for _, user := range userList {
				go SendEvent(user.EventBus, event, roomName, user.Address)
			}
			break

		case AddUserEvent:
			evt := event.(AddUserEvent)
			var userNames []string
			for _, user := range userList {
				outputEvent := CreateOutputEvent("*", []byte("new user joined chat: "+evt.User.Name))
				go SendEvent(user.EventBus, outputEvent, roomName, user.Address)
				userNames = append(userNames, user.Name)
			}
			userList[evt.User.Name] = evt.User
			resultEvent := CreateJoinedChatRoomEvent(roomName, userNames, eventBus)
			go SendEvent(evt.User.EventBus, resultEvent, roomName, evt.User.Address)
			break

		case LeaveChatRoomEvent:
			evt := event.(LeaveChatRoomEvent)
			for _, user := range userList {
				message := "user has left chat: " + evt.UserName
				if user.Name == evt.UserName {
					message += " (** this is you)"
				}
				go SendEvent(user.EventBus, CreateOutputEvent("*", []byte(message)), roomName, user.Address)
			}
			delete(userList, evt.UserName)
			break

		case KillChatRoomEvent:
			return
			break

		default:
			break
		}
	}
}
