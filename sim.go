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
	updateChannels map[pubkey]chan *update
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

func (mo *mockOverlay) receivedUpdateChannel() chan *update {
	return mo.net.updateChannels[mo.peer]
}

func (mo *mockOverlay) sendEvent(id *pubkey, e *event) {

	// simulate serialization
	d := e.marshal()
	// TODO log traffic
	ec, _ := unmarshalEvent(d)

	mo.net.eventChannels[*id] <- ec
}

func (mo *mockOverlay) sendUpdate(id *pubkey, filters [][]byte) {
	u := newUpdate(&mo.peer, filters)

	// simulate serialization
	d := u.marshal()
	// TODO log traffic
	uc, _ := unmarshalUpdate(d)

	mo.net.updateChannels[*id] <- uc
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
		updateChannels: make(map[pubkey]chan *update),
		eventChannels:  make(map[pubkey]chan *event),
	}

	// create peers and channels
	allPeerIds := make([]*pubkey, size, size)
	for i := 0; i < size; i++ {
		var id pubkey
		rand.Read(id[:])
		allPeerIds[i] = &id
		net.connections[id] = make([]*pubkey, 0)
		net.updateChannels[id] = make(chan *update, 2048) // XXX
		net.eventChannels[id] = make(chan *event, 2084)   // XXX
	}

	// create connections (symmetrical and unique)
	for i, id := range allPeerIds {
		exausted := false
		for len(net.connections[*id]) < peerCnt && !exausted {

			// shitty chord sim
			exp := float64(len(net.connections[*id]))
			start := (i + int(math.Pow(2, exp))) % size

			for j := 0; j < size; j++ { // try peers once then give up
				oid := allPeerIds[(start+j)%size]
				self := *oid == *id // dont link to self
				full := len(net.connections[*oid]) == peerCnt
				if self || full || net.connected(id, oid) {
					if j == size-1 {
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

func randomTopic(limit int) []byte {

	// vaguely based on twitter distribution
	// x := rand.NormFloat64() * rand.NormFloat64() * rand.NormFloat64()
	// return []byte(fmt.Sprintf("%d", int(math.Abs(x*10000.0))))

	return []byte(fmt.Sprintf("%d", rand.Intn(limit)))
}

func updateCnt(x *sync.Mutex, m map[sha256digest]int, d *sha256digest) {
	x.Lock()
	if cnt, ok := m[*d]; ok {
		m[*d] = cnt + 1
	} else {
		m[*d] = 1
	}
	x.Unlock()
}

func calcCoverage(
	subcnt map[sha256digest]int,
	pubCnt map[sha256digest]int,
	delivercnt map[sha256digest]int,
) float64 {
	expected := 0
	reality := 0
	// TODO account for duplicate delivery?
	for t, pcnt := range pubCnt {
		if scnt, ok := subcnt[t]; ok {
			expected += pcnt * scnt
		}
	}
	for _, cnt := range delivercnt {
		reality += cnt
	}
	return float64(reality) / float64(expected)
}

func calcRouting(
	pubCnt map[sha256digest]int,
	recvCnt map[sha256digest]int,
	deliverCnt map[sha256digest]int,
	dropDupCnt map[sha256digest]int,
	dropTtlCnt map[sha256digest]int,
	routeDirectCnt map[sha256digest]int,
	routeWellCnt map[sha256digest]int,
	routeRandomCnt map[sha256digest]int,
) map[string]float64 {

	r := make(map[string]float64)

	total := 0
	for _, cnt := range pubCnt {
		total += cnt
	}
	for _, cnt := range recvCnt {
		total += cnt
	}

	// dropped duplicates
	dropDupTotal := 0
	for _, cnt := range dropDupCnt {
		dropDupTotal += cnt
	}
	r["duplicates"] = float64(dropDupTotal) / float64(total)

	// dropped ttl == 0
	dropTtlTotal := 0
	for _, cnt := range dropTtlCnt {
		dropTtlTotal += cnt
	}
	r["expired"] = float64(dropTtlTotal) / float64(total)

	return r
}

func collectStats(
	l *logger,
	subcnt map[sha256digest]int,
	stop chan bool, // signal stop collection and compile results
	rc chan map[string]float64, // results channel
	logStdOut bool,
) {
	pubMutex := new(sync.Mutex)
	pubCnt := make(map[sha256digest]int)
	recvMutex := new(sync.Mutex)
	recvCnt := make(map[sha256digest]int)
	dropDupMutex := new(sync.Mutex)
	dropDupCnt := make(map[sha256digest]int)
	dropTtlMutex := new(sync.Mutex)
	dropTtlCnt := make(map[sha256digest]int)
	routeWellMutex := new(sync.Mutex)
	routeWellCnt := make(map[sha256digest]int)
	routeDirectMutex := new(sync.Mutex)
	routeDirectCnt := make(map[sha256digest]int)
	routeRandomMutex := new(sync.Mutex)
	routeRandomCnt := make(map[sha256digest]int)
	deliverMutex := new(sync.Mutex)
	deliverCnt := make(map[sha256digest]int)

	stopCollection := false
	for !stopCollection {
		select {
		case <-l.updatesSent:
		case <-l.updatesReceived:
		case <-l.updatesSuccess:
		case <-l.updatesFail:
		case le := <-l.eventsPublished:
			digest := le.entry.TopicDigest
			updateCnt(pubMutex, pubCnt, digest)
		case le := <-l.eventsReceived:
			digest := le.entry.TopicDigest
			updateCnt(recvMutex, recvCnt, digest)
		case le := <-l.eventsDeliver:
			digest := le.entry.TopicDigest
			updateCnt(deliverMutex, deliverCnt, digest)
		case le := <-l.eventsDropDuplicate:
			digest := le.entry.TopicDigest
			le.printState()
			updateCnt(dropDupMutex, dropDupCnt, digest)
		case le := <-l.eventsDropTTL:
			digest := le.entry.TopicDigest
			updateCnt(dropTtlMutex, dropTtlCnt, digest)
		case le := <-l.eventsRouteDirect:
			digest := le.entry.TopicDigest
			updateCnt(routeDirectMutex, routeDirectCnt, digest)
		case le := <-l.eventsRouteWell:
			digest := le.entry.TopicDigest
			updateCnt(routeWellMutex, routeWellCnt, digest)
		case le := <-l.eventsRouteRandom:
			digest := le.entry.TopicDigest
			updateCnt(routeRandomMutex, routeRandomCnt, digest)
		case <-stop:
			stopCollection = true
		}
	}
	r := calcRouting(
		pubCnt, recvCnt, deliverCnt, dropDupCnt, dropTtlCnt,
		routeDirectCnt, routeWellCnt, routeRandomCnt,
	)
	r["coverage"] = calcCoverage(subcnt, pubCnt, deliverCnt)
	if logStdOut {
		fmt.Printf("Coverage: %f\n", r["coverage"])
	}
	rc <- r
}

func addSubscriptions(nodes []*Node, subs int) map[sha256digest]int {
	subcnt := make(map[sha256digest]int) // digest -> sub count
	topicLimit := len(nodes)
	if subs > topicLimit {
		panic("Requested subs > available topics!")
	}
	for _, node := range nodes {
		for i := 0; i < subs; i++ {
			for {
				topicDigest := sha256sum(randomTopic(topicLimit))
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
	t := randomTopic(len(nodes))
	d := make([]byte, 20, 20)
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
func Simulate(
	c *Config,
	size int,
	pubs int,
	subs int,
	logStdOut bool, // FIXME pass logger/io.Writer instead
) map[string]float64 {

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
	stop := make(chan bool, 8) // XXX
	results := make(chan map[string]float64)
	go collectStats(l, subcnt, stop, results, logStdOut)

	// wait for filters to propagate
	delay := c.PropagationDelay * uint64(c.FiltersDepth) * 1
	if logStdOut {
		fmt.Printf("Time for propagation: %ds\n", delay/1000)
	}
	time.Sleep(time.Duration(delay) * time.Millisecond)

	createEvents(nodes, pubs, c, logStdOut)

	// let things settle
	if logStdOut {
		fmt.Printf("\nTime to settle: %ds\n", delay/2000)
	}
	time.Sleep(time.Duration(delay/2) * time.Millisecond)

	// stop nodes
	stop <- true
	for _, node := range nodes {
		node.Stop()
	}

	r := <-results
	return r
}
