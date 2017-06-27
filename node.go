package quasar

import (
	"github.com/f483/dejavu"
	"io"
	"math/rand"
	"sync"
	"time"
)

// Node holds the quasar pubsup state
type Node struct {
	net               networkOverlay
	filters           [][]byte // only to prevent malloc in sendUpdates
	subscribers       map[sha256digest][]io.Writer
	mutex             *sync.RWMutex
	peers             map[pubkey]*peerData
	log               *logger
	history           dejavu.DejaVu // memory of past events
	cfg               *Config
	stopDispatcher    chan bool
	stopPropagation   chan bool
	stopExpiredPeerGC chan bool
}

type peerData struct {
	filters   [][]byte
	timestamp uint64 // unixtime
}

func (p *peerData) isExpired(c *Config) bool {
	return (newTimestampMS() - c.FilterFreshness) > p.timestamp
}

// New create instance with the sane defaults.
func New() *Node {
	// TODO pass node id/pubkey and peers list
	return newNode(nil, nil, &StandardConfig)
}

func newNode(n networkOverlay, l *logger, c *Config) *Node {
	d := dejavu.NewProbabilistic(c.HistoryLimit, c.HistoryAccuracy)
	return &Node{
		net:               n,
		filters:           newFilters(c),
		subscribers:       make(map[sha256digest][]io.Writer),
		mutex:             new(sync.RWMutex),
		peers:             make(map[pubkey]*peerData),
		log:               l,
		history:           d,
		cfg:               c,
		stopDispatcher:    nil, // set on Start() call
		stopPropagation:   nil, // set on Start() call
		stopExpiredPeerGC: nil, // set on Start() call
	}
}

func (n *Node) processUpdate(u *update) {
	n.log.updateReceived(n, u)
	if n.net.isConnected(u.NodeId) == false {
		n.log.updateFail(n, u)
		return // ignore to prevent memory attack
	}

	n.mutex.Lock()
	data, ok := n.peers[*u.NodeId]

	if !ok { // init if doesnt exist
		data = &peerData{
			filters:   newFilters(n.cfg),
			timestamp: 0,
		}
		n.peers[*u.NodeId] = data
	}

	// update peer data
	data.filters = u.Filters
	data.timestamp = newTimestampMS()
	n.mutex.Unlock()
	n.log.updateSuccess(n, u)
}

// Publish a message on the network for given topic.
func (n *Node) Publish(topic []byte, message []byte) {
	// TODO validate input
	event := newEvent(topic, message, n.cfg.DefaultEventTTL)
	n.route(event)
	n.log.eventPublished(n, event)
}

func (n *Node) isDuplicate(e *event) bool {
	return n.history.Witness(append(e.TopicDigest[:], e.Message...))
}

func (n *Node) deliver(receivers []io.Writer, e *event) {
	for _, receiver := range receivers {
		receiver.Write(e.Message)
	}
}

func (n *Node) subscriptions() []*sha256digest {
	digests := make([]*sha256digest, 0)
	for digest := range n.subscribers {
		digests = append(digests, &digest)
	}
	return digests
}

// Algorithm 1 from the quasar paper.
func (n *Node) sendUpdates() {
	n.mutex.RLock()
	clearFilters(n.filters)
	pubkey := n.net.id()
	pubkeyDigest := sha256sum(pubkey[:])
	digests := append(n.subscriptions(), &pubkeyDigest)
	n.filters[0] = newFilterFromDigests(n.cfg, digests...)
	for _, data := range n.peers {
		// XXX better if only expiredPeerGC takes care of it?
		// if data.isExpired(n.cfg) {
		// 		continue
		// 	}
		for i := 1; uint32(i) < n.cfg.FiltersDepth; i++ {
			size := int(n.cfg.FiltersM / 8)
			for j := 0; j < size; j++ { // inline merge for performance
				n.filters[i][j] = n.filters[i][j] | data.filters[i-1][j]
			}
		}
	}
	for _, id := range n.net.connectedPeers() {
		// top filter never sent as not used by peers
		n.net.sendUpdate(id, n.filters)
		n.log.updateSent(n, n.filters, id)
	}
	n.mutex.RUnlock()
}

// Algorithm 2 from the quasar paper.
func (n *Node) route(e *event) {
	n.mutex.RLock()
	defer n.mutex.RUnlock()
	nodeId := n.net.id()
	if n.isDuplicate(e) {
		n.log.eventDropDuplicate(n, e)
		return
	}
	if receivers, ok := n.subscribers[*e.TopicDigest]; ok {
		n.deliver(receivers, e)
		n.log.eventDeliver(n, e)
		e.addPublisher(&nodeId)
		for _, peerId := range n.net.connectedPeers() {
			if !e.inPublishers(peerId) {
				n.net.sendEvent(peerId, e)
				n.log.eventRouteDirect(n, e, peerId)
			}
		}
		return
	}
	e.Ttl -= 1
	if e.Ttl == 0 {
		n.log.eventDropTTL(n, e)
		return
	}
	for i := 0; uint32(i) < n.cfg.FiltersDepth; i++ {
		for peerId, data := range n.peers {
			f := data.filters[i]
			if filterContainsDigest(f, n.cfg, e.TopicDigest) {
				negRt := false
				for _, publisher := range e.Publishers {
					if filterContainsDigest(f, n.cfg, publisher) {
						negRt = true
					}
				}
				if !negRt && !e.inPublishers(&peerId) {
					n.net.sendEvent(&peerId, e)
					n.log.eventRouteWell(n, e, &peerId)
					return
				}
			}
		}
	}
	n.sendToRandomPeer(e)
}

func (n *Node) sendToRandomPeer(e *event) {
	peers := make([]*pubkey, 0)
	for _, peerId := range n.net.connectedPeers() {
		if !e.inPublishers(peerId) {
			peers = append(peers, peerId)
		}
	}
	if len(peers) > 0 {
		peerId := peers[rand.Intn(len(peers))]
		n.net.sendEvent(peerId, e)
		n.log.eventRouteRandom(n, e, peerId)
	}
}

func (n *Node) dispatchInput() {
	for {
		select {
		case u := <-n.net.receivedUpdateChannel():
			if u.valid(n.cfg) {
				n.processUpdate(u)
			}
		case e := <-n.net.receivedEventChannel():
			if e.valid() {
				n.log.eventReceived(n, e)
				n.route(e)
			}
		case <-n.stopDispatcher:
			return
		}
	}
}

func (n *Node) removeExpiredPeers() {
	n.mutex.Lock()
	toRemove := []*pubkey{}
	for peerId, data := range n.peers {
		if data.isExpired(n.cfg) {
			toRemove = append(toRemove, &peerId)
		}
	}
	for _, peerId := range toRemove {
		delete(n.peers, *peerId)
	}
	n.mutex.Unlock()
}

func (n *Node) expiredPeerGC() {
	delay := time.Duration(n.cfg.FilterFreshness/2) * time.Millisecond
	for {
		select {
		case <-time.After(delay):
			n.removeExpiredPeers()
		case <-n.stopExpiredPeerGC:
			return
		}
	}
}

func (n *Node) propagateFilters() {
	delay := time.Duration(n.cfg.PropagationDelay) * time.Millisecond
	for {
		select {
		case <-time.After(delay):
			n.sendUpdates()
		case <-n.stopPropagation:
			return
		}
	}
}

// Start quasar system
func (n *Node) Start() {
	n.net.start()
	n.stopDispatcher = make(chan bool)
	n.stopPropagation = make(chan bool)
	n.stopExpiredPeerGC = make(chan bool)
	go n.dispatchInput()
	go n.propagateFilters()
	go n.expiredPeerGC()
}

// Stop quasar system
func (n *Node) Stop() {
	n.net.stop()
	n.stopDispatcher <- true
	n.stopPropagation <- true
	n.stopExpiredPeerGC <- true
}

// Subscribe provided message receiver to given topic.
func (n *Node) Subscribe(topic []byte, receiver io.Writer) {
	// TODO validate input
	digest := sha256sum(topic)
	n.SubscribeDigest(&digest, receiver)
}

// Subscribe provided message receiver to given topic digest.
func (n *Node) SubscribeDigest(digest *sha256digest, receiver io.Writer) {
	// TODO validate input
	n.mutex.Lock()
	receivers, ok := n.subscribers[*digest]
	if ok != true { // new subscription
		n.subscribers[*digest] = []io.Writer{receiver}
	} else { // append to existing subscribers
		n.subscribers[*digest] = append(receivers, receiver)
	}
	n.mutex.Unlock()
}

// Unsubscribe message receiver channel from topic. If nil receiver
// channel is provided all message receiver channels for given topic
// will be removed.
func (n *Node) Unsubscribe(topic []byte, receiver io.Writer) {
	// TODO validate input

	digest := sha256sum(topic)
	n.mutex.Lock()
	receivers, ok := n.subscribers[digest]

	// remove specific message receiver
	if ok && receiver != nil {
		for i, v := range receivers {
			if v == receiver {
				receivers = append(receivers[:i], receivers[i+1:]...)
				n.subscribers[digest] = receivers
				break
			}
		}
	}

	// remove sub key if no specific message
	// receiver provided or no message receiver remaining
	if ok && (receiver == nil || len(n.subscribers[digest]) == 0) {
		delete(n.subscribers, digest)
	}
	n.mutex.Unlock()
}

// Subscribed returns true if node is subscribed to given topic.
func (n *Node) Subscribed(topic []byte) bool {
	digest := sha256sum(topic)
	return n.SubscribedDigest(&digest)
}

// Subscribed returns true if node is subscribed to given topic digest.
func (n *Node) SubscribedDigest(digest *sha256digest) bool {
	n.mutex.RLock()
	_, ok := n.subscribers[*digest]
	n.mutex.RUnlock()
	return ok
}

// Subscribers retruns message receivers for given topic.
func (n *Node) Subscribers(topic []byte) []io.Writer {
	// TODO validate input
	digest := sha256sum(topic)
	results := []io.Writer{}
	n.mutex.RLock()
	if receivers, ok := n.subscribers[digest]; ok {
		results = append(results, receivers...)
	}
	n.mutex.RUnlock()
	return results
}
