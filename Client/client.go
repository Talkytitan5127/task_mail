package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

type Config struct {
	Host, Conn_type string
	Rooms           map[string]string
}

var (
	conf Config
)

func main() {
	var path_config = flag.String("config", "./client_conf.json", "path to config file")
	flag.Parse()

	err := ParseConfigFile(*path_config, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}

	conn, err := net.Dial(conf.Conn_type, conf.Host)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()
	//not working
	//defer SaveConfig(*path_config, &conf)
	fmt.Println("Connected to server: ", conn.RemoteAddr())
	SendConfig(conn)

	go ReadHandler(conn)
	WriteHandler(conn)
}

func SendConfig(conn net.Conn) {
	writer := json.NewEncoder(conn)
	writer.Encode(conf.Rooms)
}

func SaveConfig(path string, conf *Config) {
	fmt.Println("SaveConfig")
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&conf)
	if err != nil {
		fmt.Println(err)
	}

}

func ParseConfigFile(path string, conf *Config) error {
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

func WriteHandler(conn net.Conn) {
	input := bufio.NewReaderSize(os.Stdin, 255)
	writer := bufio.NewWriterSize(conn, 255)
	for {
		text, _ := input.ReadString('\n')
		if len(text) >= 255 {
			fmt.Println("Wrong len of message\n Type new message")
			continue
		}
		_, err := writer.WriteString(text)
		if err != nil {
			fmt.Println("smth wrong with writer: ", err)
			break
		}
		writer.Flush()
	}
}

func ReadHandler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		text, err := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if err != nil {
			if err == io.EOF {
				fmt.Println("Server shut down")
				conn.Close()
				os.Exit(0)
			} else {
				fmt.Println("Error reader: ", err)
			}
		}
		switch text {
		case "JSON":
			err = ReadJson(conn)
			if err != nil {
				fmt.Println(err)
			}
		default:
			fmt.Println(text)
		}

	}
}

func ReadJson(conn net.Conn) error {
	data := make(map[string]string)
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&data)
	if err != nil {
		return err
	}
	room_name := data["room"]
	nickname := data["nickname"]
	conf.Rooms[room_name] = nickname
	return nil
}
