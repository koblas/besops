package broadcast

// Event represents a server-sent event with a named type and JSON-serializable payload.
type Event struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Publisher pushes events into the broadcast system.
type Publisher interface {
	Publish(event Event)
}

// Subscriber receives events from the broadcast system.
type Subscriber interface {
	// Subscribe returns a channel that receives broadcast events.
	// The caller must call Unsubscribe when done to avoid leaking resources.
	Subscribe() <-chan Event
	Unsubscribe(ch <-chan Event)
}

// Broadcaster combines publishing and subscribing.
type Broadcaster interface {
	Publisher
	Subscriber
}
