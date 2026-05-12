package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

const counterKey = "counter"

type Server struct {
	node   *maelstrom.Node
	kv     *maelstrom.KV
	nodeID string
}

type ReadMessage struct {
	Type string `json:"type"`
}

type ReadMessageResponse struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type AddMessage struct {
	Type  string `json:"type"`
	Delta int    `json:"delta"`
}

type AddMessageResponse struct {
	Type string `json:"type"`
}

func NewServer(n *maelstrom.Node) *Server {
	kv := maelstrom.NewSeqKV(n)
	s := &Server{node: n, kv: kv}

	s.node.Handle("init", s.initHandler)
	s.node.Handle("read", s.readHandler)
	s.node.Handle("add", s.addHandler)

	return s
}

func (s *Server) initHandler(msg maelstrom.Message) error {
	var body maelstrom.InitMessageBody
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	s.nodeID = body.NodeID

	log.Printf("Node id set to: %s", s.nodeID)

	return nil
}

func (s *Server) readHandler(msg maelstrom.Message) error {
	var body ReadMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	fenceKey := fmt.Sprintf("fence-%d", rand.Int())
	if err := s.kv.Write(ctx, fenceKey, 1); err != nil {
		return err
	}

	counter, err := s.readCounter(ctx)
	if err != nil {
		return err
	}

	log.Printf("Read counter: %d", counter)

	return s.node.Reply(msg, ReadMessageResponse{Type: "read_ok", Value: counter})
}

func (s *Server) addHandler(msg maelstrom.Message) error {
	var body AddMessage
	if err := json.Unmarshal(msg.Body, &body); err != nil {
		return err
	}

	ctx := context.Background()
	log.Printf("Adding delta: %d", body.Delta)
	for {
		cur, err := s.readCounter(ctx)
		if err != nil {
			return err
		}

		err = s.kv.CompareAndSwap(ctx, counterKey, cur, cur+body.Delta, true)
		if err == nil {
			log.Printf("CAS succeeded: %d -> %d", cur, cur+body.Delta)
			break
		}
		if rpcErr, ok := errors.AsType[*maelstrom.RPCError](err); ok && rpcErr.Code == maelstrom.PreconditionFailed {
			log.Printf("CAS conflict at %d, retrying", cur)
			continue
		}
		return err
	}

	return s.node.Reply(msg, AddMessageResponse{Type: "add_ok"})
}

func (s *Server) readCounter(ctx context.Context) (int, error) {
	val, err := s.kv.ReadInt(ctx, counterKey)
	if err != nil {
		if rpcErr, ok := errors.AsType[*maelstrom.RPCError](err); ok && rpcErr.Code == maelstrom.KeyDoesNotExist {
			return 0, nil
		}
		return 0, err
	}
	return val, nil
}

func (s *Server) Run() error {
	return s.node.Run()
}
