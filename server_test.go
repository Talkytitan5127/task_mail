package tests

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func RunServer() {
	gopath := os.Getenv("GOPATH")
	cmd := exec.Command(gopath+"/bin/Server", "--config=./Server/config.json")
	err := cmd.Start()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
		os.Exit(1)
	}
}

func TestMain(m *testing.M) {
	RunServer()
	exitcode := m.Run()
	defer fmt.Println("shutdown")
	os.Exit(exitcode)
}

func TestConnection(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:2233")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()
}

type User struct {
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
	rooms  map[string]string
}

func SetUser(conn net.Conn) *User {
	user := new(User)
	user.conn = conn
	user.reader = bufio.NewReader(conn)
	user.writer = bufio.NewWriter(conn)

	emptydict := map[string]string{"loby": "pavel"}
	json.NewEncoder(conn).Encode(&emptydict)
	return user
}

func (user *User) ReadMessage() (string, error) {
	text, err := user.reader.ReadString('\n')
	text = strings.TrimSuffix(text, "\n")
	if err != nil {
		return "", errors.New("something wrong")
	}
	return text, nil
}

func (user *User) WriteMessage(message string) {
	user.writer.WriteString(message + "\n")
	user.writer.Flush()
}

func TestSending(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:2233")
	if err != nil {
		t.Error(err)
		os.Exit(1)
	}
	defer conn.Close()
	user := SetUser(conn)
	user.WriteMessage("hello world")

	response, err := user.ReadMessage()
	if err != nil {
		t.Error("something wrong")
	}
	if response != "unknown command" {
		t.Error("server not get message")
	}
}

func TestSubscribe(t *testing.T) {
	conn, err := net.Dial("tcp", "127.0.0.1:2233")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()
	user := SetUser(conn)
	user.WriteMessage("subscribe kitchen")
	response, _ := user.ReadMessage()
	if response != "Enter you nickname" {
		t.Error("room \"kitchen\" not exist")
	}
	user.WriteMessage("peter")

	response, _ = user.ReadMessage()
	if response != "Subscribe successful" {
		t.Error("nickname already occupied")
	}

	response, _ = user.ReadMessage()
	if response != "JSON" {
		t.Error("server don't send word JSON")
	}

	data := make(map[string]string)
	json.NewDecoder(user.conn).Decode(&data)

	fmt.Printf("%+v\n", data)
	if data["room"] != "kitchen" && data["nickname"] != "peter" {
		t.Error("subscribe didn't work\n")
	}

}
