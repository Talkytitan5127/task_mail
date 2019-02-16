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
)

type Config struct {
	Host, Port, Conn_type string
	Room_name []string
}

func main() {
	var path_config = flag.String("config", "./config.json", "path to config file")
	flag.Parse()

	var conf Config
	err := ParseConfig(*path_config, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}

	var rooms = make(map[string]*room.Room)
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
		fmt.Print("received ", text)
		conn.Write([]byte("Message received."))
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