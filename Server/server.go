package main

import (
	"fmt"
	"net"
	"os"
	"encoding/gob"
)

var (
	host string = "127.0.0.1"
	port string = "2233"
	conn_type string = "tcp"
)

func main() {
	l, err := net.Listen(conn_type, host + ":" + port)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
		os.Exit(1)
	}
	defer l.Close()

	fmt.Printf("Listening on %s:%s\n", host, port)

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

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
	}
	fmt.Println("received ", msg)
	conn.Write([]byte("Message received."))
}