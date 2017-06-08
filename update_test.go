package quasar

import "testing"

func TestUpdate(t *testing.T) {

	var pk pubkey
	cfg := &config{
		defaultEventTTL:     1024,
		dispatcherDelay:     1,
		filterFreshness:     180,
		propagationInterval: 60,
		historyLimit:        65536,
		historyAccuracy:     0.000001,
		filtersDepth:        1024,
		filtersM:            8192,
		filtersK:            6,
	}

	// valid update
	u := &update{
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
	u = &update{
		peer:   nil,
		index:  0,
		filter: make([]byte, 8192/8, 8192/8),
	}
	if validUpdate(u, cfg) {
		t.Errorf("Expected valid update!")
	}

	// index to large
	u = &update{
		peer:   &pk,
		index:  1024,
		filter: make([]byte, 8192/8, 8192/8),
	}
	if validUpdate(u, cfg) {
		t.Errorf("Expected valid update!")
	}

	// incorrect filter len
	u = &update{
		peer:   &pk,
		index:  1,
		filter: make([]byte, 1024/8, 1024/8),
	}
	if validUpdate(u, cfg) {
		t.Errorf("Expected valid update!")
	}
}
