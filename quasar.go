package quasar

import (
	"github.com/btcsuite/btcutil"
	"sort"
	"sync"
	"time"
)

// FIXME messages and topics should be byte splices

// FIXME move constants to config
const TTL uint32 = 128        // max event hops before being dropped
const DEPTH uint32 = 128      // attenuated bloom filter depth
const SIZE uint32 = 512       // bloom filter size
const REFRESH uint32 = 600    // interval when own filters are published
const EXPIRE uint32 = 660     // time until peer filters become stale
const HISTORY uint32 = 65536  // historic memory size to prevent duplicats
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
	network               OverlayNetwork
	subs                  map[string][]chan string // topic -> subscribers
	subsMutex             *sync.RWMutex
	peers                 []Peer // know peer filters and time last seen
	peersMutex            *sync.Mutex
	history               map[Hash160Digest]int64 // msg digest -> timestamp
	histroyMutex          *sync.Mutex
	cacheFilters          *Filters // own filters (subs + peers)
	cacheFiltersMutex     *sync.RWMutex
	cacheSubDigests       map[Hash160Digest]string // digest -> topic
	cacheSubDigestsMutex  *sync.RWMutex
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

func NewQuasar(network OverlayNetwork) *Quasar {
	return &Quasar{
		network:               network,
		subs:                  make(map[string][]chan string),
		subsMutex:             new(sync.RWMutex),
		peers:                 make([]Peer, 0),
		peersMutex:            new(sync.Mutex),
		history:               make(map[Hash160Digest]int64),
		histroyMutex:          new(sync.Mutex),
		cacheFilters:          new(Filters),
		cacheFiltersMutex:     new(sync.RWMutex),
		cacheSubDigests:       make(map[Hash160Digest]string),
		cacheSubDigestsMutex:  new(sync.RWMutex),
		stopUpdateDispatcher:  make(chan bool),
		stopEventDispatcher:   make(chan bool),
		stopFilterPropagation: make(chan bool),
	}
}

func (q *Quasar) rebuildCaches() {

	// rebuild sub digest mapping
	q.cacheSubDigestsMutex.Lock()
	q.cacheSubDigests = make(map[Hash160Digest]string) // clear
	for _, sub := range q.Subscriptions() {
		q.cacheSubDigests[*Hash160(sub)] = sub
	}
	q.cacheSubDigestsMutex.Unlock()

	// TODO rebuild q.cacheFilters
}

func (q *Quasar) cullHistory() {
	// TODO implement
}

func (q *Quasar) updateHistory(event *Event) bool {

	q.histroyMutex.Lock()
	digest := Hash160(event.message) // FIXME use topic + message
	updated := false
	if q.history[*digest] == 0 {
		q.history[*digest] = time.Now().Unix()
		go q.cullHistory()
		updated = true
	}
	q.histroyMutex.Lock()
	return updated
}

func (q *Quasar) processUpdate(update *Update) {
	// TODO implement
}

func (q *Quasar) deliver(event *Event) {
	q.cacheSubDigestsMutex.RLock()
	if topic, ok := q.cacheSubDigests[*event.topicDigest]; ok {
		q.subsMutex.RLock()
		for _, receiver := range q.subs[topic] {
			receiver <- event.message
		}
		q.subsMutex.RUnlock()
	}
	q.cacheSubDigestsMutex.RUnlock()
}

func (q *Quasar) Publish(topic string, message string) {
	q.processEvent(NewEvent(topic, message))
}

func (q *Quasar) processEvent(event *Event) {

	if q.updateHistory(event) == false {
		return // already processed
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
			return
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
			return
		}
		time.Sleep(DELAY * time.Millisecond)
	}
}

func (q *Quasar) propagateFilters() {
	// TODO implement
}

func (q *Quasar) Start() {
	q.network.Start()
	go q.dispatchUpdates()
	go q.dispatchEvents()
	go q.propagateFilters()
}

func (q *Quasar) Stop() {

	q.network.Stop()

	// TODO set flags for workers to stop
	// TODO wait for workers to stop?
}

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
