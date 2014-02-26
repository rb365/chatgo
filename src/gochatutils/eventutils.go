package gochatutils

import (
	"log"
	"net"
	"reflect"
)

type Event interface {
}

var debugEventsEnabled bool = false

func SendEvent(eventBus chan Event, event Event, from string, to string, message ...string) {
	if debugEventsEnabled {
		log.Println(from, "->", to, reflect.TypeOf(event).String(), message, " SENDING...")
	}
	eventBus <- event
	if debugEventsEnabled {
		// in case we need to debug clogged channels
		log.Println(from, "->", to, reflect.TypeOf(event).String(), message, " SENT")
	}
}

// ---------------------------
// NEW CONNECTION EVENT
// ---------------------------

type NewConnectionEvent struct {
	Connection net.Conn
}

func CreateNewConnectionEvent(conn net.Conn) NewConnectionEvent {
	var event NewConnectionEvent
	event.Connection = conn
	return event
}

// ---------------------------
// INPUT EVENT
// ---------------------------

type InputEvent struct {
	Data []byte
}

func CreateInputEvent(data []byte) InputEvent {
	var event InputEvent
	event.Data = data
	return event
}

// ---------------------------
// OUTPUT EVENT
// ---------------------------

type OutputEvent struct {
	Data []byte
	From string
}

func CreateOutputEvent(from string, data []byte) OutputEvent {
	var event OutputEvent
	event.Data = data
	event.From = from
	return event
}

// ---------------------------
// ASSIGN USER NAME EVENT
// ---------------------------

type AssignUserNameEvent struct {
	UserName string
	Address  string
}

func CreateAssignUserNameEvent(userName string, address string) AssignUserNameEvent {
	var event AssignUserNameEvent
	event.UserName = userName
	event.Address = address
	return event
}

// ---------------------------
// GET CHAT ROOMS EVENT
// ---------------------------

type GetChatRoomsEvent struct {
	UserName string
}

func CreateGetChatRoomsEvent(userName string) GetChatRoomsEvent {
	var event GetChatRoomsEvent
	event.UserName = userName
	return event
}

// ---------------------------
// LEAVE CHAT ROOM EVENT
// ---------------------------

type LeaveChatRoomEvent struct {
	UserName string
	ChatRoom string
}

func CreateLeaveChatRoomEvent(userName string, chatRoom string) LeaveChatRoomEvent {
	var event LeaveChatRoomEvent
	event.UserName = userName
	event.ChatRoom = chatRoom
	return event
}

// ---------------------------
// USER QUIT EVENT
// ---------------------------

type UserQuitEvent struct {
	Address string
}

func CreateUserQuitEvent(address string) UserQuitEvent {
	var event UserQuitEvent
	event.Address = address
	return event
}

// ---------------------------
// AVAILABLE CHAT ROOMS EVENT
// ---------------------------

type AvailableChatRoomsEvent struct {
	RoomNames    []string
	UserQuantity []int
}

func CreateAvailableChatRoomsEvent(roomNames []string, userQuantity []int) AvailableChatRoomsEvent {
	var event AvailableChatRoomsEvent
	event.RoomNames = roomNames
	event.UserQuantity = userQuantity
	return event
}

// ---------------------------
// JOIN CHAT ROOM EVENT
// ---------------------------

type JoinChatRoomEvent struct {
	UserName string
	ChatRoom string
}

func CreateJoinChatRoomEvent(userName string, chatRoom string) JoinChatRoomEvent {
	var event JoinChatRoomEvent
	event.UserName = userName
	event.ChatRoom = chatRoom
	return event
}

// ---------------------------
// CONNECTION CLOSED EVENT
// ---------------------------

type ConnectionClosedEvent struct {
}

func CreateConnectionClosedEvent() ConnectionClosedEvent {
	var event ConnectionClosedEvent
	return event
}

// ---------------------------
// JOINED CHAT ROOM EVENT
// ---------------------------

type JoinedChatRoomEvent struct {
	ChatRoom         string
	UserList         []string
	ChatRoomEventBus chan Event
}

func CreateJoinedChatRoomEvent(chatRoom string, userList []string, chatRoomEventBus chan Event) JoinedChatRoomEvent {
	var event JoinedChatRoomEvent
	event.ChatRoom = chatRoom
	event.UserList = userList
	event.ChatRoomEventBus = chatRoomEventBus
	return event
}

// ---------------------------
// KILL CHAT ROOM EVENT
// ---------------------------

type KillChatRoomEvent struct {
}

func CreateKillChatRoomEvent() KillChatRoomEvent {
	var event KillChatRoomEvent
	return event
}

// ---------------------------
// ADD USER EVENT
// ---------------------------

type AddUserEvent struct {
	User User
}

func CreateAddUserEvent(user User) AddUserEvent {
	var event AddUserEvent
	event.User = user
	return event
}

// ---------------------------
// NO OP EVENT
// ---------------------------

type NoOpEvent struct {
}
