package quasar

type QuasarUpdateLog struct {
	node   *Quasar
	entry  *peerUpdate
	target *pubkey
}

type QuasarEventLog struct {
	node   *Quasar
	entry  *event
	target *pubkey
}

type QuasarLog struct {
	UpdatesSent         chan *QuasarUpdateLog
	UpdatesReceived     chan *QuasarUpdateLog
	UpdatesSuccess      chan *QuasarUpdateLog // (added to filters)
	UpdatesFail         chan *QuasarUpdateLog // (not from neighbour)
	EventsPublished     chan *QuasarEventLog
	EventsReceived      chan *QuasarEventLog
	EventsDeliver       chan *QuasarEventLog
	EventsDropDuplicate chan *QuasarEventLog
	EventsDropTTL       chan *QuasarEventLog
	EventsRouteDirect   chan *QuasarEventLog
	EventsRouteWell     chan *QuasarEventLog
	EventsRouteRandom   chan *QuasarEventLog
}

func NewQuasarLog() *QuasarLog {
	return &QuasarLog{
		UpdatesSent:         make(chan *QuasarUpdateLog),
		UpdatesReceived:     make(chan *QuasarUpdateLog),
		UpdatesSuccess:      make(chan *QuasarUpdateLog),
		UpdatesFail:         make(chan *QuasarUpdateLog),
		EventsPublished:     make(chan *QuasarEventLog),
		EventsReceived:      make(chan *QuasarEventLog),
		EventsDeliver:       make(chan *QuasarEventLog),
		EventsDropDuplicate: make(chan *QuasarEventLog),
		EventsDropTTL:       make(chan *QuasarEventLog),
		EventsRouteDirect:   make(chan *QuasarEventLog),
		EventsRouteWell:     make(chan *QuasarEventLog),
		EventsRouteRandom:   make(chan *QuasarEventLog),
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
		l.UpdatesSent <- &QuasarUpdateLog{
			node: n, entry: u, target: t,
		}
	}
}

func (l *QuasarLog) updateReceived(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesReceived != nil {
		l.UpdatesReceived <- &QuasarUpdateLog{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *QuasarLog) updateSuccess(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesSuccess != nil {
		l.UpdatesSuccess <- &QuasarUpdateLog{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *QuasarLog) updateFail(n *Quasar, u *peerUpdate) {
	if l != nil && l.UpdatesFail != nil {
		l.UpdatesFail <- &QuasarUpdateLog{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *QuasarLog) eventPublished(n *Quasar, e *event) {
	if l != nil && l.EventsPublished != nil {
		l.EventsPublished <- &QuasarEventLog{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventReceived(n *Quasar, e *event) {
	if l != nil && l.EventsReceived != nil {
		l.EventsReceived <- &QuasarEventLog{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventDeliver(n *Quasar, e *event) {
	if l != nil && l.EventsDeliver != nil {
		l.EventsDeliver <- &QuasarEventLog{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventDropDuplicate(n *Quasar, e *event) {
	if l != nil && l.EventsDropDuplicate != nil {
		l.EventsDropDuplicate <- &QuasarEventLog{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventDropTTL(n *Quasar, e *event) {
	if l != nil && l.EventsDropTTL != nil {
		l.EventsDropTTL <- &QuasarEventLog{
			node: n, entry: e, target: nil,
		}
	}
}

func (l *QuasarLog) eventRouteDirect(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteDirect != nil {
		l.EventsRouteDirect <- &QuasarEventLog{
			node: n, entry: e, target: t,
		}
	}
}

func (l *QuasarLog) eventRouteWell(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteWell != nil {
		l.EventsRouteWell <- &QuasarEventLog{
			node: n, entry: e, target: t,
		}
	}
}

func (l *QuasarLog) eventRouteRandom(n *Quasar, e *event, t *pubkey) {
	if l != nil && l.EventsRouteRandom != nil {
		l.EventsRouteRandom <- &QuasarEventLog{
			node: n, entry: e, target: t,
		}
	}
}
