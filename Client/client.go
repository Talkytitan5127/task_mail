package main

import (
	"fmt"
	"encoding/gob"
	"net"
	"os"
)

const (
	host = "127.0.0.1"
	port string = "2233"
	conn_type = "tcp"
)

func main() {
	c, err := net.Dial(conn_type, host+":"+port)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer c.Close()

	msg := "Hello world"
	fmt.Println("Sending ", msg)
	err = gob.NewEncoder(c).Encode(msg)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}