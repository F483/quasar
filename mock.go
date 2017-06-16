package quasar

import "math/rand"

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

// QuasarMockNetwork creates mock network that uses the quasar protocol
// but with a mock overlay network. Can be used to test
// subscriptions/events/delivery, but not network churn or peer discovery.
func QuasarMockNetwork(l *QuasarLog, cfg config,
	size int, peerCnt int) []*Quasar {

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
			neighbour := net.peers[rand.Intn(len(net.peers))]
			net.connections[*peerId][j] = *neighbour
		}
	}

	// create quasar nodes
	nodes := make([]*Quasar, size, size)
	for i := 0; i < size; i++ {
		n := mockOverlay{peer: *net.peers[i], net: net}
		nodes[i] = newQuasar(&n, l, cfg)
	}

	return nodes
}
