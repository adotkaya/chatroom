package main

import (
	"fmt"
	"log"
	"testing"

	"github.com/gorilla/websocket"
)

type TestConfig struct {
	clientCount int
}

func DialServer() {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial("ws://localhost%s", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("connected to server", conn.LocalAddr().String())
	defer conn.Close()

}

func TestConnection(t *testing.T) {
	config := TestConfig{
		clientCount: 10,
	}
	for i := 0; i < config.clientCount; i++ {
		go DialServer()
	}
}
