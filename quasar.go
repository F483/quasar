package quasar

import (
	"github.com/f483/dejavu"
	"math/rand"
	"sync"
	"time"
)

type overlayNetwork interface {
	Id() pubkey
	ConnectedPeers() []pubkey
	ReceivedEventChannel() chan *event
	ReceivedUpdateChannel() chan *update
	SendEvent(*pubkey, *event)
	SendUpdate(receiver *pubkey, index uint, filter []byte)
	Start()
	Stop()
}

type event struct {
	topicDigest *hash160digest
	message     []byte
	publishers  []pubkey
	ttl         uint32
}

type update struct {
	peer   *pubkey
	index  uint
	filter []byte
}

type peer struct {
	pubkey    *pubkey
	filters   [][]byte
	timestamp []int64 // unixtime
}

type Config struct {
	DefaultEventTTL     uint32        // decremented every hop
	DispatcherDelay     time.Duration // in ms
	PeerFiltersExpire   uint32        // in seconds
	PropagationInterval uint32        // in seconds
	HistoryLimit        uint32        // entries remembered
	HistoryAccuracy     float64       // chance of error
	FiltersDepth        uint32        // filter stack height
	FiltersM            uint32        // filter size in bits
	FiltersK            uint32        // number of hashes
}

type Quasar struct {
	network         overlayNetwork
	subscribers     map[hash160digest][]chan []byte
	topics          map[hash160digest][]byte
	mutex           *sync.RWMutex
	peers           []peer
	peersMutex      *sync.Mutex
	history         dejavu.DejaVu // memory of past events
	config          Config
	filters         [][]byte // own (subs + peers)
	stopDispatcher  chan bool
	stopPropagation chan bool
}

func newEvent(topic []byte, message []byte, ttl uint32) *event {
	digest := hash160(topic)
	return &event{
		topicDigest: &digest,
		message:     message,
		publishers:  []pubkey{},
		ttl:         ttl,
	}
}

func NewQuasar(network overlayNetwork, c Config) *Quasar {
	d := dejavu.NewProbabilistic(c.HistoryLimit, c.HistoryAccuracy)
	return &Quasar{
		network:         network,
		subscribers:     make(map[hash160digest][]chan []byte),
		topics:          make(map[hash160digest][]byte),
		mutex:           new(sync.RWMutex),
		peers:           make([]peer, 0),
		peersMutex:      new(sync.Mutex),
		history:         d,
		config:          c,
		filters:         nil,
		stopDispatcher:  nil, // set on Start() call
		stopPropagation: nil, // set on Start() call
	}
}

func (q *Quasar) processUpdate(u *update) {
	// TODO implement
}

func (q *Quasar) Publish(topic []byte, message []byte) {
	q.route(newEvent(topic, message, q.config.DefaultEventTTL))
}

func (q *Quasar) isDuplicate(e *event) bool {
	return q.history.Witness(append(e.topicDigest[:20], e.message...))
}

func (q *Quasar) deliver(receivers []chan []byte, e *event) {
	for _, receiver := range receivers {
		receiver <- e.message
	}
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
		e.publishers = append(e.publishers, q.network.Id())
		for _, p := range q.peers {
			q.network.SendEvent(p.pubkey, e)
		}
		q.mutex.RUnlock()
		return
	}
	e.ttl -= 1
	if e.ttl == 0 {
		q.mutex.RUnlock()
		return
	}
	for i := 0; uint32(i) < q.config.FiltersDepth; i++ {
		for _, p := range q.peers {
			f := newFilter(p.filters[i], &q.config)
			if f.containsDigest(*e.topicDigest) {
				negRt := false
				for _, publisher := range e.publishers {
					if f.contains(publisher[:]) {
						negRt = true
					}
				}
				if !negRt {
					q.network.SendEvent(p.pubkey, e)
					q.mutex.RUnlock()
					return
				}
			}
		}
	}
	if len(q.peers) > 0 {
		p := q.peers[rand.Intn(len(q.peers))]
		q.network.SendEvent(p.pubkey, e)
	}
	q.mutex.RUnlock()
}

func (q *Quasar) dispatchInput() {
	for {
		select {
		case update := <-q.network.ReceivedUpdateChannel():
			go q.processUpdate(update)
		case e := <-q.network.ReceivedEventChannel():
			go q.route(e)
		case <-q.stopDispatcher:
			return // TODO confirm stopped
		}
		time.Sleep(q.config.DispatcherDelay * time.Millisecond)
	}
}

func (q *Quasar) propagateFilters() {
	// TODO implement
}

// Start quasar system
func (q *Quasar) Start() {
	q.network.Start()
	q.stopDispatcher = make(chan bool)
	q.stopPropagation = make(chan bool)
	go q.dispatchInput()
	go q.propagateFilters()
}

// Stop quasar system
func (q *Quasar) Stop() {
	q.network.Stop()
	q.stopDispatcher <- true
	q.stopPropagation <- true
}

// Subscribe provided message receiver channel to given topic.
func (q *Quasar) Subscribe(topic []byte, receiver chan []byte) {
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
