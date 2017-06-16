package quasar

// QuasarLogUpdate used for monitoring internal filter updates.
type QuasarLogUpdate struct {
	node   *Quasar
	entry  *peerUpdate
	target *pubkey
}

// QuasarLogEvent used for monitoring internal events.
type QuasarLogEvent struct {
	node   *Quasar
	entry  *event
	target *pubkey
}

// QuasarLog provides a logger used by Quasar nodes for logging internals.
type QuasarLog struct {
	UpdatesSent         chan *QuasarLogUpdate
	UpdatesReceived     chan *QuasarLogUpdate
	UpdatesSuccess      chan *QuasarLogUpdate // added to filters
	UpdatesFail         chan *QuasarLogUpdate // not from neighbour
	EventsPublished     chan *QuasarLogEvent
	EventsReceived      chan *QuasarLogEvent
	EventsDeliver       chan *QuasarLogEvent
	EventsDropDuplicate chan *QuasarLogEvent
	EventsDropTTL       chan *QuasarLogEvent
	EventsRouteDirect   chan *QuasarLogEvent
	EventsRouteWell     chan *QuasarLogEvent
	EventsRouteRandom   chan *QuasarLogEvent
}

// NewQuasarLog creats a new default logger instance.
func NewQuasarLog() *QuasarLog {
	return &QuasarLog{
		UpdatesSent:         make(chan *QuasarLogUpdate),
		UpdatesReceived:     make(chan *QuasarLogUpdate),
		UpdatesSuccess:      make(chan *QuasarLogUpdate),
		UpdatesFail:         make(chan *QuasarLogUpdate),
		EventsPublished:     make(chan *QuasarLogEvent),
		EventsReceived:      make(chan *QuasarLogEvent),
		EventsDeliver:       make(chan *QuasarLogEvent),
		EventsDropDuplicate: make(chan *QuasarLogEvent),
		EventsDropTTL:       make(chan *QuasarLogEvent),
		EventsRouteDirect:   make(chan *QuasarLogEvent),
		EventsRouteWell:     make(chan *QuasarLogEvent),
		EventsRouteRandom:   make(chan *QuasarLogEvent),
	}
}

func (l *QuasarLog) updateSent(n *Quasar, i uint32, f []byte, t *pubkey) {
	if l != nil && l.UpdatesSent != nil {
		var id *pubkey = nil
		if n != nil {
			idv := n.net.Id()
			id = &idv
		}
		u := &peerUpdate{peer: id, index: i, filter: f}
		l.UpdatesSent <- &QuasarLogUpdate{
			node: n, entry: u, target: t,
		}
	}
}

func (l *QuasarLog) updateReceived(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesReceived != nil {
		l.UpdatesReceived <- &QuasarLogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *QuasarLog) updateSuccess(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesSuccess != nil {
		l.UpdatesSuccess <- &QuasarLogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *QuasarLog) updateFail(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesFail != nil {
		l.UpdatesFail <- &QuasarLogUpdate{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *QuasarLog) eventPublished(n *Quasar, e *event) {
	if l != nil && l.EventsPublished != nil {
		l.EventsPublished <- &QuasarLogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventReceived(n *Quasar, e *event) {
	if l != nil && l.EventsReceived != nil {
		l.EventsReceived <- &QuasarLogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventDeliver(n *Quasar, e *event) {
	if l != nil && l.EventsDeliver != nil {
		l.EventsDeliver <- &QuasarLogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventDropDuplicate(n *Quasar, e *event) {
	if l != nil && l.EventsDropDuplicate != nil {
		l.EventsDropDuplicate <- &QuasarLogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventDropTTL(n *Quasar, e *event) {
	if l != nil && l.EventsDropTTL != nil {
		l.EventsDropTTL <- &QuasarLogEvent{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventRouteDirect(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteDirect != nil {
		l.EventsRouteDirect <- &QuasarLogEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *QuasarLog) eventRouteWell(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteWell != nil {
		l.EventsRouteWell <- &QuasarLogEvent{
			node: n, entry: e, target: t,
		}
	}
}

func (l *QuasarLog) eventRouteRandom(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteRandom != nil {
		l.EventsRouteRandom <- &QuasarLogEvent{
			node: n, entry: e, target: t,
		}
	}
}
