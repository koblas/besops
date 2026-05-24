package broadcast

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHubPublishSubscribe(t *testing.T) {
	hub := NewHub(8)

	sub := hub.Subscribe()

	hub.Publish(Event{Type: "heartbeat", Data: map[string]string{"id": "m1"}})

	select {
	case ev := <-sub:
		assert.Equal(t, "heartbeat", ev.Type)
		assert.Equal(t, map[string]string{"id": "m1"}, ev.Data)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestHubMultipleSubscribers(t *testing.T) {
	hub := NewHub(8)

	sub1 := hub.Subscribe()
	sub2 := hub.Subscribe()

	hub.Publish(Event{Type: "test", Data: "hello"})

	for _, sub := range []<-chan Event{sub1, sub2} {
		select {
		case ev := <-sub:
			assert.Equal(t, "test", ev.Type)
			assert.Equal(t, "hello", ev.Data)
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for event")
		}
	}
}

func TestHubUnsubscribe(t *testing.T) {
	hub := NewHub(8)

	sub := hub.Subscribe()
	hub.Unsubscribe(sub)

	hub.Publish(Event{Type: "after-unsub", Data: nil})

	select {
	case _, ok := <-sub:
		require.False(t, ok, "channel should be closed after unsubscribe")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("channel should be closed immediately")
	}
}

func TestHubDropsWhenFull(t *testing.T) {
	hub := NewHub(2)

	sub := hub.Subscribe()

	hub.Publish(Event{Type: "1"})
	hub.Publish(Event{Type: "2"})
	hub.Publish(Event{Type: "3"}) // should be dropped

	ev1 := <-sub
	ev2 := <-sub

	assert.Equal(t, "1", ev1.Type)
	assert.Equal(t, "2", ev2.Type)

	select {
	case <-sub:
		t.Fatal("third event should have been dropped")
	default:
	}
}
