package main

import (
	"fmt"
	"net"
	"os"
	"bufio"
	"flag"
	"encoding/json"
	"github.com/task_mail/Server/Room"
	"io"
	"strings"
	"errors"
)

type Config struct {
	Host, Port, Conn_type string
	Room_name map[string][]string
}

type User struct {
	conn net.Conn
	rooms map[string]string
}

var (
	Rooms map[string]*room.Room
)

func main() {
	var path_config = flag.String("config", "./config.json", "path to config file")
	flag.Parse()

	var conf Config
	err := ParseConfig(*path_config, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}

	Rooms = make(map[string]*room.Room)
	for name, users := range conf.Room_name {
		room := room.Create_room(users)
		Rooms[name] = room
	}
	
	l, err := net.Listen(conf.Conn_type, conf.Host+":"+conf.Port)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Printf("Listening on %s:%s\n", conf.Host, conf.Port)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		defer conn.Close()

		fmt.Println(conn.RemoteAddr())

		user := SetUser(conn)
		go handleRequest(user)
	}

}

func handleRequest(user *User) {
	conn := user.conn
	for {
		text, err := ReadMessage(conn)
		if err != nil {
			return
		}

		fmt.Println(text)
		text = strings.TrimSuffix(text, "\n")
		err = CheckInput(text)
		if err != nil {
			fmt.Println(err)
			WriteMessage(conn, "wrong input command")
			continue
		}
		status := Process(user, text)
		fmt.Println(status)
	}
}

func CheckInput(text string) error {
	data := strings.Split(text, " ")
	if text == "" || text == "\n" || len(data) < 2 {
		return errors.New("wrong input command")
	}
	return nil
}

func SetUser(conn net.Conn) *User {
	user := new(User)
	user.conn = conn

	reader := json.NewDecoder(conn)
	reader.Decode(&user.rooms)
	return user
}

func ReadMessage(conn net.Conn) (string, error) {
	reader := bufio.NewReader(conn)
	text, err := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	if err != nil {
		if err == io.EOF {
			fmt.Println("User disonnected")
		} else {
			fmt.Println("Error readnig: ", err.Error())
		}
		return "", errors.New("smth wrong")
	}
	return text, nil
}

func WriteMessage(conn net.Conn, message string) {
	writer := bufio.NewWriter(conn)
	writer.WriteString(message + "\n")
	writer.Flush()
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

func Process(user *User, text string) string {
	fmt.Println("Process method")
	data := strings.Split(text, " ")
	command, name_room := data[0], data[1]
	fmt.Println(command,"=>", name_room)
	var status string
	switch command {
	case "publish":
		status = user.Publish(name_room)
		return status
	case "subscribe":
		status = user.Subscribe(name_room)
		return status
	case "get_history":
		status = user.Get_History(name_room)
		return status
	default:
		return "unknown command"
	}
}

func (user *User) Get_History(name_room string) string {
	room := Rooms[name_room]
	messages := room.Get_messages()
	writer := bufio.NewWriter(user.conn)
	writer.WriteString("----"+name_room+"----\n")
	for _, message := range(messages) {
		writer.WriteString(message + "\n")
	}
	writer.WriteString(strings.Repeat("-", (8+len(name_room)))+"\n")
	writer.Flush()
	return "History was sent"
}

func (user *User) Publish(name_room string) string {
	obj_room, ok := Rooms[name_room]
	if ok == false {
		return "Room doesn't exists"
	}

	username := user.rooms[name_room]
	ok = obj_room.Is_user_in_room(username)
	if ok == false {
		return "User doesn't subscribe to this room"
	}

	WriteMessage(user.conn, "Type message to send")
	fmt.Println("Wait message from user")
	text, err := ReadMessage(user.conn)
	if err != nil {
		return "Message error"
	}

	obj_room.Add_message(text)
	return "Send message is successful"
	 
}

func (user *User) Subscribe(name_room string) string {
	obj_room, ok := Rooms[name_room]
	if ok == false {
		return "Room doesn't exists"
	}

	WriteMessage(user.conn, "Enter you nickname")
	fmt.Println("Wait nickname from user")
	username, err := ReadMessage(user.conn)
	if err != nil {
		return "Message error"
	}

	err = obj_room.Add_user(username)
	if err != nil {
		WriteMessage(user.conn, "Type message to send")
		return "subscribe failed"
	}
	WriteMessage(user.conn, "Subscribe successful")
	user.rooms[name_room] = username

	return "user add successful"
}