package quasar

import (
	"github.com/f483/dejavu"
	"math/rand"
	"sync"
	"time"
)

// FIXME use io.Reader and io.Writer where possible

// Quasar holds the pubsup state
type Quasar struct {
	net               networkOverlay
	subscribers       map[hash160digest][]chan []byte
	topics            map[hash160digest][]byte
	mutex             *sync.RWMutex
	peers             map[pubkey]*peerData
	log               *Logger
	history           dejavu.DejaVu // memory of past events
	cfg               *Config
	stopDispatcher    chan bool
	stopPropagation   chan bool
	stopExpiredPeerGC chan bool
}

// NewQuasar create instance with the sane defaults.
func NewQuasar() *Quasar {
	return newQuasar(nil, nil, &DefaultConfig)
}

// NewQuasarCustom create instance with custom logging/setup for testing.
func NewQuasarCustom(l *Logger, c *Config) *Quasar {
	// FIXME enable passing of nodeId/pubkey
	// FIXME add default network
	return newQuasar(nil, l, c)
}

func newQuasar(n networkOverlay, l *Logger, c *Config) *Quasar {
	// TODO set default config if nil provided
	d := dejavu.NewProbabilistic(c.HistoryLimit, c.HistoryAccuracy)
	return &Quasar{
		net:               n,
		subscribers:       make(map[hash160digest][]chan []byte),
		topics:            make(map[hash160digest][]byte),
		mutex:             new(sync.RWMutex),
		peers:             make(map[pubkey]*peerData),
		log:               l,
		history:           d,
		cfg:               c,
		stopDispatcher:    nil, // set on Start() call
		stopPropagation:   nil, // set on Start() call
		stopExpiredPeerGC: nil, // set on Start() call
	}
}

func (q *Quasar) isConnected(peerId *pubkey) bool {
	for _, connectedPeerId := range q.net.ConnectedPeers() {
		if connectedPeerId == *peerId {
			return true
		}
	}
	return false
}

func (q *Quasar) processUpdate(u *peerUpdate) {
	go q.log.updateReceived(q, u)
	if q.isConnected(u.peer) == false {
		go q.log.updateFail(q, u)
		return // ignore to prevent memory attack
	}

	q.mutex.Lock()
	data, ok := q.peers[*u.peer]

	if !ok { // init if doesnt exist
		depth := q.cfg.FiltersDepth
		data = &peerData{
			filters:    newFilters(q.cfg),
			timestamps: make([]uint64, depth, depth),
		}
		q.peers[*u.peer] = data
	}

	// update peer data
	data.filters[u.index] = u.filter
	data.timestamps[u.index] = makePeerTimestamp()
	q.mutex.Unlock()
	go q.log.updateSuccess(q, u)
}

// Publish a message on the network for given topic.
func (q *Quasar) Publish(topic []byte, message []byte) {
	// TODO validate input
	event := newEvent(topic, message, q.cfg.DefaultEventTTL)
	go q.log.eventPublished(q, event)
	go q.route(event)
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
		filters[0] = filterInsertDigest(filters[0], q.cfg, digest)
	}
	pubkey := q.net.Id()
	filters[0] = filterInsert(filters[0], q.cfg, pubkey[:])
	for _, data := range q.peers {
		if peerDataExpired(data, q.cfg) {
			continue
		}
		for i := 1; uint32(i) < q.cfg.FiltersDepth; i++ {
			filters[i] = mergeFilters(filters[i], data.filters[i-1])
		}
	}
	for _, id := range q.net.ConnectedPeers() {
		for i := 0; uint32(i) < (q.cfg.FiltersDepth - 1); i++ {
			// top filter never sent as not used by peers
			go q.net.SendUpdate(&id, uint32(i), filters[i])
			go q.log.updateSent(q, uint32(i), filters[i], &id)
		}
	}
	q.mutex.RUnlock()
}

// Algorithm 2 from the quasar paper.
func (q *Quasar) route(e *event) {
	q.mutex.RLock()
	id := q.net.Id()
	if q.isDuplicate(e) {
		go q.log.eventDropDuplicate(q, e)
		q.mutex.RUnlock()
		return
	}
	if receivers, ok := q.subscribers[*e.topicDigest]; ok {
		q.log.eventDeliver(q, e)
		q.deliver(receivers, e)
		e.publishers = append(e.publishers, id)
		for _, peerId := range q.net.ConnectedPeers() {
			go q.net.SendEvent(&peerId, e)
			go q.log.eventRouteDirect(q, e, &peerId)
		}
		q.mutex.RUnlock()
		return
	}
	e.ttl -= 1
	if e.ttl == 0 {
		go q.log.eventDropTTL(q, e)
		q.mutex.RUnlock()
		return
	}
	for i := 0; uint32(i) < q.cfg.FiltersDepth; i++ {
		for peerId, data := range q.peers {
			f := data.filters[i]
			if filterContainsDigest(f, q.cfg, *e.topicDigest) {
				negRt := false
				for _, publisher := range e.publishers {
					if filterContains(f, q.cfg, publisher[:]) {
						negRt = true
					}
				}
				if !negRt {
					go q.net.SendEvent(&peerId, e)
					go q.log.eventRouteWell(q, e, &peerId)
					q.mutex.RUnlock()
					return
				}
			}
		}
	}
	peerId := q.randomPeer()
	if peerId != nil {
		go q.net.SendEvent(peerId, e)
		go q.log.eventRouteRandom(q, e, peerId)
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
				go q.log.eventReceived(q, event)
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
	delay := time.Duration(q.cfg.PropagationDelay) * time.Millisecond
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

// Subscriptions retruns a slice of currently subscribed topics.
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
