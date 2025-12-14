package main

import (
	"log"
	"sync"
	"time"
)

type Batcher struct {
	mu        sync.Mutex
	batches   map[string][]int
	ticker    *time.Ticker
	flushChan chan FlushEvent
}

type FlushEvent struct {
	PeerID   string
	Messages []int
}

func NewBatcher(batchTimeout time.Duration) *Batcher {
	return &Batcher{
		ticker:    time.NewTicker(batchTimeout),
		batches:   make(map[string][]int),
		flushChan: make(chan FlushEvent),
	}
}

func (b *Batcher) Run() {
	for range b.ticker.C {
		b.mu.Lock()
		for peerID, messages := range b.batches {
			if len(messages) > 0 {
				b.flushChan <- FlushEvent{
					PeerID:   peerID,
					Messages: messages,
				}
			}
		}
		b.batches = make(map[string][]int)
		b.mu.Unlock()
	}
}

func (b *Batcher) Add(peerID string, message int) {
	b.mu.Lock()
	b.batches[peerID] = append(b.batches[peerID], message)
	b.mu.Unlock()
}

func (b *Batcher) Close() {
	log.Printf("Closing batcher")

	b.ticker.Stop()
	close(b.flushChan)
}
