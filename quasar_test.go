package quasar

import (
	"reflect"
	"testing"
)

type MockNetwork struct {
	peers          []pubkey
	connections    map[pubkey][]pubkey
	updateChannels map[pubkey]chan update
	eventChannels  map[pubkey]chan event
}

type MockOverlay struct {
	peer    pubkey
	network MockNetwork
}

func (mqt *MockOverlay) Id() pubkey {
	return mqt.peer
}

func (mqt *MockOverlay) ConnectedPeers() []pubkey {
	return mqt.network.connections[mqt.peer]
}

func (mqt *MockOverlay) ReceivedEventChannel() chan event {
	return mqt.network.eventChannels[mqt.peer]

}

func (mqt *MockOverlay) ReceivedUpdateChannel() chan update {
	return mqt.network.updateChannels[mqt.peer]
}

func (mqt *MockOverlay) SendEvent(peer *pubkey, e event) {
	mqt.network.eventChannels[*peer] <- e
}

func (mqt *MockOverlay) SendUpdate(peer *pubkey, index uint, filter []byte) {
	u := update{peer: &mqt.peer, index: index, filter: filter}
	mqt.network.updateChannels[*peer] <- u
}

func (mqt *MockOverlay) Start() {}

func (mqt *MockOverlay) Stop() {}

func TestNewEvent(t *testing.T) {
	ttl := uint32(42)
	event := newEvent([]byte("test topic"), []byte("test message"), ttl)
	if event == nil {
		t.Errorf("Expected event!")
	}
	if string(event.message) != "test message" {
		t.Errorf("event message not set!")
	}
	if event.ttl != ttl {
		t.Errorf("event ttl not set!")
	}
	if event.publishers == nil && len(event.publishers) != 0 {
		t.Errorf("event publishers not set!")
	}
	if event.topicDigest == nil {
		t.Errorf("event digest not set!")
	}
}

func TestSubscriptions(t *testing.T) {
	q := NewQuasar(nil, &Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     180,
		PropagationInterval: 60,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.000001,
		FiltersDepth:        1024,
		FiltersM:            8192, // m 1k
		FiltersK:            6,    // hashes
	})

	a := make(chan []byte)
	q.Subscribe([]byte("a"), a)
	subs := q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a")}) {
		t.Errorf("Incorrect subscriptions! ", subs)
	}

	b1 := make(chan []byte)
	q.Subscribe([]byte("b"), b1)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions! ", subs)
	}

	b2 := make(chan []byte)
	q.Subscribe([]byte("b"), b2)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions! ", subs)
	}

	c1 := make(chan []byte)
	q.Subscribe([]byte("c"), c1)
	subs = q.Subscriptions()
	expectedSubs := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	if !checkSubs(subs, expectedSubs) {
		t.Errorf("Incorrect subscriptions!", subs)
	}

	c2 := make(chan []byte)
	q.Subscribe([]byte("c"), c2)
	subs = q.Subscriptions()
	expectedSubs = [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	if !checkSubs(subs, expectedSubs) {
		t.Errorf("Incorrect subscriptions!", subs)
	}

	// test clears all if no receiver provided
	q.Unsubscribe([]byte("c"), nil)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}

	// only clears specific receiver
	if len(q.Subscribers([]byte("b"))) != 2 {
		t.Errorf("Incorrect subscribers!")
	}
	q.Unsubscribe([]byte("b"), b1)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}
	if len(q.Subscribers([]byte("b"))) != 1 {
		t.Errorf("Incorrect subscribers!")
	}

	// clears key when last receiver removed
	q.Unsubscribe([]byte("b"), b2)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a")}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}
}

func checkSubs(given [][]byte, expected [][]byte) bool {
	if len(given) != len(expected) {
		return false
	}
	for _, e := range expected {
		found := false
		for _, g := range given {
			if reflect.DeepEqual(e, g) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
