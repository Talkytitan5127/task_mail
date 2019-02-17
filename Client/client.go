package main

import (
	"fmt"
	"encoding/json"
	"net"
	"os"
	"flag"
	"bufio"
	"io"
)

type Config struct {
	Host, Conn_type string
	Rooms map[string]string
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
	fmt.Println(conf)
	conn, err := net.Dial(conf.Conn_type, conf.Host)
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

func SendConfig(conn net.Conn) {
	writer := json.NewEncoder(conn)
	writer.Encode(conf.Rooms)
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
	for {
		input := bufio.NewReader(os.Stdin)
		writer := bufio.NewWriterSize(conn, 255)

		text, _ := input.ReadString('\n')
		_, err := writer.WriteString(text)
		if err != nil {
			fmt.Println("smth wrong with writer: ", err)
			break
		}
		writer.Flush()
	}
	//command, data = Parse(text)
}

func ReadHandler(conn net.Conn) {
	for {
		reader := bufio.NewReader(conn)
		text, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Server shut down")
				conn.Close()
				os.Exit(0)
			} else {
				fmt.Println("Error reader: ", err)
			}
		}
		fmt.Print("Response:\n", text)
	}
}

func Parse(text string) (string, string) {
	fmt.Println(text)
	return "", ""
}