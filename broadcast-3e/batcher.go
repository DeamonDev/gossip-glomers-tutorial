package main

type Batcher struct {
	flushChan chan FlushEvent
}

type FlushEvent struct {
	PeerID   string
	Messages []int
}
