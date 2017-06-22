package quasar

import (
	"fmt"
	// "time"
)

type logUpdate struct {
	node   *Node
	entry  *peerUpdate
	target *pubkey
}

type logEvent struct {
	node   *Node
	entry  *event
	target *pubkey
}

type logger struct {
	updatesSent         chan *logUpdate
	updatesReceived     chan *logUpdate
	updatesSuccess      chan *logUpdate // added to filters
	updatesFail         chan *logUpdate // not from neighbour
	eventsPublished     chan *logEvent
	eventsReceived      chan *logEvent
	eventsDeliver       chan *logEvent
	eventsDropDuplicate chan *logEvent
	eventsDropTTL       chan *logEvent
	eventsRouteDirect   chan *logEvent
	eventsRouteWell     chan *logEvent
	eventsRouteRandom   chan *logEvent
	// TODO add overlay network logging
}

func newLogger(bufsize int) *logger {
	return &logger{
		updatesSent:         make(chan *logUpdate, bufsize),
		updatesReceived:     make(chan *logUpdate, bufsize),
		updatesSuccess:      make(chan *logUpdate, bufsize),
		updatesFail:         make(chan *logUpdate, bufsize),
		eventsPublished:     make(chan *logEvent, bufsize),
		eventsReceived:      make(chan *logEvent, bufsize),
		eventsDeliver:       make(chan *logEvent, bufsize),
		eventsDropDuplicate: make(chan *logEvent, bufsize),
		eventsDropTTL:       make(chan *logEvent, bufsize),
		eventsRouteDirect:   make(chan *logEvent, bufsize),
		eventsRouteWell:     make(chan *logEvent, bufsize),
		eventsRouteRandom:   make(chan *logEvent, bufsize),
	}
}

func (l *logger) updateSent(n *Node, i uint32, f []byte, t *pubkey) {
	if l != nil && l.updatesSent != nil {
		var id *pubkey
		if n != nil {
			idv := n.net.id()
			id = &idv
		}
		u := &peerUpdate{peer: id, index: i, filter: f}
		l.updatesSent <- &logUpdate{
			node: n, entry: u, target: t,
		}
	}
}

func (l *logger) updateReceived(n *Node, u *peerUpdate) {
	if l != nil && l.updatesReceived != nil {
		l.updatesReceived <- &logUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *logger) updateSuccess(n *Node, u *peerUpdate) {
	if l != nil && l.updatesSuccess != nil {
		l.updatesSuccess <- &logUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *logger) updateFail(n *Node, u *peerUpdate) {
	if l != nil && l.updatesFail != nil {
		l.updatesFail <- &logUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *logger) eventPublished(n *Node, e *event) {
	if l != nil && l.eventsPublished != nil {
		l.eventsPublished <- &logEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *logger) eventReceived(n *Node, e *event) {
	if l != nil && l.eventsReceived != nil {
		l.eventsReceived <- &logEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *logger) eventDeliver(n *Node, e *event) {
	if l != nil && l.eventsDeliver != nil {
		l.eventsDeliver <- &logEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *logger) eventDropDuplicate(n *Node, e *event) {
	if l != nil && l.eventsDropDuplicate != nil {
		l.eventsDropDuplicate <- &logEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *logger) eventDropTTL(n *Node, e *event) {
	if l != nil && l.eventsDropTTL != nil {
		l.eventsDropTTL <- &logEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *logger) eventRouteDirect(n *Node, e *event, t *pubkey) {
	if l != nil && l.eventsRouteDirect != nil {
		l.eventsRouteDirect <- &logEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *logger) eventRouteWell(n *Node, e *event, t *pubkey) {
	if l != nil && l.eventsRouteWell != nil {
		l.eventsRouteWell <- &logEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *logger) eventRouteRandom(n *Node, e *event, t *pubkey) {
	if l != nil && l.eventsRouteRandom != nil {
		l.eventsRouteRandom <- &logEvent{
			node: n, entry: e, target: t,
		}
	}
}

func printlogUpdate(prefix string, src string, lu *logUpdate) {
	fmt.Printf("%s: %s\n", prefix, src)
	// TODO log more info
}

func printlogEvent(prefix string, src string, le *logEvent) {
	fmt.Printf("%s: %s\n", prefix, src)
	// TODO log more info
}

func logToConsole(prefix string, stopLogging chan bool) *logger {
	mustNotBeNil(stopLogging)
	l := newLogger(10)
	go func() {
		for {
			select {
			case lu := <-l.updatesSent:
				go printlogUpdate(prefix, "updatesSent", lu)
			case lu := <-l.updatesReceived:
				go printlogUpdate(prefix, "updatesReceived", lu)
			case lu := <-l.updatesSuccess:
				go printlogUpdate(prefix, "updatesSuccess", lu)
			case lu := <-l.updatesFail:
				go printlogUpdate(prefix, "updatesFail", lu)
			case le := <-l.eventsPublished:
				go printlogEvent(prefix, "eventsPublished", le)
			case le := <-l.eventsReceived:
				go printlogEvent(prefix, "eventsReceived", le)
			case le := <-l.eventsDeliver:
				go printlogEvent(prefix, "eventsDeliver", le)
			case le := <-l.eventsDropDuplicate:
				go printlogEvent(prefix, "eventsDropDuplicate", le)
			case le := <-l.eventsDropTTL:
				go printlogEvent(prefix, "eventsDropTTL", le)
			case le := <-l.eventsRouteDirect:
				go printlogEvent(prefix, "eventsRouteDirect", le)
			case le := <-l.eventsRouteWell:
				go printlogEvent(prefix, "eventsRouteWell", le)
			case le := <-l.eventsRouteRandom:
				go printlogEvent(prefix, "eventsRouteRandom", le)
			case <-stopLogging:
				go fmt.Println("Stop logging received!")
				return
			}
			// time.Sleep(time.Millisecond)
		}
	}()
	return l
}
