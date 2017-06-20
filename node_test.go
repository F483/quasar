package quasar

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestSubscriptions(t *testing.T) {
	n := newNode(nil, nil, &testConfig)

	var a bytes.Buffer
	n.Subscribe([]byte("a"), &a)
	subs := n.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	var b1 bytes.Buffer
	n.Subscribe([]byte("b"), &b1)
	subs = n.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	var b2 bytes.Buffer
	n.Subscribe([]byte("b"), &b2)
	subs = n.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	var c1 bytes.Buffer
	n.Subscribe([]byte("c"), &c1)
	subs = n.Subscriptions()
	expectedSubs := [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	if !checkSubs(subs, expectedSubs) {
		t.Errorf("Incorrect subscriptions!")
	}

	var c2 bytes.Buffer
	n.Subscribe([]byte("c"), &c2)
	subs = n.Subscriptions()
	expectedSubs = [][]byte{[]byte("a"), []byte("b"), []byte("c")}
	if !checkSubs(subs, expectedSubs) {
		t.Errorf("Incorrect subscriptions!")
	}

	// test clears all if no receiver provided
	n.Unsubscribe([]byte("c"), nil)
	subs = n.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}

	// only clears specific receiver
	if len(n.Subscribers([]byte("b"))) != 2 {
		t.Errorf("Incorrect subscribers!")
	}
	n.Unsubscribe([]byte("b"), &b1)
	subs = n.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a"), []byte("b")}) {
		t.Errorf("Incorrect subscriptions!")
	}
	if len(n.Subscribers([]byte("b"))) != 1 {
		t.Errorf("Incorrect subscribers!")
	}

	// clears key when last receiver removed
	n.Unsubscribe([]byte("b"), &b2)
	subs = n.Subscriptions()
	if !checkSubs(subs, [][]byte{[]byte("a")}) {
		t.Errorf("Incorrect subscriptions!")
	}
}

func checkSubs(given [][]byte, expected [][]byte) bool {
	if len(given) != len(expected) {
		return false
	}
	for _, e := range expected {
		found := false
		for _, g := range given {
			if reflect.DeepEqual(e, g) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestEventDelivery(t *testing.T) {
	cfg := &testConfig

	// stopa := make(chan bool)
	// stopb := make(chan bool)
	// la := LogToConsole("Subscriber", stopa)
	// lb := LogToConsole("Publisher", stopb)

	nodes := newMockNetwork(nil, cfg, 2)
	// nodes[0].log = la
	// nodes[1].log = lb

	// set subscriptions
	var fooReceiver bytes.Buffer
	nodes[0].Subscribe([]byte("foo"), &fooReceiver)

	// start nodes and wait for filters to propagate
	for _, node := range nodes {
		node.Start()
	}
	time.Sleep(time.Millisecond * time.Duration(cfg.PropagationDelay*2))

	// create event
	nodes[1].Publish([]byte("foo"), []byte("foodata"))
	time.Sleep(time.Millisecond * time.Duration(cfg.PropagationDelay*2))

	if !reflect.DeepEqual(fooReceiver.Bytes(), []byte("foodata")) {
		t.Errorf("Incorrect event data!")
	}

	// stop nodes and logger
	// stopa <- true
	// stopb <- true
	for _, node := range nodes {
		node.Stop()
	}
}

func TestEventTimeout(t *testing.T) {
	// get coverage for dropping ttl = 0 events
	// may be dropped by history with few nodes

	cfg := testConfig
	cfg.DefaultEventTTL = 1

	nodes := newMockNetwork(nil, &cfg, 20)

	// start nodes and wait for filters to propagate
	for _, node := range nodes {
		node.Start()
	}
	time.Sleep(time.Millisecond * time.Duration(cfg.PropagationDelay))

	// create event
	nodes[1].Publish([]byte("bar"), []byte("bardata"))
	time.Sleep(time.Duration(cfg.PropagationDelay) * time.Millisecond)

	// stop nodes
	for _, node := range nodes {
		node.Stop()
	}
}

func TestExpiredPeerData(t *testing.T) {
	cfg := &testConfig

	nodes := newMockNetwork(nil, cfg, 2)

	// start nodes and wait for filters to propagate
	nodes[0].Start()
	nodes[1].Start()

	// let filters propagate
	time.Sleep(time.Millisecond * time.Duration(cfg.PropagationDelay*2))

	nodes[0].Stop()

	// let filters expire
	time.Sleep(time.Millisecond * time.Duration(cfg.FilterFreshness*2))

	nodes[1].Stop()
}

func TestNoPeers(t *testing.T) {
	cfg := &testConfig

	nodes := newMockNetwork(nil, cfg, 1)
	nodes[0].Start()
	nodes[0].Publish([]byte("bar"), []byte("bardata"))
	nodes[0].Stop()
}
