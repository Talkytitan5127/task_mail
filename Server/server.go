package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	room "github.com/task_mail/Server/Room"
)

//Config Server
type Config struct {
	Host, Port, ConnType string
	Room_name            map[string][]string
}

//User parametres
type User struct {
	conn   net.Conn
	reader *json.Decoder
	writer *json.Encoder
	rooms  map[string]string
}

//Request packet
type Request struct {
	CMD, Username, Room, Message string
}

//Response packet
type Response struct {
	CMD, Status, Error string
}

//History packet
type History struct {
	Room     string
	Messages []string
}

var (
	conf        Config
	pathConfig  *string
	Rooms       map[string]*room.Room
	UserConnect []*User
)

func main() {
	pathConfig = flag.String("config", "./config.json", "path to config file")
	flag.Parse()

	err := ParseConfig(*pathConfig, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}

	Rooms = make(map[string]*room.Room)
	for name, users := range conf.Room_name {
		room := room.CreateRoom(users)
		Rooms[name] = room
	}

	host := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
	listen, err := net.Listen(conf.ConnType, host)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer listen.Close()
	SetupCloseHandler(listen)
	UserConnect = make([]*User, 0)

	fmt.Printf("Listening on %s\n", host)
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			return
		}
		defer conn.Close()

		fmt.Println("User connect from:", conn.RemoteAddr())

		user := SetUser(conn)
		UserConnect = append(UserConnect, user)

		user.SendHistoryWhenConnect()
		go handleRequest(user)
	}

}

//SetupCloseHandler to save Room config
func SetupCloseHandler(listen net.Listener) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		defer listen.Close()
		SaveConfig(*pathConfig, &conf)
		os.Exit(0)
	}()
}

//SaveConfig save current client's rooms to json file
func SaveConfig(path string, conf *Config) {
	fmt.Println("SaveConfig")
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()
	NewRoom := make(map[string][]string)
	for nameRoom, obj := range Rooms {
		NewRoom[nameRoom] = obj.GetUsers()
	}
	conf.Room_name = NewRoom
	encoder := json.NewEncoder(file)
	err = encoder.Encode(&conf)
	if err != nil {
		fmt.Println(err)
	}

}

//ParseConfig of Server connection
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
		fmt.Println(status, desc)
		user.AnswerClient(data, status, desc)
	}
}

//SetUser set parametres for connected user
func SetUser(conn net.Conn) *User {
	user := new(User)
	user.conn = conn
	user.reader = json.NewDecoder(conn)
	user.writer = json.NewEncoder(conn)

	user.reader.Decode(&user.rooms)
	if user.rooms == nil {
		user.rooms = make(map[string]string)
	}
	return user
}

//ReadPacket from Client
func (user *User) ReadPacket() (*Request, error) {
	var data Request
	err := user.reader.Decode(&data)
	return &data, err
}

//WritePacket to Client
func (user *User) WritePacket(data *Response) {
	user.writer.Encode(&data)
}

//Process command from Client
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
		status, err = user.CheckHistory(data)
	default:
		status, err = "ERROR", "unknown command"
	}
	return status, err
}

//AnswerClient after Process
func (user *User) AnswerClient(data *Request, status, err string) {
	fmt.Println("Answer")
	answer := Response{Status: status, Error: err, CMD: data.CMD}
	user.WritePacket(&answer)
	if status != "OK" {
		return
	}
	fmt.Printf("%+v\n", data)
	NameRoom := data.Room
	switch data.CMD {
	case "subscribe":
		UserConf := map[string]string{"room": NameRoom, "nickname": user.rooms[NameRoom]}
		user.writer.Encode(&UserConf)
	case "get_history":
		user.SendHistory(data)
	case "publish":
		SendMessageToSub(data)
	}
}

//Publish message to Room
func (user *User) Publish(data *Request) (string, string) {
	NameRoom, message := data.Room, data.Message
	var mux sync.Mutex
	mux.Lock()
	ObjRoom, ok := Rooms[NameRoom]
	mux.Unlock()
	if ok == false {
		return "ERROR", "Room doesn't exists"
	}
	username := user.rooms[NameRoom]
	ok = ObjRoom.IsUserInRoom(username)
	if ok == false {
		return "ERROR", "User doesn't subscribe to this room"
	}

	message = fmt.Sprintf("%s: %s", username, message)
	ObjRoom.AddMessage(message)
	return "OK", "Send message is successful"

}

//SendMessageToSub send mes to all subscribers of room
func SendMessageToSub(data *Request) {
	NameRoom := data.Room
	message := Rooms[NameRoom].GetLastMessage()
	packet := History{Room: NameRoom, Messages: []string{message}}
	for _, user := range UserConnect {
		_, ok := user.rooms[NameRoom]
		if ok != true {
			continue
		}
		answer := Response{Status: "OK", Error: "OK", CMD: "get_message"}
		user.WritePacket(&answer)
		user.writer.Encode(&packet)
	}

}

//Subscribe to Room
func (user *User) Subscribe(data *Request) (string, string) {
	NameRoom, username := data.Room, data.Username
	var mux sync.Mutex
	mux.Lock()
	ObjRoom, ok := Rooms[NameRoom]
	mux.Unlock()
	if ok == false {
		return "ERROR", "Room doesn't exists"
	}

	err := ObjRoom.AddUser(username)
	if err != nil {
		return "ERROR", "user already exist's"
	}
	user.rooms[NameRoom] = username
	return "OK", "user add successful"
}

//CheckHistory and Room
func (user *User) CheckHistory(data *Request) (string, string) {
	NameRoom := data.Room
	var mux sync.Mutex
	mux.Lock()
	_, ok := Rooms[NameRoom]
	mux.Unlock()
	if ok == false {
		return "ERROR", "Room doesn't exists"
	}
	return "OK", "history was sent"

}

//SendHistory to Client
func (user *User) SendHistory(data *Request) (string, string) {
	NameRoom := data.Room
	ObjRoom, _ := Rooms[NameRoom]
	room := ObjRoom
	messages := room.GetMessages()
	packet := History{Room: NameRoom, Messages: messages}
	err := user.writer.Encode(&packet)
	if err != nil {
		fmt.Println(err)
		return "ERROR", "couldn't send json"
	}
	return "OK", "history was sent"
}

//SendHistoryWhenConnect to Client
func (user *User) SendHistoryWhenConnect() {
	for name := range user.rooms {
		req := Request{Room: name}
		answer := Response{Status: "new_connect", Error: "", CMD: "get_history"}
		user.WritePacket(&answer)
		fmt.Println(user.SendHistory(&req))
	}
}
