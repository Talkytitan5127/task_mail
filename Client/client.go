package main

import (
	"fmt"
	"encoding/json"
	"net"
	"os"
	"flag"
	"bufio"
	"strings"
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

	InputHandler(conn)

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

func InputHandler(conn net.Conn) {
	fmt.Println("Connected to server: ", conn.RemoteAddr())

	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	fmt.Println(text)
	//command, data = Parse(text)
}

func Parse(text string) (string, string) {
	fmt.Println(text)
	return "", ""
}