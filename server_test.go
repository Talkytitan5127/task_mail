package server_test

// before run test, run server

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
)

var (
	host   string = "127.0.0.1"
	port   string = "2233"
	ctype  string = "tcp"
	reader *json.Decoder
	writer *json.Encoder
)

type Request struct {
	CMD, Username, Room, Message string
}

type Response struct {
	CMD, Status, Error string
}

type History struct {
	Room     string
	Messages []string
}

func TestConnection(t *testing.T) {
	hostname := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial(ctype, hostname)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
}

func SendConfig(conn net.Conn) {
	data := map[string]string{"loby": "pavel"}
	json.NewEncoder(conn).Encode(&data)
}

func TestGetGreetingHistory(t *testing.T) {
	hostname := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial(ctype, hostname)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()

	ConnectServer(conn, "no message yet", t)
}

func ConnectServer(conn net.Conn, mes string, t *testing.T) {
	SendConfig(conn)
	var history *History
	var resp *Response
	writer := json.NewDecoder(conn)
	writer.Decode(&resp)
	if resp.Status != "new_connect" || resp.CMD != "get_history" {
		t.Error("incorrect response")
	}
	writer.Decode(&history)
	if history.Room != "loby" || history.Messages[0] != mes {
		t.Error("incorrect history packet")
	}

}

func TestPublish(t *testing.T) {
	hostname := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial(ctype, hostname)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	ConnectServer(conn, "no message yet", t)

	req := Request{CMD: "publish", Room: "loby", Message: "hello world"}
	json.NewEncoder(conn).Encode(&req)

	var resp *Response
	json.NewDecoder(conn).Decode(&resp)
	if resp.Status != "OK" || resp.CMD != "publish" {
		t.Error("status != OK")
	}
}

func TestSubscribe(t *testing.T) {
	hostname := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.Dial(ctype, hostname)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	ConnectServer(conn, "pavel: hello world", t)

	req := Request{CMD: "subscribe", Room: "kitchen", Username: "butcher"}
	json.NewEncoder(conn).Encode(&req)

	var resp *Response
	json.NewDecoder(conn).Decode(&resp)
	if resp.Status != "OK" || resp.CMD != "subscribe" {
		t.Error("status != OK")
	}

	var data map[string]string
	json.NewDecoder(conn).Decode(&data)
	if data["nickname"] != "butcher" || data["room"] != "kitchen" {
		t.Error("error subscribe")
	}
}
