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
	clients       []*Client
	mu            *sync.RWMutex
	joinServerCH  chan *Client
	leaveServerCH chan *Client
}

func NewServer() *Server {
	return &Server{
		clients: []*Client{},
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
	fmt.Println(len(s.clients))
	s.mu.Unlock()
}

func (s *Server) AcceptLoop() {
	select {
	case client := <-s.joinServerCH:
		s.mu.Lock()
		s.clients = append(s.clients, client)
		s.mu.Unlock()
	case client := <-s.leaveServerCH:
		s.mu.Lock()
		for i, c := range s.clients {
			if c.ID == client.ID {
				s.clients = append(s.clients[:i], s.clients[i+1:]...)
				break
			}
		}
		s.mu.Unlock()
	}
}

func createWSServer() {
	s := NewServer()
	http.HandleFunc("/", s.handleWS)

	fmt.Printf("starting server on port : %s\n", WSPort)
	log.Fatal(http.ListenAndServe(WSPort, nil))
}

func main() {
	createWSServer()
}
