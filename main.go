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
	clients       map[string]*Client
	mu            *sync.RWMutex
	joinServerCH  chan *Client
	leaveServerCH chan *Client
}

func NewServer() *Server {
	return &Server{
		clients:       make(map[string]*Client),
		mu:            new(sync.RWMutex),
		joinServerCH:  make(chan *Client, 64),
		leaveServerCH: make(chan *Client, 64),
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
		fmt.Printf("Error on HTTP connection upgrade to Websocket %v\n", err)
		return
	}
	client := NewClient(conn)
	s.mu.Lock()
	s.joinServerCH <- client
	s.mu.Unlock()
}

func (s *Server) AcceptLoop() {
	for {
		select {
		case client := <-s.joinServerCH:
			s.joinServer(client)
		case client := <-s.leaveServerCH:
			s.leaveServer(client)
		}
	}
}

func (s *Server) joinServer(client *Client) {
	s.mu.Lock()
	s.clients[client.ID] = client
	fmt.Printf("Client %s joined the server: \n", client.ID)
	s.mu.Unlock()
}

func (s *Server) leaveServer(client *Client) {
	s.mu.Lock()
	delete(s.clients, client.ID)
	fmt.Printf("Client %s left the server: \n", client.ID)
	s.mu.Unlock()
}

func createWSServer() {
	s := NewServer()
	go s.AcceptLoop()
	http.HandleFunc("/", s.handleWS)

	fmt.Printf("starting server on port : %s\n", WSPort)
	log.Fatal(http.ListenAndServe(WSPort, nil))
}

func main() {
	createWSServer()
}
