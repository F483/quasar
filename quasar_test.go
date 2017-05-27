package quasar

import (
	"encoding/hex"
	"reflect"
	"testing"
)

type MockNetwork struct {
	peers          []PeerID
	connections    map[PeerID][]PeerID
	updateChannels map[PeerID]chan Update
	eventChannels  map[PeerID]chan Event
}

type MockOverlayNetwork struct {
	id      PeerID
	network MockNetwork
}

func (mqt *MockOverlayNetwork) Id() PeerID {
	return mqt.id
}

func (mqt *MockOverlayNetwork) ConnectedPeers() []PeerID {
	return mqt.network.connections[mqt.id]
}

func (mqt *MockOverlayNetwork) ReceivedEventChannel() chan Event {
	return mqt.network.eventChannels[mqt.id]

}

func (mqt *MockOverlayNetwork) ReceivedUpdateChannel() chan Update {
	return mqt.network.updateChannels[mqt.id]
}

func (mqt *MockOverlayNetwork) SendEvent(id PeerID, e Event) {
	mqt.network.eventChannels[id] <- e
}

func (mqt *MockOverlayNetwork) SendUpdate(id PeerID, f Filters) {
	update := Update{peerId: &mqt.id, filters: &f}
	mqt.network.updateChannels[id] <- update
}

func (mqt *MockOverlayNetwork) Start() {}

func (mqt *MockOverlayNetwork) Stop() {}

func TestNewEvent(t *testing.T) {
	event := NewEvent("test topic", "test message")
	if event == nil {
		t.Errorf("Expected event!")
	}
	if event.message != "test message" {
		t.Errorf("Event message not set!")
	}
	if event.ttl != TTL {
		t.Errorf("Event ttl not set!")
	}
	if event.publishers == nil && len(event.publishers) != 0 {
		t.Errorf("Event publishers not set!")
	}
	if event.topicDigest == nil {
		t.Errorf("Event digest not set!")
	}
}

func TestHashTopic(t *testing.T) {

	// decode topic
	topicHex := []byte("f483")
	topicBytes := make([]byte, hex.DecodedLen(len(topicHex)))
	topicBytesLen, err := hex.Decode(topicBytes, topicHex)
	if err != nil {
		t.Fatal(err)
	}
	topic := string(topicBytes[:topicBytesLen])

	// decode expected digest
	expectedHex := []byte("4e0123796bee558240c5945ac9aff553fcc6256d")
	expected := Hash160Digest{}
	expectedBytesLen, err := hex.Decode(expected[:], expectedHex)
	if err != nil {
		t.Fatal(err)
	}
	if expectedBytesLen != 20 {
		t.Errorf("Incorrect digest size! %i", expectedBytesLen)
	}

	digest := Hash160(topic)
	if *digest != expected {
		t.Errorf("Hash160 failed!")
	}
}

func TestSubscriptions(t *testing.T) {
	q := NewQuasar(nil)

	a := make(chan string)
	q.Subscribe("a", a)
	subs := q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a"}) {
		t.Errorf("Incorrect subscriptions! ", subs)
	}

	b1 := make(chan string)
	q.Subscribe("b", b1)
	subs = q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a", "b"}) {
		t.Errorf("Incorrect subscriptions! ", subs)
	}

	b2 := make(chan string)
	q.Subscribe("b", b2)
	subs = q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a", "b"}) {
		t.Errorf("Incorrect subscriptions! ", subs)
	}

	c1 := make(chan string)
	q.Subscribe("c", c1)
	subs = q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a", "b", "c"}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}

	c2 := make(chan string)
	q.Subscribe("c", c2)
	subs = q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a", "b", "c"}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}

	// test clears all if no receiver provided
	q.Unsubscribe("c", nil)
	subs = q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a", "b"}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}

	// only clears specific receiver
	q.Unsubscribe("b", b1)
	subs = q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a", "b"}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}

	// clears key when last receiver removed
	q.Unsubscribe("b", b2)
	subs = q.Subscriptions()
	if !reflect.DeepEqual(subs, []string{"a"}) {
		t.Errorf("Incorrect subscriptions!", subs)
	}
}
