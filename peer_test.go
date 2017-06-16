package quasar

import "testing"

func TestUpdate(t *testing.T) {

	var pk pubkey
	cfg := &Config{
		DefaultEventTTL:  1024,
		FilterFreshness:  180,
		PropagationDelay: 60000,
		HistoryLimit:     65536,
		HistoryAccuracy:  0.000001,
		FiltersDepth:     1024,
		FiltersM:         8192,
		FiltersK:         6,
	}

	// valid update
	u := &peerUpdate{
		peer:   &pk,
		index:  0,
		filter: make([]byte, 8192/8, 8192/8),
	}
	if !validUpdate(u, cfg) {
		t.Errorf("Expected valid update!")
	}

	// nil update
	if validUpdate(nil, cfg) {
		t.Errorf("Expected valid update!")
	}

	// nil peer
	u = &peerUpdate{
		peer:   nil,
		index:  0,
		filter: make([]byte, 8192/8, 8192/8),
	}
	if validUpdate(u, cfg) {
		t.Errorf("Expected valid update!")
	}

	// index to large
	u = &peerUpdate{
		peer:   &pk,
		index:  1024,
		filter: make([]byte, 8192/8, 8192/8),
	}
	if validUpdate(u, cfg) {
		t.Errorf("Expected valid update!")
	}

	// incorrect filter len
	u = &peerUpdate{
		peer:   &pk,
		index:  1,
		filter: make([]byte, 1024/8, 1024/8),
	}
	if validUpdate(u, cfg) {
		t.Errorf("Expected valid update!")
	}
}
