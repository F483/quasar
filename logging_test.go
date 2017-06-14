package quasar

import (
	"testing"
	"time"
)

func sendLogEntries(l *logger, allSent chan bool) {
	time.Sleep(time.Millisecond * time.Duration(100))
	l.updateSent(nil, 0, nil, nil)
	l.updateReceived(nil, nil)
	l.updateSuccess(nil, nil)
	l.updateFail(nil, nil)
	l.eventPublished(nil, nil)
	l.eventReceived(nil, nil)
	l.eventDeliver(nil, nil)
	l.eventDropDuplicate(nil, nil)
	l.eventDropTTL(nil, nil)
	l.eventRouteDirect(nil, nil, nil)
	l.eventRouteWell(nil, nil, nil)
	l.eventRouteRandom(nil, nil, nil)
	time.Sleep(time.Millisecond * time.Duration(100))
	allSent <- true
}

func setReceived(b *bool, t *testing.T) {
	if *b == true {
		t.Errorf("Unexpected extra log!")
	}
	*b = true
}

func TestLogging(t *testing.T) {

	l := newLogger()
	allSent := make(chan bool)
	go sendLogEntries(l, allSent)

	rUpdatesSent := false
	rUpdatesReceived := false
	rUpdatesSuccess := false
	rUpdatesFail := false
	rEventsPublished := false
	rEventsReceived := false
	rEventsDeliver := false
	rEventsDropDuplicate := false
	rEventsDropTTL := false
	rEventsRouteDirect := false
	rEventsRouteWell := false
	rEventsRouteRandom := false

	exitLoop := false
	for !exitLoop {
		select {
		case <-l.updatesSent:
			setReceived(&rUpdatesSent, t)
		case <-l.updatesReceived:
			setReceived(&rUpdatesReceived, t)
		case <-l.updatesSuccess:
			setReceived(&rUpdatesSuccess, t)
		case <-l.updatesFail:
			setReceived(&rUpdatesFail, t)
		case <-l.eventsPublished:
			setReceived(&rEventsPublished, t)
		case <-l.eventsReceived:
			setReceived(&rEventsReceived, t)
		case <-l.eventsDeliver:
			setReceived(&rEventsDeliver, t)
		case <-l.eventsDropDuplicate:
			setReceived(&rEventsDropDuplicate, t)
		case <-l.eventsDropTTL:
			setReceived(&rEventsDropTTL, t)
		case <-l.eventsRouteDirect:
			setReceived(&rEventsRouteDirect, t)
		case <-l.eventsRouteWell:
			setReceived(&rEventsRouteWell, t)
		case <-l.eventsRouteRandom:
			setReceived(&rEventsRouteRandom, t)
		case <-allSent:
			exitLoop = true
		}
	}

	if !(rUpdatesSent && rUpdatesReceived &&
		rUpdatesSuccess && rUpdatesFail &&
		rEventsPublished && rEventsReceived &&
		rEventsDeliver && rEventsDropDuplicate &&
		rEventsDropTTL && rEventsRouteDirect &&
		rEventsRouteWell && rEventsRouteRandom) {
		t.Errorf("Missing log event!")
	}
}
