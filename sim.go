package quasar

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"sync"
	"time"
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

func newMockNetwork(l *Logger, c *Config, size int) []*Node {

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

func updateCnt(x *sync.Mutex, m map[hash160digest]int, d *hash160digest) {
	x.Lock()
	if cnt, ok := m[*d]; ok {
		m[*d] = cnt + 1
	} else {
		m[*d] = 1
	}
	x.Unlock()
}

func calcCoverage(
	subcnt map[hash160digest]int,
	pubcnt map[hash160digest]int,
	delivercnt map[hash160digest]int,
) float64 {
	expected := 0
	reality := 0
	for t, cnt := range pubcnt {
		expected += subcnt[t] * cnt
	}
	for _, cnt := range delivercnt {
		reality += cnt
	}
	return float64(reality) / float64(expected)
}

func collectStats(
	l *Logger,
	topics map[hash160digest][]byte,
	subcnt map[hash160digest]int,
	stop chan bool, // signal stop collection and compile results
	rc chan map[string]float64, // results channel
) {

	mutex := new(sync.Mutex)
	pubcnt := make(map[hash160digest]int)
	delivercnt := make(map[hash160digest]int)

	stopCollection := false
	for !stopCollection {
		select {
		case <-l.UpdatesSent:
		case <-l.UpdatesReceived:
		case <-l.UpdatesSuccess:
		case <-l.UpdatesFail:
		case le := <-l.EventsPublished:
			updateCnt(mutex, pubcnt, le.entry.topicDigest)
		case <-l.EventsReceived:
		case le := <-l.EventsDeliver:
			updateCnt(mutex, delivercnt, le.entry.topicDigest)
		case <-l.EventsDropDuplicate:
		case <-l.EventsDropTTL:
		case <-l.EventsRouteDirect:
		case <-l.EventsRouteWell:
		case <-l.EventsRouteRandom:
		case <-stop:
			stopCollection = true
		}
	}
	r := make(map[string]float64)
	r["coverage"] = calcCoverage(subcnt, pubcnt, delivercnt)
	fmt.Printf("coverage: %f\n", r["coverage"])
	rc <- r
}

func Simulate(c *Config, size int, pubs int, subs int) map[string]float64 {

	l := NewLogger(1024)
	nodes := newMockNetwork(l, c, size)
	topics := make(map[hash160digest][]byte) // digest -> topic
	subcnt := make(map[hash160digest]int)    // digest -> sub count

	// add subscriptions
	for _, node := range nodes {
		for i := 0; i < subs; i++ {
			t := randomTopic()
			d := hash160(t)
			if cnt, ok := subcnt[d]; ok {
				subcnt[d] = cnt + 1
			} else {
				topics[d] = t
				subcnt[d] = 1
			}
			go node.Subscribe(t, ioutil.Discard)
		}
	}

	// start nodes
	for _, node := range nodes {
		go node.Start()
	}

	// start stats collector
	stop := make(chan bool)
	results := make(chan map[string]float64)
	go collectStats(l, topics, subcnt, stop, results)

	// wait for filters to propagate
	delay := c.PropagationDelay * uint64(c.FiltersDepth) * 3
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// create events
	for _, node := range nodes {
		for i := 0; i < pubs; i++ {
			t := randomTopic()
			d := make([]byte, 20, 20)
			rand.Read(d)
			go node.Publish(t, d)
		}
	}

	// wait for events to propagate
	delay = uint64(c.DefaultEventTTL) * c.PropagationDelay * uint64(size)
	fmt.Printf("waiting for event propagation: %dms\n", delay/10)
	time.Sleep(time.Duration(delay/10) * time.Millisecond)

	// stop nodes
	stop <- true
	for _, node := range nodes {
		go node.Stop()
	}

	return <-results
}
