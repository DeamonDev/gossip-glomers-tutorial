package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	s := NewServer(n)

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
