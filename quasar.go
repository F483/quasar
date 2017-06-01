package quasar

import (
	"github.com/btcsuite/btcutil"
	"github.com/f483/dejavu"
	"sort"
	"sync"
	"time"
)

// FIXME messages and topics should be byte slices

// FIXME move constants to config
const TTL uint32 = 128        // max event hops before being dropped
const DEPTH uint32 = 128      // attenuated bloom filter depth
const SIZE uint32 = 512       // bloom filter size
const REFRESH uint32 = 600    // interval when own filters are published
const EXPIRE uint32 = 660     // time until peer filters become stale
const DELAY time.Duration = 1 // event/update dispatcher sleep time in ms

type Filters [DEPTH][SIZE]byte
type Hash160Digest [20]byte // ripemd160(sha256(topic))
type PubKey [65]byte        // compressed secp256k1 public key
type PeerID PubKey          // pubkey to enable authentication/encryption

type Event struct {
	topicDigest *Hash160Digest
	message     string
	publishers  []PeerID
	ttl         uint32
}

type Update struct {
	peerId  *PeerID
	filters *Filters
}

type Peer struct {
	id        *PeerID
	filters   *Filters
	timestamp int64
}

type OverlayNetwork interface {
	Id() PeerID
	ConnectedPeers() []PeerID
	ReceivedEventChannel() chan *Event
	ReceivedUpdateChannel() chan *Update
	SendEvent(PeerID, Event)
	SendUpdate(PeerID, Filters)
	Start()
	Stop()
}

type Quasar struct {
	network    OverlayNetwork
	subs       map[string][]chan string // topic -> subscribers
	subsMutex  *sync.RWMutex
	peers      []Peer // know peer filters and time last seen
	peersMutex *sync.Mutex
	history    dejavu.DejaVu

	cachedFilters         *Filters // own filters (subs + peers)
	cachedFiltersMutex    *sync.RWMutex
	cachedSubDigests      map[Hash160Digest]string // digest -> topic
	cachedSubDigestsMutex *sync.RWMutex
	stopEventDispatcher   chan bool
	stopUpdateDispatcher  chan bool
	stopFilterPropagation chan bool
}

func Hash160(topic string) *Hash160Digest {
	digest := Hash160Digest{}
	copy(digest[:], btcutil.Hash160([]byte(topic)))
	return &digest
}

func NewEvent(topic string, message string) *Event {
	return &Event{
		topicDigest: Hash160(topic),
		message:     message,
		publishers:  []PeerID{},
		ttl:         TTL,
	}
}

func NewQuasar(
	network OverlayNetwork, historyLimit uint32, historyAccuracy float64,
) *Quasar {

	djv := dejavu.NewProbabilistic(historyLimit, historyAccuracy)
	return &Quasar{
		network:               network,
		subs:                  make(map[string][]chan string),
		subsMutex:             new(sync.RWMutex),
		peers:                 make([]Peer, 0),
		peersMutex:            new(sync.Mutex),
		history:               djv,
		cachedFilters:         new(Filters),
		cachedFiltersMutex:    new(sync.RWMutex),
		cachedSubDigests:      make(map[Hash160Digest]string),
		cachedSubDigestsMutex: new(sync.RWMutex),
		stopUpdateDispatcher:  nil, // set on Start() call
		stopEventDispatcher:   nil, // set on Start() call
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

	// TODO rebuild q.cachedFilters
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
	q.processEvent(NewEvent(topic, message))
}

func (q *Quasar) processEvent(event *Event) {
	eventData := string(event.topicDigest[:20]) + event.message
	if q.history.Witness([]byte(eventData)) {
		return // dejavu, already seen
	}

	q.deliver(event)

	// TODO propagate
}

func (q *Quasar) dispatchUpdates() {
	for {
		select {
		case update := <-q.network.ReceivedUpdateChannel():
			q.processUpdate(update)
		case <-q.stopUpdateDispatcher:
			return // TODO confirm stopped
		}
		time.Sleep(DELAY * time.Millisecond)
	}
}

func (q *Quasar) dispatchEvents() {
	for {
		select {
		case event := <-q.network.ReceivedEventChannel():
			q.processEvent(event)
		case <-q.stopEventDispatcher:
			return // TODO confirm stopped
		}
		time.Sleep(DELAY * time.Millisecond)
	}
}

func (q *Quasar) propagateFilters() {
	// TODO implement
}

// Start quasar system
func (q *Quasar) Start() {
	q.network.Start()
	q.stopUpdateDispatcher = make(chan bool)
	q.stopEventDispatcher = make(chan bool)
	q.stopFilterPropagation = make(chan bool)
	go q.dispatchUpdates()
	go q.dispatchEvents()
	go q.propagateFilters()
}

// Stop quasar system
func (q *Quasar) Stop() {
	q.network.Stop()
	q.stopEventDispatcher <- true
	q.stopUpdateDispatcher <- true
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

// Unsubscribe from message receiver channel from topic. If nil channel is
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
