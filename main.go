package main

import (
	"crypto/rand"
	"fmt"
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

func NewClient(conn *websocket.Conn) *Client {
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

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  512,
		WriteBufferSize: 512,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Error on HTTP conn upgrade %v\n", err)
		return
	}
	client := NewClient(conn)
	s.mu.Lock()
	s.clients = append(s.clients, client)
	s.mu.Unlock()
}

func main() {
	s := NewServer()
	http.HandleFunc("/", s.handleWS)

	fmt.Printf("starting server on port : %s\n", WSPort)
	log.Fatal(http.ListenAndServe(WSPort, nil))
}
