package quasar

import "testing"

func TestNewEvent(t *testing.T) {
	ttl := uint32(42)
	event := newEvent([]byte("test topic"), []byte("test message"), ttl)
	if event == nil {
		t.Errorf("Expected event!")
	}
	if string(event.message) != "test message" {
		t.Errorf("event message not set!")
	}
	if event.ttl != ttl {
		t.Errorf("event ttl not set!")
	}
	if event.publishers == nil && len(event.publishers) != 0 {
		t.Errorf("event publishers not set!")
	}
	if event.topicDigest == nil {
		t.Errorf("event digest not set!")
	}
}

func TestEvent(t *testing.T) {

	// new always creates valid event
	if !validEvent(newEvent([]byte("foo"), []byte("bar"), 1)) {
		t.Errorf("Expected valid event!")
	}
	if !validEvent(newEvent(nil, nil, 0)) {
		t.Errorf("Expected valid event!")
	}

	// valid event
	digest := hash160([]byte("foo"))
	e := &event{
		topicDigest: &digest,
		message:     []byte("bar"),
		publishers:  []*hash160digest{},
		ttl:         42,
	}
	if !validEvent(e) {
		t.Errorf("Expected invalid event!")
	}

	// nil event
	if validEvent(nil) {
		t.Errorf("Expected invalid event!")
	}

	// message nil
	e = &event{
		topicDigest: &digest,
		message:     nil,
		publishers:  []*hash160digest{},
		ttl:         42,
	}
	if validEvent(e) {
		t.Errorf("Expected invalid event!")
	}

	// publishers nil
	e = &event{
		topicDigest: &digest,
		message:     []byte("bar"),
		publishers:  nil,
		ttl:         42,
	}
	if validEvent(e) {
		t.Errorf("Expected invalid event!")
	}

	// topic digest nil
	e = &event{
		topicDigest: nil,
		message:     []byte("bar"),
		publishers:  []*hash160digest{},
		ttl:         42,
	}
	if validEvent(e) {
		t.Errorf("Expected invalid event!")
	}
}
