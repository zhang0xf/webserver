package common

import (
	"sync"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	PongWait = 10 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = 5 * time.Second
	// Maximum message size allowed from peer.
	MaxMessageSize = 4096
)

type WaitGroupWrapper struct {
	sync.WaitGroup
}

func (waitGroupWrapper *WaitGroupWrapper) Wrap(cb func()) {
	waitGroupWrapper.Add(1)
	go func() {
		cb()
		waitGroupWrapper.Done()
	}()
}
