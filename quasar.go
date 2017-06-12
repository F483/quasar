package quasar

import (
	"github.com/f483/dejavu"
	"math/rand"
	"sync"
	"time"
)

// FIXME use io.Reader and io.Writer where possible

type config struct {
	defaultEventTTL  uint32  // decremented every hop
	filterFreshness  uint64  // in seconds TODO change to ms
	propagationDelay uint64  // in seconds TODO change to ms
	historyLimit     uint32  // entries remembered
	historyAccuracy  float64 // chance of error
	filtersDepth     uint32  // filter stack height
	filtersM         uint64  // size in bits (multiple of 64)
	filtersK         uint64  // number of hashes
}

// FIXME use individual constants once optimal values found
var defaultConfig = config{
	defaultEventTTL:  32,              // 4 missed wells
	filterFreshness:  50,              // 50sec (>2.5 propagation)
	propagationDelay: 20,              // 20sec (~0.5M/min const traffic)
	historyLimit:     1048576,         // remember last 1G
	historyAccuracy:  1.0 / 1048576.0, // avg 1 false positive per 1G
	filtersDepth:     8,               // reaches many many nodes
	filtersM:         8192,            // 1K (packet under safe MTU 1400)
	filtersK:         6,               // (m / n) log(2)
}

type Quasar struct {
	net               networkOverlay
	subscribers       map[hash160digest][]chan []byte
	topics            map[hash160digest][]byte
	mutex             *sync.RWMutex
	peers             map[pubkey]*peerData
	history           dejavu.DejaVu // memory of past events
	cfg               config
	stopDispatcher    chan bool
	stopPropagation   chan bool
	stopExpiredPeerGC chan bool
}

// Create new quasar instance
func NewQuasar() *Quasar {
	return newQuasar(nil, defaultConfig) // FIXME add default network
}

func newQuasar(net networkOverlay, c config) *Quasar {
	d := dejavu.NewProbabilistic(c.historyLimit, c.historyAccuracy)
	return &Quasar{
		net:               net,
		subscribers:       make(map[hash160digest][]chan []byte),
		topics:            make(map[hash160digest][]byte),
		mutex:             new(sync.RWMutex),
		peers:             make(map[pubkey]*peerData),
		history:           d,
		cfg:               c,
		stopDispatcher:    nil, // set on Start() call
		stopPropagation:   nil, // set on Start() call
		stopExpiredPeerGC: nil, // set on Start() call
	}
}

func (q *Quasar) processUpdate(u *peerUpdate) {
	q.mutex.Lock()
	data, ok := q.peers[*u.peer]

	if !ok { // init if doesnt exist
		depth := q.cfg.filtersDepth
		data = &peerData{
			filters:    newFilters(q.cfg),
			timestamps: make([]uint64, depth, depth),
		}
		q.peers[*u.peer] = data
	}

	// update peer data
	data.filters[u.index] = u.filter
	data.timestamps[u.index] = uint64(time.Now().Unix())
	q.mutex.Unlock()
}

func (q *Quasar) Publish(topic []byte, message []byte) {
	// TODO validate input
	go q.route(newEvent(topic, message, q.cfg.defaultEventTTL))
}

func (q *Quasar) isDuplicate(e *event) bool {
	return q.history.Witness(append(e.topicDigest[:20], e.message...))
}

func (q *Quasar) deliver(receivers []chan []byte, e *event) {
	for _, receiver := range receivers {
		receiver <- e.message
	}
}

// Algorithm 1 from the quasar paper.
func (q *Quasar) sendUpdates() {
	q.mutex.RLock()
	filters := newFilters(q.cfg)
	for digest := range q.subscribers {
		filters[0] = filterInsertDigest(filters[0], digest, q.cfg)
	}
	pubkey := q.net.Id()
	filters[0] = filterInsert(filters[0], pubkey[:], q.cfg)
	for _, data := range q.peers {
		if peerDataExpired(data, q.cfg) {
			continue
		}
		for i := 1; uint32(i) < q.cfg.filtersDepth; i++ {
			filters[i] = mergeFilters(filters[i], data.filters[i-1])
		}
	}
	for _, peerId := range q.net.ConnectedPeers() {
		for i := 0; uint32(i) < (q.cfg.filtersDepth - 1); i++ {
			// top filter never sent as not used by peers
			go q.net.SendUpdate(&peerId, uint32(i), filters[i])
		}
	}
	q.mutex.RUnlock()
}

// Algorithm 2 from the quasar paper.
func (q *Quasar) route(e *event) {
	q.mutex.RLock()
	if q.isDuplicate(e) {
		q.mutex.RUnlock()
		return
	}
	if receivers, ok := q.subscribers[*e.topicDigest]; ok {
		q.deliver(receivers, e)
		e.publishers = append(e.publishers, q.net.Id())
		for _, peerId := range q.net.ConnectedPeers() {
			go q.net.SendEvent(&peerId, e)
		}
		q.mutex.RUnlock()
		return
	}
	e.ttl -= 1
	if e.ttl == 0 {
		q.mutex.RUnlock()
		return
	}
	for i := 0; uint32(i) < q.cfg.filtersDepth; i++ {
		for peerId, data := range q.peers {
			f := data.filters[i]
			if filterContainsDigest(f, *e.topicDigest, q.cfg) {
				negRt := false
				for _, publisher := range e.publishers {
					if filterContains(f, publisher[:], q.cfg) {
						negRt = true
					}
				}
				if !negRt {
					go q.net.SendEvent(&peerId, e)
					q.mutex.RUnlock()
					return
				}
			}
		}
	}
	peerId := q.randomPeer()
	if peerId != nil {
		go q.net.SendEvent(peerId, e)
	}
	q.mutex.RUnlock()
}

func (q *Quasar) randomPeer() *pubkey {
	peers := q.net.ConnectedPeers()
	if len(peers) == 0 {
		return nil
	}
	peerId := peers[rand.Intn(len(peers))]
	return &peerId
}

func (q *Quasar) dispatchInput() {
	for {
		select {
		case peerUpdate := <-q.net.ReceivedUpdateChannel():
			if validUpdate(peerUpdate, q.cfg) {
				go q.processUpdate(peerUpdate)
			}
		case event := <-q.net.ReceivedEventChannel():
			if validEvent(event) {
				go q.route(event)
			}
		case <-q.stopDispatcher:
			return
		}
	}
}

func (q *Quasar) removeExpiredPeers() {
	q.mutex.Lock()
	toRemove := []*pubkey{}
	for peerId, data := range q.peers {
		if peerDataExpired(data, q.cfg) {
			toRemove = append(toRemove, &peerId)
		}
	}
	for _, peerId := range toRemove {
		delete(q.peers, *peerId)
	}
	q.mutex.Unlock()
}

func (q *Quasar) expiredPeerGC() {
	delay := time.Duration(1) * time.Millisecond
	for {
		select {
		case <-time.After(delay):
			go q.removeExpiredPeers()
		case <-q.stopExpiredPeerGC:
			return
		}
	}
}

func (q *Quasar) propagateFilters() {
	delay := time.Duration(q.cfg.propagationDelay) * time.Second
	for {
		select {
		case <-time.After(delay):
			go q.sendUpdates()
		case <-q.stopPropagation:
			return
		}
	}
}

// Start quasar system
func (q *Quasar) Start() {
	q.net.Start()
	q.stopDispatcher = make(chan bool)
	q.stopPropagation = make(chan bool)
	q.stopExpiredPeerGC = make(chan bool)
	go q.dispatchInput()
	go q.propagateFilters()
	go q.expiredPeerGC()
}

// Stop quasar system
func (q *Quasar) Stop() {
	q.net.Stop()
	q.stopDispatcher <- true
	q.stopPropagation <- true
	q.stopExpiredPeerGC <- true
}

// Subscribe provided message receiver channel to given topic.
func (q *Quasar) Subscribe(topic []byte, receiver chan []byte) {
	// TODO validate input
	digest := hash160(topic)
	q.mutex.Lock()
	receivers, ok := q.subscribers[digest]
	if ok != true { // new subscription
		q.subscribers[digest] = []chan []byte{receiver}
		q.topics[digest] = topic
	} else { // append to existing subscribers
		q.subscribers[digest] = append(receivers, receiver)
	}
	q.mutex.Unlock()
}

// Unsubscribe message receiver channel from topic. If nil receiver
// channel is provided all message receiver channels for given topic
// will be removed.
func (q *Quasar) Unsubscribe(topic []byte, receiver chan []byte) {
	// TODO validate input

	digest := hash160(topic)
	q.mutex.Lock()
	receivers, ok := q.subscribers[digest]

	// remove specific message receiver
	if ok && receiver != nil {
		for i, v := range receivers {
			if v == receiver {
				receivers = append(receivers[:i], receivers[i+1:]...)
				q.subscribers[digest] = receivers
				break
			}
		}
	}

	// remove sub key if no specific message
	// receiver provided or no message receiver remaining
	if ok && (receiver == nil || len(q.subscribers[digest]) == 0) {
		delete(q.subscribers, digest)
		delete(q.topics, digest)
	}
	q.mutex.Unlock()
}

// Subscribers retruns message receivers for given topic.
func (q *Quasar) Subscribers(topic []byte) []chan []byte {
	// TODO validate input
	digest := hash160(topic)
	results := []chan []byte{}
	q.mutex.RLock()
	if receivers, ok := q.subscribers[digest]; ok {
		results = append(results, receivers...)
	}
	q.mutex.RUnlock()
	return results
}

// SubscribedTopics retruns a slice of currently subscribed topics.
func (q *Quasar) Subscriptions() [][]byte {
	q.mutex.RLock()
	topics := make([][]byte, len(q.topics))
	i := 0
	for _, topic := range q.topics {
		topics[i] = topic
		i++
	}
	q.mutex.RUnlock()
	return topics
}
