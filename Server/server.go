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
	reader *bufio.Reader
	writer *bufio.Writer
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
	
	host := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
	l, err := net.Listen(conf.Conn_type, host)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Printf("Listening on %s\n", host)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
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
	user.reader = bufio.NewReader(conn)
	user.writer = bufio.NewWriter(conn)

	reader := json.NewDecoder(conn)
	reader.Decode(&user.rooms)
	return user
}

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

func (user *User) WriteMessage(message string) {
	user.writer.WriteString(message + "\n")
	user.writer.Flush()
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

func (user *User) Process(text string) string {
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
	user.writer.WriteString(fmt.Sprintf("----%s----\n", name_room))
	for _, message := range(messages) {
		user.writer.WriteString(message + "\n")
	}
	user.writer.WriteString(strings.Repeat("-", (8+len(name_room)))+"\n")
	user.writer.Flush()
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

func (user *User) Subscribe(name_room string) string {
	obj_room, ok := Rooms[name_room]
	if ok == false {
		return "Room doesn't exists"
	}

	user.WriteMessage("Enter you nickname")
	fmt.Println("Wait nickname from user")
	username, err := user.ReadMessage()
	if err != nil {
		return "Message error"
	}

	err = obj_room.Add_user(username)
	if err != nil {
		user.WriteMessage("Type message to send")
		return "subscribe failed"
	}
	user.WriteMessage("Subscribe successful")
	user.rooms[name_room] = username
	user.Get_History(name_room)
	return "user add successful"
}