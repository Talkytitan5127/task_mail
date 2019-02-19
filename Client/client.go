package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

type Config struct {
	Host, Conn_type string
	Rooms           map[string]string
}

type Request struct {
	CMD      string
	Username string
	Room     string
	Message  string
}

type Response struct {
	CMD    string
	Status string
	Error  string
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

	fmt.Println("Connected to server: ", conn.RemoteAddr())
	conf.SendPacket(conn)

	go ReadHandler(conn)
	WriteHandler(conn)
}

func (conf Config) SendPacket(conn net.Conn) {
	writer := json.NewEncoder(conn)
	writer.Encode(conf.Rooms)
}

func (r *Request) SendPacket(conn net.Conn) {
	writer := json.NewEncoder(conn)
	writer.Encode(r)
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
	for {
		text, _ := input.ReadString('\n')
		request, status := ParseText(text)
		if status != "" {
			fmt.Println("Error: ", status)
			break
		}
		request.SendPacket(conn)
	}
}

func ReadHandler(conn net.Conn) {
	var resp *Response
	reader := json.NewDecoder(conn)
	for {
		reader.Decode(&resp)
		status := resp.Status
		if status == "ERROR" {
			fmt.Println(resp.Error)
			continue
		}
		fmt.Println(status)
		switch resp.CMD {
		case "get_history":
			PrintHistory(conn)
		case "subscribe":
			PrintSub(conn)
		default:
			continue
		}

	}
}

func PrintSub(conn net.Conn) {
	var resp map[string]string
	json.NewDecoder(conn).Decode(&resp)
	conf.Rooms[resp["room"]] = resp["nickname"]
	PrintHistory(conn)

}

func PrintHistory(conn net.Conn) {
	var history []string
	json.NewDecoder(conn).Decode(history)
	output := "----history----\n"
	for _, mes := range history {
		output += (mes + "\n")
	}
	output += "---------------\n"
	fmt.Print(output)
}

func ParseText(text string) (*Request, string) {
	if len(text) == 0 {
		return nil, "wrong input"
	}
	data := strings.SplitN(text, " ", 2)
	if len(data) < 2 {
		return nil, "not enough argument"
	}
	fmt.Println("data: ", data)
	var req Request
	switch command := data[0]; command {
	case "publish":
		data = strings.SplitAfterN(data[1], " ", 2)
		if len(data) < 2 {
			return nil, "not enough argument"
		}
		if mes := data[1]; len(mes) > 254 {
			return nil, "message > 255"
		}
		req = Request{CMD: command, Room: data[0], Message: data[1]}
		return &req, ""
	case "subscribe":
		data = strings.SplitAfterN(data[1], " ", 2)
		if len(data) < 2 {
			return nil, "not enough argument"
		}
		req = Request{CMD: command, Room: data[0], Username: data[1]}
		return &req, ""
	case "get_history":
		req = Request{CMD: command, Room: data[1]}
		return &req, ""
	default:
		return nil, "unknown command"
	}
}
