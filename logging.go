package quasar

type updateLog struct {
	node   *pubkey
	entry  *peerUpdate
	target *pubkey
}

type eventLog struct {
	node   *pubkey
	entry  *event
	target *pubkey
}

type logger struct {
	updatesSent         chan *updateLog
	updatesReceived     chan *updateLog
	updatesSuccess      chan *updateLog // (added to own filters)
	updatesFail         chan *updateLog // (not from neighbour)
	eventsPublished     chan *eventLog
	eventsReceived      chan *eventLog
	eventsDeliver       chan *eventLog
	eventsDropDuplicate chan *eventLog
	eventsDropTTL       chan *eventLog
	eventsRouteDirect   chan *eventLog
	eventsRouteWell     chan *eventLog
	eventsRouteRandom   chan *eventLog
}

func (l *logger) updateSent(n *pubkey, i uint32, f []byte, t *pubkey) {
	if l != nil {
		u := &peerUpdate{peer: n, index: i, filter: f}
		l.updatesSent <- &updateLog{node: n, entry: u, target: t}
	}
}

func (l *logger) updateReceived(n *pubkey, u *peerUpdate) {
	if l != nil {
		l.updatesReceived <- &updateLog{
			node: n, entry: u, target: nil,
		}
	}
}

func (l *logger) updateSuccess(n *pubkey, u *peerUpdate) {
	if l != nil {
		l.updatesSuccess <- &updateLog{node: n, entry: u, target: nil}
	}
}

func (l *logger) updateFail(n *pubkey, u *peerUpdate) {
	if l != nil {
		l.updatesFail <- &updateLog{node: n, entry: u, target: nil}
	}
}

func (l *logger) eventPublished(n *pubkey, e *event) {
	if l != nil {
		l.eventsPublished <- &eventLog{node: n, entry: e, target: nil}
	}
}

func (l *logger) eventReceived(n *pubkey, e *event) {
	if l != nil {
		l.eventsReceived <- &eventLog{node: n, entry: e, target: nil}
	}
}

func (l *logger) eventDeliver(n *pubkey, e *event) {
	if l != nil {
		l.eventsDeliver <- &eventLog{node: n, entry: e, target: nil}
	}
}

func (l *logger) eventDropDuplicate(n *pubkey, e *event) {
	if l != nil {
		l.eventsDropDuplicate <- &eventLog{node: n, entry: e, target: nil}
	}
}

func (l *logger) eventDropTTL(n *pubkey, e *event) {
	if l != nil {
		l.eventsDropTTL <- &eventLog{node: n, entry: e, target: nil}
	}
}

func (l *logger) eventRouteDirect(n *pubkey, e *event, t *pubkey) {
	if l != nil {
		l.eventsRouteDirect <- &eventLog{node: n, entry: e, target: t}
	}
}

func (l *logger) eventRouteWell(n *pubkey, e *event, t *pubkey) {
	if l != nil {
		l.eventsRouteWell <- &eventLog{node: n, entry: e, target: t}
	}
}

func (l *logger) eventRouteRandom(n *pubkey, e *event, t *pubkey) {
	if l != nil {
		l.eventsRouteRandom <- &eventLog{node: n, entry: e, target: t}
	}
}
