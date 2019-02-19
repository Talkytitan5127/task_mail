package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	room "github.com/task_mail/Server/Room"
)

type Config struct {
	Host, Port, Conn_type string
	Room_name             map[string][]string
}

type User struct {
	conn   net.Conn
	reader *json.Decoder
	writer *json.Encoder
	rooms  map[string]string
}

type Request struct {
	CMD, Username, Room, Message string
}

type Response struct {
	CMD, Status, Error string
}

var (
	Rooms map[string]*room.Room
)

func main() {
	var pathConfig = flag.String("config", "./config.json", "path to config file")
	flag.Parse()

	var conf Config
	err := ParseConfig(*pathConfig, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}

	Rooms = make(map[string]*room.Room)
	for name, users := range conf.Room_name {
		room := room.Create_room(users)
		Rooms[name] = room
	}

	host := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
	listen, err := net.Listen(conf.Conn_type, host)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer listen.Close()

	fmt.Printf("Listening on %s\n", host)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			continue
		}
		defer conn.Close()

		fmt.Println("User connect from:", conn.RemoteAddr())

		user := SetUser(conn)
		go handleRequest(user)
	}

}

func ParseConfig(path string, conf *Config) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&conf)
	if err != nil {
		return err
	}
	return nil
}

func handleRequest(user *User) {
	for {
		data, err := user.ReadPacket()
		if err != nil {
			fmt.Println("User disconnected")
			return
		}
		status, desc := user.Process(data)
		fmt.Println(status)
		user.AnswerClient(data, status, desc)
	}
}

func SetUser(conn net.Conn) *User {
	user := new(User)
	user.conn = conn
	user.reader = json.NewDecoder(conn)
	user.writer = json.NewEncoder(conn)

	reader := json.NewDecoder(conn)
	reader.Decode(&user.rooms)
	return user
}

func (user *User) ReadPacket() (*Request, error) {
	var data Request
	err := user.reader.Decode(&data)
	return &data, err
}

func (user *User) WritePacket(data *Response) {
	user.writer.Encode(&data)
}

func (user *User) Process(data *Request) (string, string) {
	fmt.Println("Process method")
	fmt.Println("server get packet")
	fmt.Printf("%+v\n", data)
	command := data.CMD
	var status, err string
	switch command {
	case "publish":
		status, err = user.Publish(data)
	case "subscribe":
		status, err = user.Subscribe(data)
	case "get_history":
		status, err = "OK", "history will be send"
	default:
		status, err = "ERROR", "unknown command"
	}
	return status, err
}

func (user *User) AnswerClient(data *Request, status, err string) {
	answer := Response{Status: status, Error: err, CMD: data.CMD}
	user.WritePacket(&answer)
	if status != "OK" {
		return
	}
	name_room := data.Room
	switch data.CMD {
	case "subscribe":
		user_conf := map[string]string{"room": name_room, "nickname": user.rooms[name_room]}
		user.writer.Encode(user_conf)
		user.SendHistory(name_room)
	case "get_history":
		user.SendHistory(name_room)
	}
}

func (user *User) Publish(data *Request) (string, string) {
	name_room, message := data.Room, data.Message
	obj_room, ok := Rooms[name_room]
	if ok == false {
		return "ERROR", "Room doesn't exists"
	}

	username := user.rooms[name_room]
	ok = obj_room.Is_user_in_room(username)
	if ok == false {
		return "ERROR", "User doesn't subscribe to this room"
	}

	message = fmt.Sprintf("%s: %s", username, message)
	obj_room.Add_message(message)
	return "OK", "Send message is successful"

}

func (user *User) Subscribe(data *Request) (string, string) {
	name_room, username := data.Room, data.Username
	obj_room, ok := Rooms[name_room]
	if ok == false {
		return "ERROR", "Room doesn't exists"
	}

	err := obj_room.Add_user(username)
	if err != nil {
		return "ERROR", "user already exist's"
	}
	user.rooms[name_room] = username

	return "OK", "user add successful"
}

func (user *User) SendHistory(name_room string) (string, string) {
	obj_room, ok := Rooms[name_room]
	if ok == false {
		return "ERROR", "Room doesn't exists"
	}

	room := obj_room
	messages := room.Get_messages()
	packet := map[string][]string{"history": messages}
	user.writer.Encode(packet)
	return "OK", "history was sent"
}
