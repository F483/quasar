package quasar

// LogUpdate used for monitoring internal filter updates.
type LogUpdate struct {
	node   *Quasar
	entry  *peerUpdate
	target *pubkey
}

// LogEvent used for monitoring internal events.
type LogEvent struct {
	node   *Quasar
	entry  *event
	target *pubkey
}

// Logger provides a logger used by Quasar nodes for logging internals.
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
}

// NewLogger creats a new default logger instance.
func NewLogger() *Logger {
	return &Logger{
		UpdatesSent:         make(chan *LogUpdate),
		UpdatesReceived:     make(chan *LogUpdate),
		UpdatesSuccess:      make(chan *LogUpdate),
		UpdatesFail:         make(chan *LogUpdate),
		EventsPublished:     make(chan *LogEvent),
		EventsReceived:      make(chan *LogEvent),
		EventsDeliver:       make(chan *LogEvent),
		EventsDropDuplicate: make(chan *LogEvent),
		EventsDropTTL:       make(chan *LogEvent),
		EventsRouteDirect:   make(chan *LogEvent),
		EventsRouteWell:     make(chan *LogEvent),
		EventsRouteRandom:   make(chan *LogEvent),
	}
}

func (l *Logger) updateSent(n *Quasar, i uint32, f []byte, t *pubkey) {
	if l != nil && l.UpdatesSent != nil {
		var id *pubkey
		if n != nil {
			idv := n.net.Id()
			id = &idv
		}
		u := &peerUpdate{peer: id, index: i, filter: f}
		l.UpdatesSent <- &LogUpdate{
			node: n, entry: u, target: t,
		}
	}
}

func (l *Logger) updateReceived(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesReceived != nil {
		l.UpdatesReceived <- &LogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *Logger) updateSuccess(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesSuccess != nil {
		l.UpdatesSuccess <- &LogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *Logger) updateFail(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesFail != nil {
		l.UpdatesFail <- &LogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *Logger) eventPublished(n *Quasar, e *event) {
	if l != nil && l.EventsPublished != nil {
		l.EventsPublished <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventReceived(n *Quasar, e *event) {
	if l != nil && l.EventsReceived != nil {
		l.EventsReceived <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventDeliver(n *Quasar, e *event) {
	if l != nil && l.EventsDeliver != nil {
		l.EventsDeliver <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventDropDuplicate(n *Quasar, e *event) {
	if l != nil && l.EventsDropDuplicate != nil {
		l.EventsDropDuplicate <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventDropTTL(n *Quasar, e *event) {
	if l != nil && l.EventsDropTTL != nil {
		l.EventsDropTTL <- &LogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *Logger) eventRouteDirect(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteDirect != nil {
		l.EventsRouteDirect <- &LogEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *Logger) eventRouteWell(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteWell != nil {
		l.EventsRouteWell <- &LogEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *Logger) eventRouteRandom(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteRandom != nil {
		l.EventsRouteRandom <- &LogEvent{
			node: n, entry: e, target: t,
		}
	}
}
