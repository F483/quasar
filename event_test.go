package quasar

import "testing"

func TestNewEvent(t *testing.T) {
	ttl := uint32(42)
	event := newEvent([]byte("test topic"), []byte("test message"), ttl)
	if event == nil {
		t.Errorf("Expected event!")
	}
	if string(event.Message) != "test message" {
		t.Errorf("event message not set!")
	}
	if event.Ttl != ttl {
		t.Errorf("event ttl not set!")
	}
	if event.Publishers == nil && len(event.Publishers) != 0 {
		t.Errorf("event publishers not set!")
	}
	if event.TopicDigest == nil {
		t.Errorf("event digest not set!")
	}
}

func TestEvent(t *testing.T) {

	// new always creates valid event
	e := newEvent([]byte("foo"), []byte("bar"), 1)
	if !e.valid() {
		t.Errorf("Expected valid event!")
	}
	e = newEvent(nil, nil, 0)
	if !e.valid() {
		t.Errorf("Expected valid event!")
	}

	// valid event
	digest := sha256sum([]byte("foo"))
	e = &event{
		TopicDigest: &digest,
		Message:     []byte("bar"),
		Publishers:  []*sha256digest{},
		Ttl:         42,
	}
	if !e.valid() {
		t.Errorf("Expected invalid event!")
	}

	// nil event
	e = nil
	if e.valid() {
		t.Errorf("Expected invalid event!")
	}

	// message nil
	e = &event{
		TopicDigest: &digest,
		Message:     nil,
		Publishers:  []*sha256digest{},
		Ttl:         42,
	}
	if e.valid() {
		t.Errorf("Expected invalid event!")
	}

	// publishers nil
	e = &event{
		TopicDigest: &digest,
		Message:     []byte("bar"),
		Publishers:  nil,
		Ttl:         42,
	}
	if e.valid() {
		t.Errorf("Expected invalid event!")
	}

	// topic digest nil
	e = &event{
		TopicDigest: nil,
		Message:     []byte("bar"),
		Publishers:  []*sha256digest{},
		Ttl:         42,
	}
	if e.valid() {
		t.Errorf("Expected invalid event!")
	}
}
