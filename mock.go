package quasar

import (
	"fmt"
	//"io/ioutil"
	"math"
	"math/rand"
)

type mockNetwork struct {
	peers          []*pubkey
	connections    map[pubkey][]pubkey
	updateChannels map[pubkey]chan *peerUpdate
	eventChannels  map[pubkey]chan *event
}

type mockOverlay struct {
	peer pubkey
	net  *mockNetwork
}

func (mo *mockOverlay) Id() pubkey {
	return mo.peer
}

func (mo *mockOverlay) ConnectedPeers() []pubkey {
	return mo.net.connections[mo.peer]
}

func (mo *mockOverlay) ReceivedEventChannel() chan *event {
	return mo.net.eventChannels[mo.peer]
}

func (mo *mockOverlay) ReceivedUpdateChannel() chan *peerUpdate {
	return mo.net.updateChannels[mo.peer]
}

func (mo *mockOverlay) SendEvent(id *pubkey, e *event) {
	mo.net.eventChannels[*id] <- e
}

func (mo *mockOverlay) SendUpdate(id *pubkey, i uint32, filter []byte) {
	u := &peerUpdate{peer: &mo.peer, index: i, filter: filter}
	mo.net.updateChannels[*id] <- u
}

func (mo *mockOverlay) Start() {

}

func (mo *mockOverlay) Stop() {

}

// NewMockNetwork creates network that uses the quasar protocol  but
// with a mock overlay network. Can be used to test
// subscriptions/events/delivery, but not network churn or peer discovery.
func NewMockNetwork(l *Logger, c *Config, size int) []*Node {

	// TODO add chance of dropped package to args

	peerCnt := 20
	if peerCnt >= size {
		peerCnt = size - 1
	}

	net := &mockNetwork{
		peers:          make([]*pubkey, size, size),
		connections:    make(map[pubkey][]pubkey),
		updateChannels: make(map[pubkey]chan *peerUpdate),
		eventChannels:  make(map[pubkey]chan *event),
	}

	// create peers and channels
	for i := 0; i < size; i++ {
		var peerId pubkey
		rand.Read(peerId[:])
		net.peers[i] = &peerId
		net.updateChannels[peerId] = make(chan *peerUpdate)
		net.eventChannels[peerId] = make(chan *event)
	}

	// create connections
	for i := 0; i < size; i++ {
		peerId := net.peers[i]
		net.connections[*peerId] = make([]pubkey, peerCnt, peerCnt)
		for j := 0; j < peerCnt; j++ {
			// FIXME do not set self as peer
			neighbour := net.peers[randIntnExcluding(len(net.peers), i)]
			net.connections[*peerId][j] = *neighbour
		}
	}

	// create quasar nodes
	nodes := make([]*Node, size, size)
	for i := 0; i < size; i++ {
		n := mockOverlay{peer: *net.peers[i], net: net}
		nodes[i] = newNode(&n, l, c)
	}

	return nodes
}

func randomTopic() []byte {
	// vaguely based on twitter distribution
	x := rand.NormFloat64() * rand.NormFloat64() * rand.NormFloat64()
	return []byte(fmt.Sprintf("%d", int(math.Abs(x*10000.0))))
}

/*
func RunSimulation(l *Logger, c *Config, size int, subsPerNode int) (map[hash160digest][]byte, map[hash160digest]int) {

	topics := make(map[hash160digest][]byte) // digest -> topic
	subcnt := make(map[hash160digest]int)    // digest -> sub count
	// receivers := make(map[hash160digest]chan [][]byte)
	nodes := NewMockNetwork(l, c, size)

	// add subscriptions
	for node := range nodes {
		for i := 0; i < subsPerNode; i++ {

			// update topics/subcnt
			t := randomTopic()
			d := hash160(t)
			if cnt, ok := subcnt[d]; ok {
				subcnt[d] = cnt + 1
			} else {
				topics[d] = t
				subcnt[d] = 1
			}

			node.Subscribe(t, nil) // FIXME
			// node.Subscribe(t, ioutil.Disgard)
		}
	}

	return topics, subcnt
}
*/
