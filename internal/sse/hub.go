package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Hub struct {
	mu     sync.RWMutex
	subs   map[int64]map[chan []byte]struct{} // trackID -> set of subscribers
	closed bool
}

func NewHub() *Hub {
	return &Hub{
		subs: make(map[int64]map[chan []byte]struct{}),
	}
}

func (h *Hub) Subscribe(trackID int64) (ch chan []byte, unsubscribe func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	ch = make(chan []byte, 16)
	if h.subs[trackID] == nil {
		h.subs[trackID] = make(map[chan []byte]struct{})
	}

	h.subs[trackID][ch] = struct{}{}

	return ch, func() {
		h.mu.Lock()
		defer h.mu.Unlock()
		if h.subs[trackID] != nil {
			delete(h.subs[trackID], ch)
			if len(h.subs[trackID]) == 0 {
				delete(h.subs, trackID)
			}
		}
		close(ch)
	}
}

func (h *Hub) Publish(trackID int64, ev Event) {
	b, err := json.Marshal(ev)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[trackID] {
		select {
		case ch <- b:
		default:
		}
	}
}

func WriteSSE(w http.ResponseWriter, data []byte) error {
	_, err := fmt.Fprintf(w, "data: %s\n\n", data)
	return err
}
