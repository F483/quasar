package quasar

import "testing"

func TestFilters(t *testing.T) {
	cfg := config{
		defaultEventTTL:  1024,
		filterFreshness:  180,
		propagationDelay: 60,
		historyLimit:     65536,
		historyAccuracy:  0.000001,
		filtersDepth:     1024,
		filtersM:         8192, // m 1k
		filtersK:         6,    // hashes
	}

	filters := newFilters(cfg)
	if uint32(len(filters)) != cfg.filtersDepth {
		t.Errorf("Incorrect filter depth!")
	}

	a := filterInsert(filters[0], []byte("foo"), cfg)
	if filterContains(a, []byte("foo"), cfg) == false {
		t.Errorf("Filter doesnt contain added value!")
	}
	if filterContains(a, []byte("bar"), cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}
	if filterContains(a, []byte("baz"), cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}

	b := filterInsert(filters[1], []byte("bar"), cfg)
	if filterContains(b, []byte("bar"), cfg) == false {
		t.Errorf("Filter doesnt contain added value!")
	}
	if filterContains(b, []byte("foo"), cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}
	if filterContains(b, []byte("baz"), cfg) != false {
		t.Errorf("Filter contains unexpecded value!")
	}

	c := mergeFilters(a, b)
	if filterContains(c, []byte("foo"), cfg) == false {
		t.Errorf("Merged filter missing expected value!")
	}
	if filterContains(c, []byte("bar"), cfg) == false {
		t.Errorf("Merged filter missing expected value!")
	}
	if filterContains(c, []byte("baz"), cfg) != false {
		t.Errorf("Merged filter contains unexpecded value!")
	}
}
