package quasar

import (
	"github.com/btcsuite/btcutil"
	"sort"
)

const TTL uint32 = 128       // Hops event can travel before being dropped.
const DEPTH uint32 = 128     // Attenuated bloom filter depth.
const SIZE uint32 = 512      // Bloom filter size.
const REFRESH uint32 = 600   // Interval when own filters are rebuilt.
const FRESHNESS uint32 = 660 // Time after which peer filters become stale.

type Filters [DEPTH][SIZE]byte
type TopicDigest [20]byte // ripemd160(sha256(topic))
type PubKey [65]byte      // compressed secp256k1 public key
type PeerID PubKey        // pubkey to enable authentication/encryption
type Subscriptions map[string][]chan string

type Event struct {
	topicDigest *TopicDigest
	message     string
	publishers  []PeerID
	ttl         uint32
}

func HashTopic(topic string) *TopicDigest {
	digest := TopicDigest{}
	copy(digest[:], btcutil.Hash160([]byte(topic)))
	return &digest
}

func NewEvent(topic string, message string) *Event {
	return &Event{
		topicDigest: HashTopic(topic),
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
	timestamp uint32
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
	network       OverlayNetwork
	filters       *Filters           // own filters
	subscriptions Subscriptions      // topic name -> message chan
	peers         map[PeerID]Filters // known peer filters
	// TODO add flags for workers to stop
}

func NewQuasar(network OverlayNetwork) *Quasar {
	q := Quasar{
		network:       network,
		filters:       new(Filters),
		subscriptions: Subscriptions{},
		peers:         map[PeerID]Filters{},
	}
	return &q
}

func (q *Quasar) Start() {
	// TODO make thread safe
	q.network.Start()
	// TODO reset worker flags
	// TODO start process received events worker
	// TODO start process received updates worker
	// TODO start propagate filters worker
}

func (q *Quasar) Stop() {
	// TODO make thread safe
	q.network.Stop()
	// TODO set flags for workers to stop
	// TODO wait for workers to stop?
}

func (q *Quasar) Subscribe(topic string, msgReceiver chan string) {
	// TODO make thread safe
	receivers, ok := q.subscriptions[topic]
	if ok != true {
		q.subscriptions[topic] = []chan string{msgReceiver}
	} else {
		q.subscriptions[topic] = append(receivers, msgReceiver)
	}
}

func (q *Quasar) Unsubscribe(topic string, msgReceiver chan string) {
	// TODO make thread safe

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
	if ok && (len(q.subscriptions[topic]) == 0 || msgReceiver == nil) {
		delete(q.subscriptions, topic)
	}
}

func (q *Quasar) Subscriptions() []string {
	// TODO make thread safe
	topics := make([]string, len(q.subscriptions))
	i := 0
	for topic := range q.subscriptions {
		topics[i] = topic
		i++
	}
	sort.Strings(topics)
	return topics
}
