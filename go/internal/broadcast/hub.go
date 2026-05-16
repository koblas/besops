package broadcast

import "sync"

// Hub is a fan-out broadcaster that delivers published events to all active subscribers.
// It is safe for concurrent use.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[chan Event]struct{}
	bufferSize  int
}

// NewHub creates a broadcaster with the given per-subscriber buffer size.
// If a subscriber's buffer is full, events are dropped for that subscriber.
func NewHub(bufferSize int) *Hub {
	return &Hub{
		subscribers: make(map[chan Event]struct{}),
		bufferSize:  bufferSize,
	}
}

func (h *Hub) Publish(event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.subscribers {
		select {
		case ch <- event:
		default:
			// subscriber too slow, drop event
		}
	}
}

func (h *Hub) Subscribe() <-chan Event {
	ch := make(chan Event, h.bufferSize)

	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()

	return ch
}

func (h *Hub) Unsubscribe(sub <-chan Event) {
	// We need the underlying chan to delete from the map.
	// Since Subscribe returns the same channel, the caller passes it back.
	h.mu.Lock()
	for ch := range h.subscribers {
		if sub == ch {
			delete(h.subscribers, ch)
			close(ch)
			break
		}
	}
	h.mu.Unlock()
}
