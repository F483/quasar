package quasar

import (
	"github.com/btcsuite/btcutil"
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

type Event struct {
	topicDigest *TopicDigest
	message     *string
	publishers  *[]PeerID
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
		message:     &message,
		publishers:  &[]PeerID{},
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
}

type Quasar struct {
	network       OverlayNetwork
	filters       Filters                  // own filters
	subscriptions map[string][]chan string // topic name -> message channel
	peers         []Peer                   // known peer filters
}

func (q Quasar) Subscribe(topic string, msgReceiver chan string) {

}

func (q Quasar) Unsubscribe(topic string, msgReceiver chan string) {

}

func (q Quasar) Subscriptions() []string {
	return nil
}
