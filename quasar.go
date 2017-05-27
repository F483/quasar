package quasar

import (
	"github.com/btcsuite/btcutil"
	"sort"
	"sync"
	"time"
)

const TTL uint32 = 128       // max event hops before being dropped
const DEPTH uint32 = 128     // attenuated bloom filter depth
const SIZE uint32 = 512      // bloom filter size
const REFRESH uint32 = 600   // interval when own filters are rese
const EXPIRE uint32 = 660    // time until peer filters become stale
const HISTORY uint32 = 65536 // historic memory size to prevent duplicats

type Filters [DEPTH][SIZE]byte
type Hash160Digest [20]byte // ripemd160(sha256(topic))
type PubKey [65]byte        // compressed secp256k1 public key
type PeerID PubKey          // pubkey to enable authentication/encryption
type Subscriptions map[string][]chan string

type Event struct {
	topicDigest *Hash160Digest
	message     string
	publishers  []PeerID
	ttl         uint32
}

type OverlayNetwork interface {
	Id() PeerID
	ConnectedPeers() []PeerID
	ReceivedEventChannel() chan Event
	ReceivedUpdateChannel() chan Update
	SendEvent(PeerID, Event)
	SendUpdate(PeerID, Filters)
	Start()
	Stop()
}

type Quasar struct {
	network             OverlayNetwork
	filters             *Filters // own filters
	filters_mutex       *sync.Mutex
	subscriptions       Subscriptions // topic name -> message chan
	subscriptions_mutex *sync.RWMutex
	peers               []Peer // known peers
	peers_mutex         *sync.Mutex
	history             map[Hash160Digest]int64 // message digest -> timestamp
	histroy_mutex       *sync.Mutex
	// TODO add flags for workers to stop
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

type Update struct {
	peerId  *PeerID
	filters *Filters
}

type Peer struct {
	id        *PeerID
	filters   *Filters
	timestamp int64
}

func NewQuasar(network OverlayNetwork) *Quasar {
	return &Quasar{
		network:             network,
		filters:             new(Filters),
		filters_mutex:       new(sync.Mutex),
		subscriptions:       Subscriptions{},
		subscriptions_mutex: new(sync.RWMutex),
		peers:               []Peer{},
		peers_mutex:         new(sync.Mutex),
		history:             map[Hash160Digest]int64{},
		histroy_mutex:       new(sync.Mutex),
	}
}

func (q *Quasar) updateHistory(message string) bool {
	q.histroy_mutex.Lock()
	digest := Hash160(message)
	updated := false
	if q.history[*digest] == 0 {
		q.history[*digest] = time.Now().Unix()
		updated = true
	}
	q.histroy_mutex.Lock()
	return updated
}

func (q *Quasar) Start() {

	q.network.Start()

	// TODO reset worker flags
	// TODO start process received events worker
	// TODO start process received updates worker
	// TODO start propagate filters worker
	// TODO start cull history worker
}

func (q *Quasar) Stop() {

	q.network.Stop()

	// TODO set flags for workers to stop
	// TODO wait for workers to stop?
}

func (q *Quasar) Subscribe(topic string, msgReceiver chan string) {
	q.subscriptions_mutex.Lock()
	receivers, ok := q.subscriptions[topic]
	if ok != true {
		q.subscriptions[topic] = []chan string{msgReceiver}
	} else {
		q.subscriptions[topic] = append(receivers, msgReceiver)
	}
	q.subscriptions_mutex.Unlock()
}

func (q *Quasar) Unsubscribe(topic string, msgReceiver chan string) {
	q.subscriptions_mutex.Lock()
	receivers, ok := q.subscriptions[topic]

	// remove specific message receiver
	if ok && msgReceiver != nil {
		for i, v := range receivers {
			if v == msgReceiver {
				receivers = append(receivers[:i], receivers[i+1:]...)
				q.subscriptions[topic] = receivers
				break
			}
		}
	}

	// remove subscription key if no specific message
	// receiver provided or no message receiver remaining
	if ok && (msgReceiver == nil || len(q.subscriptions[topic]) == 0) {
		delete(q.subscriptions, topic)
	}
	q.subscriptions_mutex.Unlock()
}

func (q *Quasar) Subscriptions() []string {
	q.subscriptions_mutex.RLock()
	topics := make([]string, len(q.subscriptions))
	i := 0
	for topic := range q.subscriptions {
		topics[i] = topic
		i++
	}
	sort.Strings(topics)
	q.subscriptions_mutex.RUnlock()
	return topics
}
