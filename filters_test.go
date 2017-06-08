package quasar

import (
	"testing"
)

func TestFilters(t *testing.T) {
	cfg := Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     180,
		PropagationInterval: 60,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.000001,
		FiltersDepth:        1024,
		FiltersM:            8192, // m 1k
		FiltersK:            6,    // hashes
	}

	filters := newFilters(&cfg)
	if uint32(len(filters)) != cfg.FiltersDepth {
		t.Errorf("Incorrect filter depth!")
	}

	a := filterInsert(filters[0], []byte("foo"), &cfg)
	if filterContains(a, []byte("foo"), &cfg) == false {
		t.Errorf("Filter doesnt contain added value!")
	}
	if filterContains(a, []byte("bar"), &cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}
	if filterContains(a, []byte("baz"), &cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}

	b := filterInsert(filters[1], []byte("bar"), &cfg)
	if filterContains(b, []byte("bar"), &cfg) == false {
		t.Errorf("Filter doesnt contain added value!")
	}
	if filterContains(b, []byte("foo"), &cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}
	if filterContains(b, []byte("baz"), &cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}

	c := mergeFilters(a, b)
	if filterContains(c, []byte("foo"), &cfg) == false {
		t.Errorf("Merged filter missing expected value!")
	}
	if filterContains(c, []byte("bar"), &cfg) == false {
		t.Errorf("Merged filter missing expected value!")
	}
	if filterContains(c, []byte("baz"), &cfg) != false {
		t.Errorf("Merged filter contains unexpecded value!")
	}
}
