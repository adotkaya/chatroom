package main

import (
	"crypto/rand"
	"encoding/json"
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
	broadcastCH   chan *ReqMsg
}

func NewServer() *Server {
	return &Server{
		clients:       make(map[string]*Client),
		mu:            new(sync.RWMutex),
		joinServerCH:  make(chan *Client, 64),
		leaveServerCH: make(chan *Client, 64),
		broadcastCH:   make(chan *ReqMsg, 64),
	}
}

type MsgType string

const (
	MsgType_Broadcast MsgType = "broadcast"
	MsgType_Join      MsgType = "join"
	MsgType_Leave     MsgType = "leave"
)

type ReqMsg struct {
	Type   MsgType
	Data   string
	Client *Client
}

type RespMsg struct {
	Type     MsgType
	Data     string
	SenderID string
}

func NewRespMsg(msgType MsgType, data string, senderID string) *RespMsg {
	return &RespMsg{
		Type:     msgType,
		Data:     data,
		SenderID: senderID,
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
	go client.readLoop(s)
}

func (s *Server) AcceptLoop() {
	for {
		select {
		case client := <-s.joinServerCH:
			s.joinServer(client)
		case client := <-s.leaveServerCH:
			s.leaveServer(client)
		case msg := <-s.broadcastCH:
			s.broadcast(msg)
		}
	}
}

func (c *Client) readLoop(srv *Server) {
	defer func() {
		c.conn.Close()
		srv.leaveServerCH <- c
	}()
	for {
		_, b, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		msg := new(ReqMsg)
		err = json.Unmarshal(b, msg)
		if err != nil {
			fmt.Printf("Error unmarshalling message: %v\n", err)
			continue
		}
		srv.broadcastCH <- msg
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

func (s *Server) broadcast(msg *ReqMsg) {
	cls := []*Client{}
	s.mu.RLock()
	for _, c := range s.clients {
		if c.ID != msg.Client.ID {
			cls = append(cls, c)
		}
	}
	s.mu.RUnlock()
	resp := NewRespMsg(msg.Type, msg.Data, msg.Client.ID)
	for _, c := range cls {
		err := c.conn.WriteJSON(resp)
		if err != nil {
			fmt.Printf("Error sending message to client %s: %v\n", c.ID, err)
		}
	}
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
