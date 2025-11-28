package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Server struct {
	node   *maelstrom.Node
	nodeID string
	peers  []string

	mu       sync.Mutex
	messages map[int]struct{}
}

type BroadcastMessage struct {
	Type    string `json:"type"`
	Message int    `json:"message"`
}

type BroadcastMessageResponse struct {
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
	s.node.Handle("read", s.readHandler)
	s.node.Handle("topology", s.topologyHandler)

	// no-op handlers
	s.node.Handle("broadcast_ok", s.noOpHandler)

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

	var peers []string
	for _, peerID := range body.NodeIDs {
		if peerID != s.nodeID {
			peers = append(peers, peerID)
		}
	}

	s.peers = peers
	log.Printf("Discovered cluster peers: %v", s.peers)

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

	// To avoid: n0->n0
	for _, peerID := range s.peers {
		err := s.node.Send(peerID, body)
		if err != nil {
			log.Printf("Failed to broadcast message to node: %s", peerID)
			continue
		}
	}

	broadcastMessageResponse := BroadcastMessageResponse{
		Type: "broadcast_ok",
	}

	return s.node.Reply(msg, broadcastMessageResponse)
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
	var body TopologyMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	// We ignore topology sent from maelstrom's controller, at least for now
	topologyMessageResponse := TopologyMessageResponse{
		Type: "topology_ok",
	}

	log.Printf("Received topology information from controller: %v", body.Topology)

	return s.node.Reply(msg, topologyMessageResponse)
}

func (s *Server) Run() error {
	return s.node.Run()
}
