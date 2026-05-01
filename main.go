package main

import (
	"crypto/rand"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var (
	WSPort = ":3223"
)

type Client struct {
	ID   string
	mu   *sync.RWMutex
	conn *websocket.Conn
}

func NewClient(id string, conn *websocket.Conn) *Client {
	ID := rand.Text()[:9]
	return &Client{
		ID:   ID,
		mu:   new(sync.RWMutex),
		conn: conn,
	}
}

type Server struct {
	clients []*Client
	mu      *sync.RWMutex
}

func NewServer() *Server {
	return &Server{
		clients: make([]*Client, 0),
		mu:      new(sync.RWMutex),
	}
}

func handleWS(r http.ResponseWriter, w *http.Request) {

}

func main() {
	http.HandleFunc("/", handleWS)
	log.Fatal(http.ListenAndServe(WSPort, nil))
}
