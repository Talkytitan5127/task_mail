package main

import (
	"fmt"
	"net"
	"os"
	"encoding/gob"
	"flag"
	"encoding/json"
)

type Config struct {
	Host, Port, Conn_type string
	Rooms []string
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
		fmt.Println(conn.RemoteAddr())
		go handleRequest(conn)
	}

}

func handleRequest(conn net.Conn) {
	defer conn.Close()
	
	var msg string
	decoder := gob.NewDecoder(conn)
	err := decoder.Decode(&msg)
	if err != nil {
		fmt.Println("Error readnig: ", err.Error())
		return
	}
	fmt.Println("received ", msg)
	conn.Write([]byte("Message received."))
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