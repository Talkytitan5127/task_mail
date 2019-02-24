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

//Config connection
type Config struct {
	Host, Port, ConnType string
	Rooms                map[string]string
}

//Request packet
type Request struct {
	CMD, Username, Room, Message string
}

//Response packet
type Response struct {
	CMD, Status, Error string
}

//History packet
type History struct {
	Room     string
	Messages []string
}

var (
	conf       Config
	pathConfig *string
	writer     *json.Encoder
	reader     *json.Decoder
)

func main() {
	pathConfig = flag.String("config", "./client_conf.json", "path to config file")
	flag.Parse()

	err := ParseConfigFile(*pathConfig, &conf)
	if err != nil {
		fmt.Println("Error config: ", err.Error())
		os.Exit(1)
	}
	host := fmt.Sprintf("%s:%s", conf.Host, conf.Port)
	conn, err := net.Dial(conf.ConnType, host)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("Connected to server: ", conn.RemoteAddr())
	writer = json.NewEncoder(conn)
	reader = json.NewDecoder(conn)

	conf.SendPacket(conn)

	go ReadHandler(conn)
	WriteHandler(conn)
}

//SendPacket with user settings to Server
func (conf Config) SendPacket(conn net.Conn) {
	writer.Encode(conf.Rooms)
}

//SendPacket request to Server
func (r *Request) SendPacket(conn net.Conn) {
	writer.Encode(r)
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

//ParseConfigFile with conn's setting
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

//WriteHandler from stdin to server
func WriteHandler(conn net.Conn) {
	input := bufio.NewReaderSize(os.Stdin, 255)
	for {
		text, _ := input.ReadString('\n')
		request, status := ParseText(text)
		if status != "" {
			fmt.Println("Error: ", status)
			continue
		}
		request.SendPacket(conn)
	}
}

//ReadHandler from Server
func ReadHandler(conn net.Conn) {
	var resp *Response
	for {
		err := reader.Decode(&resp)
		if err != nil {
			if err == io.EOF {
				fmt.Println("server shut down")
				SaveConfig(*pathConfig, &conf)
				os.Exit(0)
			}
			fmt.Println(err)
		}
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
			room := GetSubConfig(conn)
			mes := fmt.Sprintf("get_history %s\n", room)
			req, _ := ParseText(mes)
			req.SendPacket(conn)
		default:
			continue
		}
	}
}

//GetSubConfig of new subscribe room
func GetSubConfig(conn net.Conn) string {
	fmt.Println("GetSubConfig")
	var resp map[string]string
	reader.Decode(&resp)
	conf.Rooms[resp["room"]] = resp["nickname"]
	return resp["room"]
}

//PrintHistory of Room
func PrintHistory(conn net.Conn) {
	var data *History
	err := reader.Decode(&data)
	if err != nil {
		fmt.Println(err)
		return
	}
	output := fmt.Sprintf("----%s----\n", data.Room)
	for _, mes := range data.Messages {
		output += (mes + "\n")
	}
	output += strings.Repeat("-", 8+(len(data.Room))) + "\n"
	fmt.Print(output)
}

//ParseText from stdin
func ParseText(text string) (*Request, string) {
	text = strings.TrimSuffix(text, "\n")
	if len(text) == 0 {
		return nil, "wrong input"
	}
	data := strings.SplitN(text, " ", 2)
	if len(data) < 2 {
		return nil, "not enough argument"
	}

	var req Request
	switch command := data[0]; command {
	case "publish":
		data = strings.SplitN(data[1], " ", 2)
		if len(data) < 2 {
			return nil, "not enough argument"
		}
		if mes := data[1]; len(mes) > 254 {
			return nil, "message > 255"
		}

		req = Request{CMD: command, Room: data[0], Message: data[1]}
		return &req, ""
	case "subscribe":
		data = strings.SplitN(data[1], " ", 2)
		if len(data) < 2 {
			return nil, "not enough argument"
		}
		_, ok := conf.Rooms[data[0]]
		if ok == true {
			return nil, "you already subscribed to this room"
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
