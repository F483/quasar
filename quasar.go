package quasar

import (
	"github.com/btcsuite/btcutil"
	"github.com/f483/dejavu"
	"sort"
	"sync"
	"time"
)

// FIXME messages and topics should be byte slices

type Hash160Digest [20]byte // ripemd160(sha256(topic))
type PubKey [65]byte        // compressed secp256k1 public key
type PeerID PubKey          // pubkey to enable authentication/encryption

type FilterStack interface {
}

type OverlayNetwork interface {
	Id() PeerID
	ConnectedPeers() []PeerID
	ReceivedEventChannel() chan *Event
	ReceivedUpdateChannel() chan *Update
	SendEvent(PeerID, Event)
	SendUpdate(PeerID, FilterStack)
	Start()
	Stop()
}

type Event struct {
	topicDigest *Hash160Digest
	message     string
	publishers  []PeerID
	ttl         uint32
}

type Update struct {
	peerId  *PeerID
	filters FilterStack
}

type Peer struct {
	id        *PeerID
	filters   FilterStack
	timestamp int64
}

type Config struct {
	DefaultEventTTL        uint32        // decremented every hop
	InputDispatcherDelay   time.Duration // in ms
	PeerFiltersExpire      uint32        // in seconds
	PublishFiltersInterval uint32        // in seconds
	HistoryLimit           uint32
	HistoryAccuracy        float64 // chance of error
	FilterStackLimit       uint32
	FilterStackDepth       uint32
	FilterStackAccuracy    float64
}

type Quasar struct {
	network               OverlayNetwork
	subs                  map[string][]chan string // topic -> subscribers
	subsMutex             *sync.RWMutex
	peers                 []Peer
	peersMutex            *sync.Mutex
	history               dejavu.DejaVu // memory of past events
	config                Config
	filters               FilterStack              // own (subs + peers)
	cachedSubDigests      map[Hash160Digest]string // digest -> topic
	cachedSubDigestsMutex *sync.RWMutex
	stopInputDispatcher   chan bool
	stopFilterPropagation chan bool
}

func Hash160(topic string) *Hash160Digest {
	digest := Hash160Digest{}
	copy(digest[:], btcutil.Hash160([]byte(topic)))
	return &digest
}

// NewEvent greats a new Event for given data.
func NewEvent(topic string, message string, ttl uint32) *Event {
	return &Event{
		topicDigest: Hash160(topic),
		message:     message,
		publishers:  []PeerID{},
		ttl:         ttl,
	}
}

func NewQuasar(network OverlayNetwork, config Config) *Quasar {

	djv := dejavu.NewProbabilistic(config.HistoryLimit, config.HistoryAccuracy)
	return &Quasar{
		network:               network,
		subs:                  make(map[string][]chan string),
		subsMutex:             new(sync.RWMutex),
		peers:                 make([]Peer, 0),
		peersMutex:            new(sync.Mutex),
		history:               djv,
		config:                config,
		filters:               nil, // FIXME implement filtres
		cachedSubDigests:      make(map[Hash160Digest]string),
		cachedSubDigestsMutex: new(sync.RWMutex),
		stopInputDispatcher:   nil, // set on Start() call
		stopFilterPropagation: nil, // set on Start() call
	}
}

func (q *Quasar) rebuildCaches() {

	// rebuild sub digest mapping
	q.cachedSubDigestsMutex.Lock()
	q.cachedSubDigests = make(map[Hash160Digest]string) // clear
	for _, sub := range q.Subscriptions() {
		q.cachedSubDigests[*Hash160(sub)] = sub
	}
	q.cachedSubDigestsMutex.Unlock()

	// TODO rebuild q.filters
}

func (q *Quasar) processUpdate(update *Update) {
	// TODO implement
}

func (q *Quasar) deliver(event *Event) {
	q.cachedSubDigestsMutex.RLock()
	if topic, ok := q.cachedSubDigests[*event.topicDigest]; ok {
		q.subsMutex.RLock()
		for _, receiver := range q.subs[topic] {
			receiver <- event.message
		}
		q.subsMutex.RUnlock()
	}
	q.cachedSubDigestsMutex.RUnlock()
}

func (q *Quasar) Publish(topic string, message string) {
	q.processEvent(NewEvent(topic, message, q.config.DefaultEventTTL))
}

func (q *Quasar) processEvent(event *Event) {
	eventData := string(event.topicDigest[:20]) + event.message
	if q.history.Witness([]byte(eventData)) {
		return // dejavu, already seen
	}

	q.deliver(event)

	// TODO propagate
}

func (q *Quasar) dispatchInput() {
	for {
		select {
		case update := <-q.network.ReceivedUpdateChannel():
			go q.processUpdate(update)
		case event := <-q.network.ReceivedEventChannel():
			go q.processEvent(event)
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

// Subscriptions retruns a sorted slice of topics currently subscribed to.
func (q *Quasar) Subscriptions() []string {
	q.subsMutex.RLock()
	topics := make([]string, len(q.subs))
	i := 0
	for topic := range q.subs {
		topics[i] = topic
		i++
	}
	sort.Strings(topics)
	q.subsMutex.RUnlock()
	return topics
}
