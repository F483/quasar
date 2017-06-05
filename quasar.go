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
	topicSubscribers      map[hash160digest][]chan []byte // digest -> subscribers
	topicDigests          map[hash160digest]string        // digest -> topic
	topicMutex            *sync.RWMutex
	peers                 []peer
	peersMutex            *sync.Mutex
	history               dejavu.DejaVu // memory of past events
	config                Config
	filters               [][]byte // own (subs + peers)
	stopInputDispatcher   chan bool
	stopFilterPropagation chan bool
}

func newEvent(topic string, message []byte, ttl uint32) *event {
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
		network:               network,
		topicSubscribers:      make(map[hash160digest][]chan []byte),
		topicDigests:          make(map[hash160digest]string),
		topicMutex:            new(sync.RWMutex),
		peers:                 make([]peer, 0),
		peersMutex:            new(sync.Mutex),
		history:               d,
		config:                c,
		filters:               nil,
		stopInputDispatcher:   nil, // set on Start() call
		stopFilterPropagation: nil, // set on Start() call
	}
}

func (q *Quasar) processUpdate(u *update) {
	// TODO implement
}

func (q *Quasar) deliver(e *event) {
	q.topicMutex.RLock()
	if receivers, ok := q.topicSubscribers[*e.topicDigest]; ok {
		for _, receiver := range receivers {
			receiver <- e.message
		}
	}
	q.topicMutex.RUnlock()
}

func (q *Quasar) Publish(topic string, message []byte) {
	q.processEvent(newEvent(topic, message, q.config.DefaultEventTTL))
}

func (q *Quasar) processEvent(e *event) {
	eventData := append(e.topicDigest[:20], e.message...)
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
func (q *Quasar) Subscribe(topic string, msgReceiver chan []byte) {

	digest := hash160(topic)
	q.topicMutex.Lock()

	receivers, ok := q.topicSubscribers[digest]
	if ok != true { // new subscription
		q.topicSubscribers[digest] = []chan []byte{msgReceiver}
		q.topicDigests[digest] = topic
	} else { // append to existing subscribers
		q.topicSubscribers[digest] = append(receivers, msgReceiver)
	}

	q.topicMutex.Unlock()
}

// Unsubscribe message receiver channel from topic. If nil receiver channel is
// provided all message receiver channels for given topic will be removed.
func (q *Quasar) Unsubscribe(topic string, msgReceiver chan []byte) {

	digest := hash160(topic)
	q.topicMutex.Lock()
	receivers, ok := q.topicSubscribers[digest]

	// remove specific message receiver
	if ok && msgReceiver != nil {
		for i, v := range receivers {
			if v == msgReceiver {
				receivers = append(receivers[:i], receivers[i+1:]...)
				q.topicSubscribers[digest] = receivers
				break
			}
		}
	}

	// remove sub key if no specific message
	// receiver provided or no message receiver remaining
	if ok && (msgReceiver == nil || len(q.topicSubscribers[digest]) == 0) {
		delete(q.topicSubscribers, digest)
		delete(q.topicDigests, digest)
	}
	q.topicMutex.Unlock()
}

// Subscribers retruns message receivers for given topic.
func (q *Quasar) Subscribers(topic string) []chan []byte {
	q.topicMutex.RLock()

	// TODO implement

	q.topicMutex.RUnlock()
	return nil
}

// SubscribedTopics retruns a sorted slice of currently subscribed topics.
func (q *Quasar) Subscriptions() []string {
	q.topicMutex.RLock()
	topics := make([]string, len(q.topicDigests))
	i := 0
	for _, topic := range q.topicDigests {
		topics[i] = topic
		i++
	}
	q.topicMutex.RUnlock()
	sort.Strings(topics)
	return topics
}
