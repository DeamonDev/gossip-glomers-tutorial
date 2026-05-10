package main

import (
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	log.SetOutput(os.Stderr)

	n := maelstrom.NewNode()

	s := NewServer(n)

	if err := s.Run(); err != nil {
		log.Fatal(err)
	}
}
