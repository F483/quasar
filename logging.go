package quasar

import (
	"fmt"
	// "time"
)

// FIXME dont export logging

// LogUpdate used for monitoring internal filter updates.
type LogUpdate struct {
	node   *Node
	entry  *peerUpdate
	target *pubkey
}

// LogEvent used for monitoring internal events.
type LogEvent struct {
	node   *Node
	entry  *event
	target *pubkey
}

// Logger provides a logger used by Node nodes for logging internals.
type Logger struct {
	UpdatesSent         chan *LogUpdate
	UpdatesReceived     chan *LogUpdate
	UpdatesSuccess      chan *LogUpdate // added to filters
	UpdatesFail         chan *LogUpdate // not from neighbour
	EventsPublished     chan *LogEvent
	EventsReceived      chan *LogEvent
	EventsDeliver       chan *LogEvent
	EventsDropDuplicate chan *LogEvent
	EventsDropTTL       chan *LogEvent
	EventsRouteDirect   chan *LogEvent
	EventsRouteWell     chan *LogEvent
	EventsRouteRandom   chan *LogEvent
	// TODO add overlay network logging
}

// NewLogger creats a new default logger instance.
func NewLogger(bufsize int) *Logger {
	// TODO add chan buffer size to args
	return &Logger{
		UpdatesSent:         make(chan *LogUpdate, bufsize),
		UpdatesReceived:     make(chan *LogUpdate, bufsize),
		UpdatesSuccess:      make(chan *LogUpdate, bufsize),
		UpdatesFail:         make(chan *LogUpdate, bufsize),
		EventsPublished:     make(chan *LogEvent, bufsize),
		EventsReceived:      make(chan *LogEvent, bufsize),
		EventsDeliver:       make(chan *LogEvent, bufsize),
		EventsDropDuplicate: make(chan *LogEvent, bufsize),
		EventsDropTTL:       make(chan *LogEvent, bufsize),
		EventsRouteDirect:   make(chan *LogEvent, bufsize),
		EventsRouteWell:     make(chan *LogEvent, bufsize),
		EventsRouteRandom:   make(chan *LogEvent, bufsize),
	}
}

func (l *Logger) updateSent(n *Node, i uint32, f []byte, t *pubkey) {
	if l != nil && l.UpdatesSent != nil {
		var id *pubkey
		if n != nil {
			idv := n.net.id()
			id = &idv
		}
		u := &peerUpdate{peer: id, index: i, filter: f}
		l.UpdatesSent <- &LogUpdate{
			node: n, entry: u, target: t,
		}
	}
}

func (l *Logger) updateReceived(n *Node, u *peerUpdate) {
	if l != nil && l.UpdatesReceived != nil {
		l.UpdatesReceived <- &LogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *Logger) updateSuccess(n *Node, u *peerUpdate) {
	if l != nil && l.UpdatesSuccess != nil {
		l.UpdatesSuccess <- &LogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *Logger) updateFail(n *Node, u *peerUpdate) {
	if l != nil && l.UpdatesFail != nil {
		l.UpdatesFail <- &LogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *Logger) eventPublished(n *Node, e *event) {
	if l != nil && l.EventsPublished != nil {
		l.EventsPublished <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventReceived(n *Node, e *event) {
	if l != nil && l.EventsReceived != nil {
		l.EventsReceived <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventDeliver(n *Node, e *event) {
	if l != nil && l.EventsDeliver != nil {
		l.EventsDeliver <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventDropDuplicate(n *Node, e *event) {
	if l != nil && l.EventsDropDuplicate != nil {
		l.EventsDropDuplicate <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventDropTTL(n *Node, e *event) {
	if l != nil && l.EventsDropTTL != nil {
		l.EventsDropTTL <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventRouteDirect(n *Node, e *event, t *pubkey) {
	if l != nil && l.EventsRouteDirect != nil {
		l.EventsRouteDirect <- &LogEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *Logger) eventRouteWell(n *Node, e *event, t *pubkey) {
	if l != nil && l.EventsRouteWell != nil {
		l.EventsRouteWell <- &LogEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *Logger) eventRouteRandom(n *Node, e *event, t *pubkey) {
	if l != nil && l.EventsRouteRandom != nil {
		l.EventsRouteRandom <- &LogEvent{
			node: n, entry: e, target: t,
		}
	}
}

func printLogUpdate(prefix string, src string, lu *LogUpdate) {
	fmt.Printf("%s: %s\n", prefix, src)
	// TODO log more info
}

func printLogEvent(prefix string, src string, le *LogEvent) {
	fmt.Printf("%s: %s\n", prefix, src)
	// TODO log more info
}

func LogToConsole(prefix string, stopLogging chan bool) *Logger {
	mustNotBeNil(stopLogging)
	l := NewLogger(10)
	go func() {
		for {
			select {
			case lu := <-l.UpdatesSent:
				go printLogUpdate(prefix, "UpdatesSent", lu)
			case lu := <-l.UpdatesReceived:
				go printLogUpdate(prefix, "UpdatesReceived", lu)
			case lu := <-l.UpdatesSuccess:
				go printLogUpdate(prefix, "UpdatesSuccess", lu)
			case lu := <-l.UpdatesFail:
				go printLogUpdate(prefix, "UpdatesFail", lu)
			case le := <-l.EventsPublished:
				go printLogEvent(prefix, "EventsPublished", le)
			case le := <-l.EventsReceived:
				go printLogEvent(prefix, "EventsReceived", le)
			case le := <-l.EventsDeliver:
				go printLogEvent(prefix, "EventsDeliver", le)
			case le := <-l.EventsDropDuplicate:
				go printLogEvent(prefix, "EventsDropDuplicate", le)
			case le := <-l.EventsDropTTL:
				go printLogEvent(prefix, "EventsDropTTL", le)
			case le := <-l.EventsRouteDirect:
				go printLogEvent(prefix, "EventsRouteDirect", le)
			case le := <-l.EventsRouteWell:
				go printLogEvent(prefix, "EventsRouteWell", le)
			case le := <-l.EventsRouteRandom:
				go printLogEvent(prefix, "EventsRouteRandom", le)
			case <-stopLogging:
				go fmt.Println("Stop logging received!")
				return
			}
			// time.Sleep(time.Millisecond)
		}
	}()
	return l
}
