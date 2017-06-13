package quasar

import "testing"

func TestFilters(t *testing.T) {
	cfg := config{
		defaultEventTTL:  1024,
		filterFreshness:  180,
		propagationDelay: 60000,
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

	a := filterInsert(filters[0], cfg, []byte("foo"))
	if filterContains(a, cfg, []byte("foo")) == false {
		t.Errorf("Filter doesnt contain added value!")
	}
	if filterContains(a, cfg, []byte("bar")) != false {
		t.Errorf("Filter contains unexpecded value!")
	}
	if filterContains(a, cfg, []byte("baz")) != false {
		t.Errorf("Filter contains unexpecded value!")
	}

	b := filterInsert(filters[1], cfg, []byte("bar"))
	if filterContains(b, cfg, []byte("bar")) == false {
		t.Errorf("Filter doesnt contain added value!")
	}
	if filterContains(b, cfg, []byte("foo")) != false {
		t.Errorf("Filter contains unexpecded value!")
	}
	if filterContains(b, cfg, []byte("baz")) != false {
		t.Errorf("Filter contains unexpecded value!")
	}

	c := filterInsert(filters[2], cfg, []byte("baz"))

	m := mergeFilters(a, b, c)
	if filterContains(m, cfg, []byte("foo")) == false {
		t.Errorf("Merged filter missing expected value!")
	}
	if filterContains(m, cfg, []byte("bar")) == false {
		t.Errorf("Merged filter missing expected value!")
	}
	if filterContains(m, cfg, []byte("baz")) == false {
		t.Errorf("Merged filter missing expected value!")
	}
	if filterContains(m, cfg, []byte("bam")) != false {
		t.Errorf("Merged filter contains unexpecded value!")
	}

	zf := mergeFilters()
	if zf != nil {
		t.Errorf("Expected nil result for merging no filters!")
	}

	z := mergeFilters(filters...)
	if z == nil {
		t.Errorf("Expected non nil result for merging all filters!")
	}
}

func TestFiltersVariadic(t *testing.T) {
	cfg := config{
		defaultEventTTL:  1024,
		filterFreshness:  180,
		propagationDelay: 60000,
		historyLimit:     65536,
		historyAccuracy:  0.000001,
		filtersDepth:     1024,
		filtersM:         8192, // m 1k
		filtersK:         6,    // hashes
	}

	filters := newFilters(cfg)

	a := filterInsert(filters[0], cfg, []byte("foo"), []byte("bar"))

	if filterContains(a, cfg, []byte("foo")) == false {
		t.Errorf("Filter doesnt contain added value!")
	}
	if filterContains(a, cfg, []byte("bar")) == false {
		t.Errorf("Filter contains unexpecded value!")
	}
	if filterContains(a, cfg, []byte("baz")) != false {
		t.Errorf("Filter contains unexpecded value!")
	}
}
