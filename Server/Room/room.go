package room

import (
	"errors"
	"sync"
)

type Room struct {
	Users    map[string]bool
	Messages []string
	mux      sync.Mutex
}

func Create_room(users []string) *Room {
	cap := 128
	room := new(Room)
	room.Users = make(map[string]bool)
	for _, name := range users {
		room.Users[name] = true
	}
	room.Messages = make([]string, 0, cap)
	return room
}

func (r *Room) Is_user_in_room(name string) bool {
	r.mux.Lock()
	_, check := r.Users[name]
	r.mux.Unlock()
	return check
}

func (r *Room) Get_messages() []string {
	return r.Messages
}

func (r *Room) Add_message(mes string) {
	if len := len(r.Messages); len >= cap(r.Messages) {
		r.Messages = r.Messages[1:]
		r.Messages = append(r.Messages, mes)
	} else {
		r.Messages = append(r.Messages, mes)
	}
	return
}

func (r *Room) Add_user(name string) error {
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

func (r *Room) Get_users() []string {
	r.mux.Lock()
	var users []string
	for name, _ := range r.Users {
		users = append(users, name)
	}
	r.mux.Unlock()
	return users
}
