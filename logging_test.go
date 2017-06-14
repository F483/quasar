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

func TestLogging(t *testing.T) {

	l := &logger{
		updatesSent:         make(chan *updateLog),
		updatesReceived:     make(chan *updateLog),
		updatesSuccess:      make(chan *updateLog),
		updatesFail:         make(chan *updateLog),
		eventsPublished:     make(chan *eventLog),
		eventsReceived:      make(chan *eventLog),
		eventsDeliver:       make(chan *eventLog),
		eventsDropDuplicate: make(chan *eventLog),
		eventsDropTTL:       make(chan *eventLog),
		eventsRouteDirect:   make(chan *eventLog),
		eventsRouteWell:     make(chan *eventLog),
		eventsRouteRandom:   make(chan *eventLog),
	}
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
			if rUpdatesSent == true {
				t.Errorf("Extra updatesSent log!")
			}
			rUpdatesSent = true
		case <-l.updatesReceived:
			if rUpdatesReceived == true {
				t.Errorf("Extra updatesReceived log!")
			}
			rUpdatesReceived = true
		case <-l.updatesSuccess:
			if rUpdatesSuccess == true {
				t.Errorf("Extra updatesSuccess log!")
			}
			rUpdatesSuccess = true
		case <-l.updatesFail:
			if rUpdatesFail == true {
				t.Errorf("Extra updatesFail log!")
			}
			rUpdatesFail = true
		case <-l.eventsPublished:
			if rEventsPublished == true {
				t.Errorf("Extra eventsPublished log!")
			}
			rEventsPublished = true
		case <-l.eventsReceived:
			if rEventsReceived == true {
				t.Errorf("Extra eventReceived log!")
			}
			rEventsReceived = true
		case <-l.eventsDeliver:
			if rEventsDeliver == true {
				t.Errorf("Extra eventsDeliver log!")
			}
			rEventsDeliver = true
		case <-l.eventsDropDuplicate:
			if rEventsDropDuplicate == true {
				t.Errorf("Extra eventsDropDuplicate log!")
			}
			rEventsDropDuplicate = true
		case <-l.eventsDropTTL:
			if rEventsDropTTL == true {
				t.Errorf("Extra eventsDropTTL log!")
			}
			rEventsDropTTL = true
		case <-l.eventsRouteDirect:
			if rEventsRouteDirect == true {
				t.Errorf("Extra eventsRouteDirect log!")
			}
			rEventsRouteDirect = true
		case <-l.eventsRouteWell:
			if rEventsRouteWell == true {
				t.Errorf("Extra eventsRouteWell log!")
			}
			rEventsRouteWell = true
		case <-l.eventsRouteRandom:
			if rEventsRouteRandom == true {
				t.Errorf("Extra eventsRouteRandom log!")
			}
			rEventsRouteRandom = true
		case <-allSent:
			exitLoop = true
		}
	}

	if !(rUpdatesSent && rUpdatesReceived && rUpdatesSuccess &&
		rUpdatesFail && rEventsPublished && rEventsReceived &&
		rEventsDeliver && rEventsDropDuplicate && rEventsDropTTL &&
		rEventsRouteDirect && rEventsRouteWell && rEventsRouteRandom) {
		t.Errorf("Missing log event!")
	}
}
