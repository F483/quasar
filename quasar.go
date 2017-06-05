package quasar

import (
	"github.com/f483/dejavu"
	"sort"
	"sync"
	"time"
)

// FIXME messages and topics should be byte slices

type overlayNetwork interface {
	Id() pubkey
	ConnectedPeers() []pubkey
	ReceivedEventChannel() chan *event
	ReceivedUpdateChannel() chan *update
	SendEvent(pubkey, event)
	SendUpdate(receiver pubkey, index uint, filter []byte)
	Start()
	Stop()
}

type event struct {
	topicDigest *hash160digest
	message     string
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
	DefaultEventTTL        uint32        // decremented every hop
	InputDispatcherDelay   time.Duration // in ms
	PeerFiltersExpire      uint32        // in seconds
	PublishFiltersInterval uint32        // in seconds
	HistoryLimit           uint32        // entries remembered
	HistoryAccuracy        float64       // chance of error
	FiltersDepth           uint32        // filter stack height
	FiltersM               uint32        // filter size in bits
	FiltersK               uint32        // number of hashes
}

type Quasar struct {
	network               overlayNetwork
	subs                  map[string][]chan string // topic -> subscribers
	subsMutex             *sync.RWMutex
	peers                 []peer
	peersMutex            *sync.Mutex
	history               dejavu.DejaVu // memory of past events
	config                Config
	filters               [][]byte                 // own (subs + peers)
	cachedSubDigests      map[hash160digest]string // digest -> topic
	cachedSubDigestsMutex *sync.RWMutex
	stopInputDispatcher   chan bool
	stopFilterPropagation chan bool
}

func newEvent(topic string, message string, ttl uint32) *event {
	return &event{
		topicDigest: hash160(topic),
		message:     message,
		publishers:  []pubkey{},
		ttl:         ttl,
	}
}

func NewQuasar(network overlayNetwork, c Config) *Quasar {

	d := dejavu.NewProbabilistic(c.HistoryLimit, c.HistoryAccuracy)
	return &Quasar{
		network:               network,
		subs:                  make(map[string][]chan string),
		subsMutex:             new(sync.RWMutex),
		peers:                 make([]peer, 0),
		peersMutex:            new(sync.Mutex),
		history:               d,
		config:                c,
		filters:               nil,
		cachedSubDigests:      make(map[hash160digest]string),
		cachedSubDigestsMutex: new(sync.RWMutex),
		stopInputDispatcher:   nil, // set on Start() call
		stopFilterPropagation: nil, // set on Start() call
	}
}

func (q *Quasar) rebuildCaches() {

	// rebuild sub digest mapping
	q.subsMutex.RLock()
	q.cachedSubDigestsMutex.Lock()

	q.cachedSubDigests = make(map[hash160digest]string) // clear
	for topic := range q.subs {
		q.cachedSubDigests[*hash160(topic)] = topic
	}

	q.cachedSubDigestsMutex.Unlock()
	q.subsMutex.RUnlock()

	// TODO rebuild q.filters
}

func (q *Quasar) processUpdate(u *update) {
	// TODO implement
}

func (q *Quasar) deliver(e *event) {
	q.cachedSubDigestsMutex.RLock()
	if topic, ok := q.cachedSubDigests[*e.topicDigest]; ok {
		q.subsMutex.RLock()
		for _, receiver := range q.subs[topic] {
			receiver <- e.message
		}
		q.subsMutex.RUnlock()
	}
	q.cachedSubDigestsMutex.RUnlock()
}

func (q *Quasar) Publish(topic string, message string) {
	q.processEvent(newEvent(topic, message, q.config.DefaultEventTTL))
}

func (q *Quasar) processEvent(e *event) {
	eventData := string(e.topicDigest[:20]) + e.message
	if q.history.Witness([]byte(eventData)) {
		return // dejavu, already seen
	}

	q.deliver(e)

	// TODO propagate
}

func (q *Quasar) dispatchInput() {
	for {
		select {
		case update := <-q.network.ReceivedUpdateChannel():
			go q.processUpdate(update)
		case e := <-q.network.ReceivedEventChannel():
			go q.processEvent(e)
		case <-q.stopInputDispatcher:
			return // TODO confirm stopped
		}
		time.Sleep(q.config.InputDispatcherDelay * time.Millisecond)
	}
}

func (q *Quasar) propagateFilters() {
	// TODO implement
}

// Start quasar system
func (q *Quasar) Start() {
	q.network.Start()
	q.stopInputDispatcher = make(chan bool)
	q.stopFilterPropagation = make(chan bool)
	go q.dispatchInput()
	go q.propagateFilters()
}

// Stop quasar system
func (q *Quasar) Stop() {
	q.network.Stop()
	q.stopInputDispatcher <- true
	q.stopFilterPropagation <- true
}

// Subscribe provided message receiver channel to given topic.
func (q *Quasar) Subscribe(topic string, msgReceiver chan string) {
	q.subsMutex.Lock()
	receivers, ok := q.subs[topic]
	if ok != true {
		q.subs[topic] = []chan string{msgReceiver}
	} else {
		q.subs[topic] = append(receivers, msgReceiver)
	}
	q.subsMutex.Unlock()
	go q.rebuildCaches()
}

// Unsubscribe message receiver channel from topic. If nil receiver channel is
// provided all message receiver channels for given topic will be removed.
func (q *Quasar) Unsubscribe(topic string, msgReceiver chan string) {
	q.subsMutex.Lock()
	receivers, ok := q.subs[topic]

	// remove specific message receiver
	if ok && msgReceiver != nil {
		for i, v := range receivers {
			if v == msgReceiver {
				receivers = append(receivers[:i], receivers[i+1:]...)
				q.subs[topic] = receivers
				break
			}
		}
	}

	// remove sub key if no specific message
	// receiver provided or no message receiver remaining
	if ok && (msgReceiver == nil || len(q.subs[topic]) == 0) {
		delete(q.subs, topic)
	}
	q.subsMutex.Unlock()
	go q.rebuildCaches()
}

// Subscriptions retruns copy of the current subscriptions map
func (q *Quasar) Subscriptions() map[string][]chan string {
	q.subsMutex.RLock()
	resultMap := make(map[string][]chan string)
	for k, v := range q.subs {
		resultMap[k] = v
	}
	q.subsMutex.RUnlock()
	return resultMap
}

// SubscribedTopics retruns a sorted slice of currently subscribed topics.
func (q *Quasar) SubscribedTopics() []string {
	subs := q.Subscriptions()
	topics := make([]string, len(subs))
	i := 0
	for topic := range subs {
		topics[i] = topic
		i++
	}
	sort.Strings(topics)
	return topics
}
