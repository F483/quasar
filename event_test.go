package quasar

import "testing"

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
		publishers:  []pubkey{},
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
		publishers:  []pubkey{},
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
		publishers:  []pubkey{},
		ttl:         42,
	}
	if validEvent(e) {
		t.Errorf("Expected invalid event!")
	}
}
