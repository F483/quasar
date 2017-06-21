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
	connections    map[pubkey][]*pubkey
	updateChannels map[pubkey]chan *peerUpdate
	eventChannels  map[pubkey]chan *event
}

func (mn mockNetwork) connected(a *pubkey, b *pubkey) bool {
	for _, x := range mn.connections[*a] {
		if *x == *b {
			return true
		}
	}
	return false
}

type mockOverlay struct {
	peer pubkey
	net  *mockNetwork
}

func (mo *mockOverlay) id() pubkey {
	return mo.peer
}

func (mo *mockOverlay) connectedPeers() []*pubkey {
	return mo.net.connections[mo.peer]
}

func (mo *mockOverlay) isConnected(peerId *pubkey) bool {
	return mo.net.connected(&mo.peer, peerId)
}

func (mo *mockOverlay) receivedEventChannel() chan *event {
	return mo.net.eventChannels[mo.peer]
}

func (mo *mockOverlay) receivedUpdateChannel() chan *peerUpdate {
	return mo.net.updateChannels[mo.peer]
}

func (mo *mockOverlay) sendEvent(id *pubkey, e *event) {
	mo.net.eventChannels[*id] <- e
}

func (mo *mockOverlay) sendUpdate(id *pubkey, i uint32, filter []byte) {
	u := &peerUpdate{peer: &mo.peer, index: i, filter: filter}
	mo.net.updateChannels[*id] <- u
}

func (mo *mockOverlay) start() {

}

func (mo *mockOverlay) stop() {

}

func newMockNetwork(l *Logger, c *Config, size int) []*Node {

	// TODO add chance of dropped package to args

	// size must be even
	peerCnt := 20
	if peerCnt >= size {
		peerCnt = size - 1
	}

	net := &mockNetwork{
		connections:    make(map[pubkey][]*pubkey),
		updateChannels: make(map[pubkey]chan *peerUpdate),
		eventChannels:  make(map[pubkey]chan *event),
	}

	// create peers and channels
	allPeerIds := make([]*pubkey, size, size)
	for i := 0; i < size; i++ {
		var id pubkey
		rand.Read(id[:])
		allPeerIds[i] = &id
		net.connections[id] = make([]*pubkey, 0)
		net.updateChannels[id] = make(chan *peerUpdate)
		net.eventChannels[id] = make(chan *event)
	}

	// create connections (symmetrical and unique)
	for i := 0; i < size; i++ {
		id := allPeerIds[i]
		for len(net.connections[*id]) < peerCnt {
			j := rand.Intn(size) // random starting point
			for {
				oid := allPeerIds[j]
				self := *oid == *id // dont link to self
				full := len(net.connections[*oid]) == peerCnt
				if self || full || net.connected(id, oid) {
					j = (j + 1) % size // walk until candidate found
					continue
				}
				net.connections[*id] = append(net.connections[*id], oid)
				net.connections[*oid] = append(net.connections[*oid], id)
				break
			}
		}
	}

	// create quasar nodes
	nodes := make([]*Node, size, size)
	for i, id := range allPeerIds {
		n := mockOverlay{peer: *id, net: net}
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
	// TODO account for duplicate delivery?
	for t, pcnt := range pubcnt {
		if scnt, ok := subcnt[t]; ok {
			expected += pcnt * scnt
		}
	}
	for _, cnt := range delivercnt {
		reality += cnt
	}
	if expected == 0 {
		return 1.0
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
			// FIXME what about duplicate subs per node?
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
	delay := time.Duration(c.PropagationDelay * uint64(c.FiltersDepth))
	fmt.Printf("Waiting for filter propagation: %dms\n", delay)
	time.Sleep(delay * time.Millisecond)

	delay = time.Duration(delay / 140)
	pn := time.Duration(size * pubs)
	fmt.Printf("Publish delay: %d * %dms => %dms\n", pn, delay, pn*delay)

	// create events
	for _, node := range nodes {
		for i := 0; i < pubs; i++ {
			t := randomTopic()
			d := make([]byte, 10, 10)
			rand.Read(d)
			go node.Publish(t, d)

			// let cpu chill
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
	}

	// stop nodes
	stop <- true
	for _, node := range nodes {
		go node.Stop()
	}

	// let things stop
	// time.Sleep(time.Duration(1000) * time.Millisecond)

	return <-results
}
