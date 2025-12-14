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

	topology   map[string][]string
	masterNode string

	role string

	batcher *Batcher
}

type BroadcastMessage struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type BroadcastMessageResponse struct {
	Type string `json:"type"`
}

type BroadcastInternalMessage struct {
	Type     string `json:"type"`
	Messages []int  `json:"messages"`
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
	b := NewBatcher(200 * time.Millisecond)
	s := &Server{node: n, messages: make(map[int]struct{}), batcher: b}

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

func (s *Server) handleFlushes() {
	for event := range s.batcher.flushChan {
		msg := BroadcastInternalMessage{
			Type:     "broadcast_internal",
			Messages: event.Messages,
		}
		go broadcastMessageToPeer(s.node, event.PeerID, msg)
	}
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

	for _, peerID := range s.topology[s.nodeID] {
		s.batcher.Add(peerID, body.Message)
	}

	if s.role == "FOLLOWER" {
		// Broadcast to the master node
		s.batcher.Add(s.masterNode, body.Message)
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

	var unseenMessages []int

	for _, m := range body.Messages {
		if _, exists := s.messages[m]; !exists {
			unseenMessages = append(unseenMessages, m)
			s.messages[m] = struct{}{}
		}
	}

	if len(unseenMessages) == 0 {
		broadcastInternalMessageResponse := BroadcastInternalMessageResponse{
			Type: "broadcast_internal_ok",
		}

		return s.node.Reply(msg, broadcastInternalMessageResponse)
	}

	for _, m := range unseenMessages {
		for _, peerID := range s.topology[s.nodeID] {
			s.batcher.Add(peerID, m)
		}
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
	s.masterNode = masterNode

	log.Printf("Using topology: %v, central node: %s", s.topology, s.masterNode)

	if s.nodeID == masterNode {
		s.role = "LEADER"
	} else {
		s.role = "FOLLOWER"
	}

	return s.node.Reply(msg, topologyMessageResponse)
}

func (s *Server) Run() error {
	go s.handleFlushes()
	go s.batcher.Run()

	return s.node.Run()
}

func (s *Server) Close() {
	log.Printf("Closing server")

	s.batcher.Close()
}
