package main

import (
	"encoding/json"
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Server struct {
	node   *maelstrom.Node
	nodeID string
}

type InitMessageResponse struct {
	Type string `json:"type"`
}

type EchoMessage struct {
	Type  string `json:"type"`
	MsgID int64  `json:"msg_id"`
	Echo  string `json:"echo"`
}

type EchoMessageResponse struct {
	Type  string `json:"type"`
	MsgID int64  `json:"msg_id"`
	Echo  string `json:"echo"`
}

func NewServer(n *maelstrom.Node) *Server {
	s := &Server{node: n}

	s.node.Handle("init", s.initHandler)
	s.node.Handle("echo", s.echoHandler)

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

func (s *Server) echoHandler(msg maelstrom.Message) error {
	var body EchoMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	echoMessageResponse := EchoMessageResponse{
		Type:  "echo_ok",
		MsgID: body.MsgID,
		Echo:  body.Echo,
	}

	return s.node.Reply(msg, echoMessageResponse)
}

func (s *Server) Run() error {
	return s.node.Run()
}
