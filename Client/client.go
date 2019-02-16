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
	Rooms []string
	Username string
}

func main() {
	var path_config = flag.String("config", "./client_conf.json", "path to config file")
	flag.Parse()

	var conf Config
	err := ParseConfig(*path_config, &conf)
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
	WriteHandler(conn)
	go ReadHandler(conn)

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

func WriteHandler(conn net.Conn) {
	for {
		input := bufio.NewReader(os.Stdin)
		write := bufio.NewWriterSize(conn, 255)

		text, _ := input.ReadString('\n')
		_, err := write.WriteString(text)
		if err != nil {
			fmt.Println("smth wrong with writer: ", err)
			break
		}
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
				return
			} else {
				fmt.Println("Error reader: ", err)
				return
			}
		}
		fmt.Print("response text: ", text)
	}
}

func Parse(text string) (string, string) {
	fmt.Println(text)
	return "", ""
}