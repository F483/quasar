package quasar

import (
	"testing"
	"time"
)

// TODO test nil channels

func sendUpdateLogEntries(l *logger, allSent chan bool) {
	time.Sleep(time.Millisecond * time.Duration(100))
	l.updateSent(nil, nil, nil)
	l.updateReceived(nil, nil)
	l.updateSuccess(nil, nil)
	l.updateFail(nil, nil)
	time.Sleep(time.Millisecond * time.Duration(100))
	allSent <- true
}

func sendEventLogEntries(l *logger, allSent chan bool) {
	time.Sleep(time.Millisecond * time.Duration(100))
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

func TestUpdateLogging(t *testing.T) {

	l := newLogger(10)
	allSent := make(chan bool)
	go sendUpdateLogEntries(l, allSent)

	sent := false
	received := false
	success := false
	fail := false

	exitLoop := false
	for !exitLoop {
		select {
		case <-l.updatesSent:
			setReceived(&sent, t)
		case <-l.updatesReceived:
			setReceived(&received, t)
		case <-l.updatesSuccess:
			setReceived(&success, t)
		case <-l.updatesFail:
			setReceived(&fail, t)
		case <-allSent:
			exitLoop = true
		}
	}

	if !(sent && received && success && fail) {
		t.Errorf("Missing log event!")
	}
}

func collectEventLogs(published *bool, received *bool, deliver *bool,
	dropDuplicate *bool, dropTTL *bool, routeDirect *bool,
	routeWell *bool, routeRandom *bool, allSent chan bool,
	l *logger, t *testing.T) {
	exitLoop := false
	for !exitLoop {
		select {
		case <-l.eventsPublished:
			setReceived(published, t)
		case <-l.eventsReceived:
			setReceived(received, t)
		case <-l.eventsDeliver:
			setReceived(deliver, t)
		case <-l.eventsDropDuplicate:
			setReceived(dropDuplicate, t)
		case <-l.eventsDropTTL:
			setReceived(dropTTL, t)
		case <-l.eventsRouteDirect:
			setReceived(routeDirect, t)
		case <-l.eventsRouteWell:
			setReceived(routeWell, t)
		case <-l.eventsRouteRandom:
			setReceived(routeRandom, t)
		case <-allSent:
			exitLoop = true
		}
	}
}

func TestEventLogging(t *testing.T) {

	l := newLogger(10)
	allSent := make(chan bool)
	go sendEventLogEntries(l, allSent)

	published := false
	received := false
	deliver := false
	dropDuplicate := false
	dropTTL := false
	routeDirect := false
	routeWell := false
	routeRandom := false
	collectEventLogs(&published, &received, &deliver, &dropDuplicate,
		&dropTTL, &routeDirect, &routeWell, &routeRandom, allSent, l, t)
	if !(published && received &&
		deliver && dropDuplicate &&
		dropTTL && routeDirect &&
		routeWell && routeRandom) {
		t.Errorf("Missing log event!")
	}
}
