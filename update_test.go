package quasar

import "testing"

func TestUpdate(t *testing.T) {

	var pk pubkey
	depth := uint32(8)
	cfg := &Config{
		DefaultEventTTL:  1024,
		FilterFreshness:  180,
		PropagationDelay: 60000,
		HistoryLimit:     65536,
		HistoryAccuracy:  0.000001,
		FiltersDepth:     depth,
		FiltersM:         8192,
		FiltersK:         6,
	}

	// valid update
	filters := newFilters(cfg)
	u := &update{
		NodeId:  &pk,
		Filters: filters,
	}
	if !u.valid(cfg) {
		t.Errorf("Expected valid update!")
	}

	// nil update
	u = nil
	if u.valid(cfg) {
		t.Errorf("Expected valid update!")
	}

	// nil peer
	u = &update{
		NodeId:  nil,
		Filters: filters,
	}
	if u.valid(cfg) {
		t.Errorf("Expected valid update!")
	}

	// incorrect filter len
	u = &update{
		NodeId:  &pk,
		Filters: filters[:depth-1],
	}
	if u.valid(cfg) {
		t.Errorf("Expected valid update!")
	}
}
