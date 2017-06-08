package quasar

import (
	"github.com/f483/dejavu"
	"math/rand"
	"sync"
	"time"
)

// FIXME use io.Reader and io.Writer where possible

type config struct {
	defaultEventTTL     uint32        // decremented every hop
	dispatcherDelay     time.Duration // in ms FIXME uint64
	filterFreshness     uint64        // in seconds
	propagationInterval uint64        // in seconds
	historyLimit        uint32        // entries remembered
	historyAccuracy     float64       // chance of error
	filtersDepth        uint32        // filter stack height
	filtersM            uint64        // size in bits (multiple of 64)
	filtersK            uint64        // number of hashes
}

type peer struct {
	pubkey     *pubkey
	filters    [][]byte
	timestamps []uint64 // unixtime
}

func (p *peer) isExpired(c *config) bool {
	now := uint64(time.Now().Unix())
	for _, timestamp := range p.timestamps {
		if timestamp >= (now - c.filterFreshness) {
			return false
		}
	}
	return true
}

type Quasar struct {
	net             overlayNetwork
	subscribers     map[hash160digest][]chan []byte
	topics          map[hash160digest][]byte
	mutex           *sync.RWMutex
	peers           []peer
	history         dejavu.DejaVu // memory of past events
	cfg             *config
	filters         [][]byte // own (subs + peers)
	stopDispatcher  chan bool
	stopPropagation chan bool
}

// Create new quasar instance
func NewQuasar() *Quasar {
	c := config{
		defaultEventTTL:     1024,
		dispatcherDelay:     1,        // responsive as possible
		filterFreshness:     50,       // 50sec (>2.5 propagation miss)
		propagationInterval: 20,       // 20sec (~0.5m/min const traffic)
		historyLimit:        65536,    // FIXME increase
		historyAccuracy:     0.000001, // FIXME increase
		filtersDepth:        8,        // reaches many many nodes
		filtersM:            8192,     // 1k
		filtersK:            6,        // (m / n) log(2)
	}
	return newQuasar(nil, c)
}

func newQuasar(net overlayNetwork, c config) *Quasar {
	d := dejavu.NewProbabilistic(c.historyLimit, c.historyAccuracy)
	return &Quasar{
		net:             net,
		subscribers:     make(map[hash160digest][]chan []byte),
		topics:          make(map[hash160digest][]byte),
		mutex:           new(sync.RWMutex),
		peers:           make([]peer, 0),
		history:         d,
		cfg:             &c,
		filters:         nil,
		stopDispatcher:  nil, // set on Start() call
		stopPropagation: nil, // set on Start() call
	}
}

func (q *Quasar) processUpdate(u *update) {
	// TODO implement
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
	q.mutex.Lock()
	filters := newFilters(q.cfg)
	for digest := range q.subscribers {
		filters[0] = filterInsertDigest(filters[0], digest, q.cfg)
	}
	pubkey := q.net.Id()
	filters[0] = filterInsert(filters[0], pubkey[:], q.cfg)
	for _, p := range q.peers {
		if p.isExpired(q.cfg) {
			continue
		}
		for i := 1; uint32(i) < q.cfg.filtersDepth; i++ {
			filters[i] = mergeFilters(filters[i], p.filters[i-1])
		}
	}
	q.filters = filters
	for _, p := range q.peers {
		for i := 0; uint32(i) < (q.cfg.filtersDepth - 1); i++ {
			// top filter never sent as not used by peers
			go q.net.SendUpdate(p.pubkey, uint(i), filters[i])
		}
	}
	q.mutex.Unlock()
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
		for _, p := range q.peers {
			q.net.SendEvent(p.pubkey, e)
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
		for _, p := range q.peers {
			f := p.filters[i]
			if filterContainsDigest(f, *e.topicDigest, q.cfg) {
				negRt := false
				for _, publisher := range e.publishers {
					if filterContains(f, publisher[:], q.cfg) {
						negRt = true
					}
				}
				if !negRt {
					q.net.SendEvent(p.pubkey, e)
					q.mutex.RUnlock()
					return
				}
			}
		}
	}
	if len(q.peers) > 0 {
		p := q.peers[rand.Intn(len(q.peers))]
		q.net.SendEvent(p.pubkey, e)
	}
	q.mutex.RUnlock()
}

func (q *Quasar) dispatchInput() {
	for {
		select {
		case update := <-q.net.ReceivedUpdateChannel():
			q.mutex.RLock()
			valid := validUpdate(update, q.cfg)
			q.mutex.RUnlock()
			if valid {
				go q.processUpdate(update)
			}
		case event := <-q.net.ReceivedEventChannel():
			if validEvent(event) {
				go q.route(event)
			}
		case <-q.stopDispatcher:
			return // TODO confirm stopped
		}
		time.Sleep(q.cfg.dispatcherDelay * time.Millisecond)
	}
}

func (q *Quasar) propagateFilters() {
	// TODO implement
}

// Start quasar system
func (q *Quasar) Start() {
	q.net.Start()
	q.stopDispatcher = make(chan bool)
	q.stopPropagation = make(chan bool)
	go q.dispatchInput()
	go q.propagateFilters()
}

// Stop quasar system
func (q *Quasar) Stop() {
	q.net.Stop()
	q.stopDispatcher <- true
	q.stopPropagation <- true
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
