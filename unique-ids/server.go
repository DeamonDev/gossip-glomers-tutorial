package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Server struct {
	node    *maelstrom.Node
	nodeID  string
	mu      sync.Mutex
	counter uint64
}

type InitMessageResponse struct {
	Type string `json:"type"`
}

type GenerateMessage struct {
	Type string `json:"type"`
}

type GenerateMessageResponse struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

func NewServer(n *maelstrom.Node) *Server {
	s := &Server{node: n, counter: 0}

	s.node.Handle("init", s.initHandler)
	s.node.Handle("generate", s.generateHandler)

	return s
}

func (s *Server) initHandler(msg maelstrom.Message) error {
	var body maelstrom.InitMessageBody
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.nodeID = body.NodeID

	log.Printf("Node id set to: %s", s.nodeID)

	initMessageResponse := InitMessageResponse{Type: "init_ok"}

	return s.node.Reply(msg, initMessageResponse)
}

func (s *Server) generateHandler(msg maelstrom.Message) error {
	var body GenerateMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("%s-%d", s.nodeID, s.counter)
	generateMessageResponse := GenerateMessageResponse{
		Type: "generate_ok",
		Id:   id,
	}

	s.counter++

	return s.node.Reply(msg, generateMessageResponse)
}

func (s *Server) Run() error {
	return s.node.Run()
}
