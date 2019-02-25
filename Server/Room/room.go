package room

import (
	"errors"
	"sync"
)

//Room struct
type Room struct {
	Users    map[string]bool
	Messages []string
	mux      sync.Mutex
}

//CreateRoom function
func CreateRoom(users []string) *Room {
	cap := 128
	room := new(Room)
	room.Users = make(map[string]bool)
	for _, name := range users {
		room.Users[name] = true
	}
	room.Messages = make([]string, 0, cap)
	return room
}

//IsUserInRoom check user exist in room
func (r *Room) IsUserInRoom(name string) bool {
	r.mux.Lock()
	_, check := r.Users[name]
	r.mux.Unlock()
	return check
}

//GetMessages from room
func (r *Room) GetMessages() []string {
	if len(r.Messages) == 0 {
		return []string{"no message yet"}
	}
	return r.Messages
}

//GetLastMessage from room
func (r *Room) GetLastMessage() string {
	return r.Messages[len(r.Messages)-1]
}

//AddMessage to room
func (r *Room) AddMessage(mes string) {
	if len := len(r.Messages); len >= cap(r.Messages) {
		r.Messages = r.Messages[1:]
		r.Messages = append(r.Messages, mes)
	} else {
		r.Messages = append(r.Messages, mes)
	}
	return
}

//AddUser to room
func (r *Room) AddUser(name string) error {
	r.mux.Lock()
	defer r.mux.Unlock()
	_, ok := r.Users[name]
	if ok {
		return errors.New("user already exist's")
	} else {
		r.Users[name] = true
		return nil
	}
}

//GetUsers from Room
func (r *Room) GetUsers() []string {
	r.mux.Lock()
	var users []string
	for name := range r.Users {
		users = append(users, name)
	}
	r.mux.Unlock()
	return users
}
