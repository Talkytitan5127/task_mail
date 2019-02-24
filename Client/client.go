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

//Config for Server
type Config struct {
	Host, ConnType string
	Rooms          map[string]string
}

var (
	conf       Config
	PathConfig *string
)

func main() {
	PathConfig = flag.String("config", "./client_conf.json", "path to config file")
	flag.Parse()

	err := ParseConfigFile(*PathConfig, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}

	conn, err := net.Dial(conf.ConnType, conf.Host)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connected to server: ", conn.RemoteAddr())
	SendConfig(conn)

	go ReadHandler(conn)
	WriteHandler(conn)
}

//SendConfig send room's client to server
func SendConfig(conn net.Conn) {
	writer := json.NewEncoder(conn)
	writer.Encode(conf.Rooms)
}

//SaveConfig save current client's rooms to json file
func SaveConfig(path string, conf *Config) {
	fmt.Println("SaveConfig")
	file, err := os.Create(path)
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

//ParseConfigFile read *.json file with server's param
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

//WriteHandler read Stdin
func WriteHandler(conn net.Conn) {
	input := bufio.NewReaderSize(os.Stdin, 255)
	writer := bufio.NewWriterSize(conn, 255)
	for {
		text, _ := input.ReadString('\n')
		if len(text) >= 255 {
			fmt.Println("Wrong len of message\n Type new message")
			continue
		}
		fmt.Print(text)
		if text == "EXIT\n" {
			SaveConfig(*PathConfig, &conf)
			return
		}
		_, err := writer.WriteString(text)
		if err != nil {
			fmt.Println("smth wrong with writer: ", err)
			break
		}
		writer.Flush()
	}
}

//ReadHandler read answers from server
func ReadHandler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		text, err := reader.ReadString('\n')
		text = strings.TrimSuffix(text, "\n")
		if err != nil {
			if err == io.EOF {
				fmt.Println("Server shut down")
				SaveConfig(*PathConfig, &conf)
				conn.Close()
				os.Exit(0)
			} else {
				return
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

//ReadJSON from server
func ReadJSON(conn net.Conn) error {
	data := make(map[string]string)
	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&data)
	if err != nil {
		return err
	}
	roomName := data["room"]
	nickname := data["nickname"]
	conf.Rooms[roomName] = nickname
	return nil
}
