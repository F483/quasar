package quasar

import (
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"runtime"
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

func newMockNetwork(l *logger, c *Config, size int) []*Node {

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
		net.updateChannels[id] = make(chan *peerUpdate, 2048) // XXX
		net.eventChannels[id] = make(chan *event, 2084)       // XXX
	}

	// create connections (symmetrical and unique)
	for _, id := range allPeerIds {
		exausted := false
		for len(net.connections[*id]) < peerCnt && !exausted {
			start := rand.Intn(size)    // random starting point
			for i := 0; i < size; i++ { // try peers once then give up
				oid := allPeerIds[(start+i)%size]
				self := *oid == *id // dont link to self
				full := len(net.connections[*oid]) == peerCnt
				if self || full || net.connected(id, oid) {
					if i == size-1 {
						exausted = true // no more connections possible
					}
					continue // go to next candidate
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
	l *logger,
	subcnt map[hash160digest]int,
	stop chan bool, // signal stop collection and compile results
	rc chan map[string]float64, // results channel
) {

	publishmutex := new(sync.Mutex)
	publishcnt := make(map[hash160digest]int)
	delivermutex := new(sync.Mutex)
	delivercnt := make(map[hash160digest]int)

	stopCollection := false
	for !stopCollection {
		select {
		case <-l.updatesSent:
		case <-l.updatesReceived:
		case <-l.updatesSuccess:
		case <-l.updatesFail:
		case le := <-l.eventsPublished:
			updateCnt(publishmutex, publishcnt, le.entry.topicDigest)
		case <-l.eventsReceived:
		case le := <-l.eventsDeliver:
			updateCnt(delivermutex, delivercnt, le.entry.topicDigest)
		case <-l.eventsDropDuplicate:
		case <-l.eventsDropTTL:
		case <-l.eventsRouteDirect:
		case <-l.eventsRouteWell:
		case <-l.eventsRouteRandom:
		case <-stop:
			stopCollection = true
		}
	}
	r := make(map[string]float64)
	r["coverage"] = calcCoverage(subcnt, publishcnt, delivercnt)
	rc <- r
}

func addSubscriptions(nodes []*Node, subs int) map[hash160digest]int {
	subcnt := make(map[hash160digest]int) // digest -> sub count
	for _, node := range nodes {
		for i := 0; i < subs; i++ {
			for {
				topicDigest := hash160(randomTopic())
				if node.SubscribedDigest(&topicDigest) {
					continue // avoid duplicate subscriptions
				}
				if cnt, ok := subcnt[topicDigest]; ok {
					subcnt[topicDigest] = cnt + 1
				} else {
					subcnt[topicDigest] = 1
				}
				node.SubscribeDigest(&topicDigest, ioutil.Discard)
				break
			}
		}
	}
	return subcnt
}

func createEvent(nodes []*Node, logStdOut bool) {

	// random topic / data
	t := randomTopic()
	d := make([]byte, 10, 10)
	rand.Read(d)

	// published by random node
	node := nodes[rand.Intn(len(nodes))]
	node.Publish(t, d)
	if logStdOut {
		fmt.Printf(".")
	}
}

func createEvents(nodes []*Node, pubs int, c *Config, logStdOut bool) {
	delay := c.PropagationDelay
	if logStdOut {
		total := delay * uint64(pubs) * 2
		fmt.Printf("Time for events: %ds\n", total/1000)
	}

	for i := 0; i < pubs; i++ {
		createEvent(nodes, logStdOut)

		// let cpu chill and collect garbage
		time.Sleep(time.Duration(delay) * time.Millisecond)
		runtime.GC()
	}
}

// Simulate network behaviour for given configuration, size,
// events published, topics subscribed to by node.
// Returns map with resulting statistics.
func Simulate(c *Config, size int, pubs int, subs int, logStdOut bool) map[string]float64 {

	// TODO use io.Writer or logger instead of logStdOut var
	if logStdOut {
		msg := "Simulate: size=%d, pubs=%d, subs=%d\n"
		fmt.Printf(msg, size, pubs, subs)
		fmt.Println("Config:", *c)
	}

	l := newLogger(1024)
	nodes := newMockNetwork(l, c, size)

	subcnt := addSubscriptions(nodes, subs)
	runtime.GC()

	// start nodes
	for _, node := range nodes {
		node.Start()
	}

	// start stats collector
	stop := make(chan bool)
	results := make(chan map[string]float64)
	go collectStats(l, subcnt, stop, results)

	// wait for filters to propagate
	delay := c.PropagationDelay * uint64(c.FiltersDepth) * 1
	if logStdOut {
		fmt.Printf("Time for propagation: %ds\n", delay/1000)
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)

	createEvents(nodes, pubs, c, logStdOut)

	// let things settle
	if logStdOut {
		fmt.Printf("\nTime to settle: %ds\n", delay/1000)
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)

	// stop nodes
	stop <- true
	for _, node := range nodes {
		node.Stop()
	}

	r := <-results
	if logStdOut {
		fmt.Printf("Coverage: %f\n", r["coverage"])
	}
	return r
}
