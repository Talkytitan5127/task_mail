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
)

type Config struct {
	Host, Port, Conn_type string
	Room_name map[string]string
}

type User struct {
	conn net.Conn
	rooms []map[string]string
}

var (
	rooms map[string]*room.Room
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

	rooms = make(map[string]*room.Room)
	for _, name := range conf.Room_name {
		room := room.Create_room(name)
		rooms[name] = room
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
		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	for {
		reader := bufio.NewReader(conn)
		text, err := reader.ReadString('\n')
		
		if err != nil {
			if err == io.EOF {
				fmt.Println("User disonnected")
			} else {
				fmt.Println("Error readnig: ", err.Error())
			}
			return
		}
		writer := bufio.NewWriter(conn)

		fmt.Println(text)
		text = strings.TrimSuffix(text, "\n")
		status := Process(text)
		writer.WriteString(status + "\n")
		writer.Flush()
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

func ParseText(text string) (string, string) {
	data := strings.SplitN(text, " ", 2)
	fmt.Printf("%v\n", data)
	return data[0], data[1]
}

func Process(text string) string {
	fmt.Println("Process method")
	command, data := ParseText(text)
	fmt.Println(command,"=>", data)
	switch command {
	case "publish":
		//Publish(data)
		return "publish"
	case "subscribe":
		//subscribe(data)
		return "subscribe"
	default:
		return "unknown command"
	}
}

func Publish(text string) {
	//room_name, data := ParseText(text)

}

func subscribe(text string) {
	//room_name, data := ParseText(text)
}