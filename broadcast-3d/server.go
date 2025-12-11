package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Server struct {
	node   *maelstrom.Node
	nodeID string

	mu       sync.Mutex
	messages map[int]struct{}

	topology    map[string][]string
	centralNode string

	masterNode bool
}

type BroadcastMessage struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type BroadcastMessageResponse struct {
	Type string `json:"type"`
}

type BroadcastInternalMessage struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type BroadcastInternalMessageResponse struct {
	Type string `json:"type"`
}

type ReadMessage struct {
	Type string `json:"type"`
}

type TopologyMessage struct {
	Type     string              `json:"type"`
	Topology map[string][]string `json:"topology"`
}

type TopologyMessageResponse struct {
	Type string `json:"type"`
}

type ReadMessageResponse struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
}

func NewServer(n *maelstrom.Node) *Server {
	s := &Server{node: n, messages: make(map[int]struct{})}

	s.node.Handle("init", s.initHandler)
	s.node.Handle("broadcast", s.broadcastHandler)
	s.node.Handle("broadcast_internal", s.broadcastInternalHandler)
	s.node.Handle("read", s.readHandler)
	s.node.Handle("topology", s.topologyHandler)

	// no-op handlers
	s.node.Handle("broadcast_ok", s.noOpHandler)
	s.node.Handle("broadcast_internal_ok", s.noOpHandler)

	return s
}

func (s *Server) initHandler(msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body maelstrom.InitMessageBody
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.nodeID = body.NodeID
	log.Printf("Node id set to: %s", s.nodeID)

	return nil
}

func (s *Server) broadcastHandler(msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body BroadcastMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// To avoid cycles: n0->n1->n2->n0
	if _, exists := s.messages[body.Message]; exists {
		broadcastMessageResponse := BroadcastMessageResponse{
			Type: "broadcast_ok",
		}

		return s.node.Reply(msg, broadcastMessageResponse)
	}

	s.messages[body.Message] = struct{}{}

	broadcastInternalMessage := BroadcastInternalMessage{
		Type:    "broadcast_internal",
		Message: body.Message,
	}

	for _, peerID := range s.topology[s.nodeID] {
		go broadcastMessageToPeer(s.node, peerID, broadcastInternalMessage)
	}

	if !s.masterNode {
		// Broadcast to the master node
		go broadcastMessageToPeer(s.node, s.centralNode, broadcastInternalMessage)
	}

	broadcastMessageResponse := BroadcastMessageResponse{
		Type: "broadcast_ok",
	}

	return s.node.Reply(msg, broadcastMessageResponse)
}

func (s *Server) broadcastInternalHandler(msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body BroadcastInternalMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// To avoid cycles: n0->n1->n2->n0
	if _, exists := s.messages[body.Message]; exists {
		broadcastInternalMessageResponse := BroadcastInternalMessageResponse{
			Type: "broadcast_internal_ok",
		}

		return s.node.Reply(msg, broadcastInternalMessageResponse)
	}

	s.messages[body.Message] = struct{}{}

	// To avoid: n0->n0
	for _, peerID := range s.topology[s.nodeID] {
		go broadcastMessageToPeer(s.node, peerID, body)
	}

	broadcastInternalMessageResponse := BroadcastInternalMessageResponse{
		Type: "broadcast_internal_ok",
	}

	return s.node.Reply(msg, broadcastInternalMessageResponse)
}

func broadcastMessageToPeer(node *maelstrom.Node, peerID string, body BroadcastInternalMessage) {
	backoff := time.Millisecond
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err := node.SyncRPC(ctx, peerID, body)
		cancel()

		if err == nil {
			return
		}

		time.Sleep(backoff + time.Duration(rand.Intn(50))*time.Millisecond)
		if backoff < 500*time.Millisecond {
			backoff *= 2
		}
	}
}

func (s *Server) noOpHandler(maelstrom.Message) error {
	return nil
}

func (s *Server) readHandler(msg maelstrom.Message) error {
	var body ReadMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	messages := make([]int, 0, len(s.messages))
	for m := range s.messages {
		messages = append(messages, m)
	}

	readMessageResponse := ReadMessageResponse{
		Type:     "read_ok",
		Messages: messages,
	}

	return s.node.Reply(msg, readMessageResponse)
}

func (s *Server) topologyHandler(msg maelstrom.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var body TopologyMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// We ignore topology sent from maelstrom's controller, at least for now
	topologyMessageResponse := TopologyMessageResponse{
		Type: "topology_ok",
	}

	log.Printf("Received topology information from controller: %v", body.Topology)

	s.topology = topology
	s.centralNode = centralNode

	log.Printf("Using topology: %v, central node: %s", s.topology, s.centralNode)

	if s.nodeID == centralNode {
		s.masterNode = true
	} else {
		s.masterNode = false
	}

	return s.node.Reply(msg, topologyMessageResponse)
}

func (s *Server) Run() error {
	return s.node.Run()
}
