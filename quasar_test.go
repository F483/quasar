package quasar

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

// FIXME only mock until network layer implemented
type MockNetwork struct {
	peers          []*pubkey
	connections    map[pubkey][]pubkey
	updateChannels map[pubkey]chan *peerUpdate
	eventChannels  map[pubkey]chan *event
}

type MockOverlay struct {
	peer pubkey
	net  *MockNetwork
}

func (mqt *MockOverlay) Id() pubkey {
	return mqt.peer
}

func (mqt *MockOverlay) ConnectedPeers() []pubkey {
	return mqt.net.connections[mqt.peer]
}

func (mqt *MockOverlay) ReceivedEventChannel() chan *event {
	return mqt.net.eventChannels[mqt.peer]
}

func (mqt *MockOverlay) ReceivedUpdateChannel() chan *peerUpdate {
	return mqt.net.updateChannels[mqt.peer]
}

func (mqt *MockOverlay) SendEvent(id *pubkey, e *event) {
	mqt.net.eventChannels[*id] <- e
}

func (mqt *MockOverlay) SendUpdate(id *pubkey, i uint32, filter []byte) {
	u := &peerUpdate{peer: &mqt.peer, index: i, filter: filter}
	mqt.net.updateChannels[*id] <- u
}

func (mqt *MockOverlay) Start() {

}

func (mqt *MockOverlay) Stop() {

}

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
	q := newQuasar(nil, config{
		defaultEventTTL:  32,
		filterFreshness:  32,
		propagationDelay: 12,
		historyLimit:     256,
		historyAccuracy:  0.000001,
		filtersDepth:     8,
		filtersM:         8192, // m 1k
		filtersK:         6,    // hashes
	})

	a := make(chan []byte)
	q.Subscribe([]byte("a"), a)
	subs := q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	b1 := make(chan []byte)
	q.Subscribe([]byte("b"), b1)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	b2 := make(chan []byte)
	q.Subscribe([]byte("b"), b2)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	c1 := make(chan []byte)
	q.Subscribe([]byte("c"), c1)
	subs = q.Subscriptions()
	expectedSubs := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	if !checkSubs(subs, expectedSubs) {
		t.Errorf("Incorrect subscriptions!")
	}

	c2 := make(chan []byte)
	q.Subscribe([]byte("c"), c2)
	subs = q.Subscriptions()
	expectedSubs = [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	if !checkSubs(subs, expectedSubs) {
		t.Errorf("Incorrect subscriptions!")
	}

	// test clears all if no receiver provided
	q.Unsubscribe([]byte("c"), nil)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	// only clears specific receiver
	if len(q.Subscribers([]byte("b"))) != 2 {
		t.Errorf("Incorrect subscribers!")
	}
	q.Unsubscribe([]byte("b"), b1)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}
	if len(q.Subscribers([]byte("b"))) != 1 {
		t.Errorf("Incorrect subscribers!")
	}

	// clears key when last receiver removed
	q.Unsubscribe([]byte("b"), b2)
	subs = q.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a")}) {
		t.Errorf("Incorrect subscriptions!")
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

func setupMockNetwork(cfg config, netSize int, connCnt int) []*Quasar {

	net := &MockNetwork{
		peers:          make([]*pubkey, netSize, netSize),
		connections:    make(map[pubkey][]pubkey),
		updateChannels: make(map[pubkey]chan *peerUpdate),
		eventChannels:  make(map[pubkey]chan *event),
	}

	// create peers and channels
	for i := 0; i < netSize; i++ {
		var peerId pubkey
		rand.Read(peerId[:])
		net.peers[i] = &peerId
		net.updateChannels[peerId] = make(chan *peerUpdate)
		net.eventChannels[peerId] = make(chan *event)
	}

	// create connections
	for i := 0; i < netSize; i++ {
		peerId := net.peers[i]
		net.connections[*peerId] = make([]pubkey, connCnt, connCnt)
		for j := 0; j < connCnt; j++ {
			neighbour := net.peers[(i+j)%connCnt]
			net.connections[*peerId][j] = *neighbour
		}
	}

	// create quasar nodes
	nodes := make([]*Quasar, netSize, netSize)
	for i := 0; i < netSize; i++ {
		no := MockOverlay{peer: *net.peers[i], net: net}
		nodes[i] = newQuasar(&no, cfg)
	}

	return nodes
}

func TestEventDelivery(t *testing.T) {
	cfg := config{
		defaultEventTTL:  32,
		filterFreshness:  32,
		propagationDelay: 12,
		historyLimit:     256,
		historyAccuracy:  0.000001,
		filtersDepth:     8,
		filtersM:         8192, // m 1k
		filtersK:         6,    // hashes
	}

	nodes := setupMockNetwork(cfg, 20, 20)

	// set subscriptions
	fooReceiver := make(chan []byte)
	nodes[0].Subscribe([]byte("foo"), fooReceiver)

	// start nodes and wait for filters to propagate
	for _, node := range nodes {
		node.Start()
	}
	time.Sleep(time.Millisecond * time.Duration(cfg.propagationDelay*3))

	// create event
	nodes[len(nodes)-1].Publish([]byte("foo"), []byte("foodata"))

	timeout := time.Duration(cfg.propagationDelay) * time.Millisecond
	select {
	case <-time.After(timeout):
		t.Errorf("Timeout event not received!")
	case data := <-fooReceiver:
		if !reflect.DeepEqual(data, []byte("foodata")) {
			t.Errorf("Incorrect event data!")
		}
	}

	// stop nodes
	for _, node := range nodes {
		node.Stop()
	}
}

func TestEventTimeout(t *testing.T) {
	// get coverage for dropping ttl = 0 events
	// may be dropped by history with few nodes

	cfg := config{
		defaultEventTTL:  2,
		filterFreshness:  32,
		propagationDelay: 12,
		historyLimit:     256,
		historyAccuracy:  0.000001,
		filtersDepth:     8,
		filtersM:         8192, // m 1k
		filtersK:         6,    // hashes
	}

	nodes := setupMockNetwork(cfg, 20, 20)

	// start nodes and wait for filters to propagate
	for _, node := range nodes {
		node.Start()
	}
	time.Sleep(time.Millisecond * time.Duration(cfg.propagationDelay*3))

	// create event
	nodes[1].Publish([]byte("bar"), []byte("bardata"))
	time.Sleep(time.Duration(cfg.propagationDelay) * time.Millisecond)

	// stop nodes
	for _, node := range nodes {
		node.Stop()
	}
}

func TestExpiredPeerData(t *testing.T) {
	cfg := config{
		defaultEventTTL:  2,
		filterFreshness:  32,
		propagationDelay: 12,
		historyLimit:     256,
		historyAccuracy:  0.000001,
		filtersDepth:     8,
		filtersM:         8192, // m 1k
		filtersK:         6,    // hashes
	}

	nodes := setupMockNetwork(cfg, 2, 2)

	// start nodes and wait for filters to propagate
	nodes[0].Start()
	nodes[1].Start()

	// let filters propagate
	time.Sleep(time.Millisecond * time.Duration(cfg.propagationDelay*2))

	nodes[0].Stop() //

	// let filters expire
	time.Sleep(time.Millisecond * time.Duration(cfg.filterFreshness*2))

	nodes[1].Stop()
}

func TestNoPeers(t *testing.T) {
	cfg := config{
		defaultEventTTL:  2,
		filterFreshness:  32,
		propagationDelay: 12,
		historyLimit:     256,
		historyAccuracy:  0.000001,
		filtersDepth:     8,
		filtersM:         8192, // m 1k
		filtersK:         6,    // hashes
	}

	nodes := setupMockNetwork(cfg, 1, 0)
	nodes[0].Start()
	nodes[0].Publish([]byte("bar"), []byte("bardata"))
	nodes[0].Stop()
}
