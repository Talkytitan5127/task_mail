package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	room "github.com/task_mail/Server/Room"
)

//Config of server
type Config struct {
	Host, Port, ConnType string
	RoomName             map[string][]string
}

//User data
type User struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	rooms  map[string]string
}

var (
	Rooms map[string]*room.Room
)

func main() {
	var PathConfig = flag.String("config", "./config.json", "path to config file")
	flag.Parse()

	var conf Config
	err := ParseConfig(*PathConfig, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}

	Rooms = make(map[string]*room.Room)
	for name, users := range conf.RoomName {
		room := room.Create_room(users)
		Rooms[name] = room
	}

	host := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
	listen, err := net.Listen(conf.ConnType, host)
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

func handleRequest(user *User) {
	for {
		text, err := user.ReadMessage()
		if err != nil {
			return
		}

		fmt.Println(text)
		text = strings.TrimSuffix(text, "\n")
		err = CheckInput(text)
		if err != nil {
			fmt.Println(err)
			user.WriteMessage("wrong input command")
			continue
		}
		status := user.Process(text)
		fmt.Println(status)
	}
}

//CheckInput control input text
func CheckInput(text string) error {
	data := strings.Split(text, " ")
	if text == "" || text == "\n" || len(data) < 2 {
		return errors.New("wrong input command")
	}
	return nil
}

//SetUser get json of user's config
func SetUser(conn net.Conn) *User {
	user := new(User)
	user.conn = conn
	user.reader = bufio.NewReader(conn)
	user.writer = bufio.NewWriter(conn)

	reader := json.NewDecoder(conn)
	reader.Decode(&user.rooms)
	return user
}

//ReadMessage process packet from Client
func (user *User) ReadMessage() (string, error) {
	text, err := user.reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	if err != nil {
		if err == io.EOF {
			fmt.Printf("User %s disonnected\n", user.conn.RemoteAddr())
		} else {
			fmt.Println("Error readnig: ", err.Error())
		}
		return "", errors.New("smth wrong")
	}
	return text, nil
}

//WriteMessage to Client
func (user *User) WriteMessage(message string) {
	user.writer.WriteString(message + "\n")
	user.writer.Flush()
}

//ParseConfig of Server
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

//Process Client's command
func (user *User) Process(text string) string {
	fmt.Println("Process method")
	data := strings.Split(text, " ")
	command, NameRoom := data[0], data[1]
	fmt.Println(command, "=>", NameRoom)
	var status string
	switch command {
	case "publish":
		status = user.Publish(NameRoom)
		return status
	case "subscribe":
		status = user.Subscribe(NameRoom)
		return status
	default:
		user.WriteMessage("unknown command")
		return "unknown command"
	}
}

//GetHistory get all message from room
func (user *User) GetHistory(NameRoom string) string {
	room := Rooms[NameRoom]
	messages := room.Get_messages()
	user.writer.WriteString(fmt.Sprintf("----%s----\n", NameRoom))
	for _, message := range messages {
		user.writer.WriteString(message + "\n")
	}
	user.writer.WriteString(strings.Repeat("-", (8+len(NameRoom))) + "\n")
	user.writer.Flush()
	return "History was sent"
}

//Publish message to room
func (user *User) Publish(NameRoom string) string {
	obj_room, ok := Rooms[NameRoom]
	if ok == false {
		return "Room doesn't exists"
	}

	username := user.rooms[NameRoom]
	ok = obj_room.Is_user_in_room(username)
	if ok == false {
		return "User doesn't subscribe to this room"
	}

	user.WriteMessage("Type message to send")
	fmt.Println("Wait message from user")
	text, err := user.ReadMessage()
	if err != nil {
		return "Message error"
	}
	text = fmt.Sprintf("%s: %s", username, text)
	obj_room.Add_message(text)
	return "Send message is successful"

}

//Subscribe Client to room
func (user *User) Subscribe(NameRoom string) string {
	ObjRoom, ok := Rooms[NameRoom]
	if ok == false {
		return "Room doesn't exists"
	}

	user.WriteMessage("Enter you nickname")
	fmt.Println("Wait nickname from user")
	username, err := user.ReadMessage()
	if err != nil {
		return "Message error"
	}

	err = ObjRoom.Add_user(username)
	if err != nil {
		user.WriteMessage("user already exist's")
		return "subscribe failed"
	}
	user.WriteMessage("Subscribe successful")
	user.rooms[NameRoom] = username
	user.GetHistory(NameRoom)

	user.WriteMessage("JSON")
	data := map[string]string{"room": NameRoom, "nickname": username}
	json.NewEncoder(user.conn).Encode(&data)
	return "user add successful"
}
