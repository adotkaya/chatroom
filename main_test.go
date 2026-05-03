package main

import (
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var (
	host = "ws://localhost"
)

type TestConfig struct {
	clientCount int
	wg          *sync.WaitGroup
}

func DialServer(wg *sync.WaitGroup) {
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.Dial(fmt.Sprintf("%s%s", host, WSPort), nil)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		conn.Close()
		wg.Done()
	}()

	fmt.Println("connected to server", conn.LocalAddr().String())
	time.Sleep(1 * time.Second)

}

func TestConnection(t *testing.T) {
	go createWSServer()
	time.Sleep(1 * time.Second)
	config := TestConfig{
		clientCount: 10,
		wg:          new(sync.WaitGroup),
	}
	config.wg.Add(config.clientCount)
	for range config.clientCount {
		go DialServer(config.wg)
	}
	config.wg.Wait()
	fmt.Println("Test exit")
}
