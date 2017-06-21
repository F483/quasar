package quasar

import "testing"

func TestFiltersBulk(t *testing.T) {
	cfg := &Config{
		DefaultEventTTL:  1024,
		FilterFreshness:  180,
		PropagationDelay: 60000,
		HistoryLimit:     65536,
		HistoryAccuracy:  0.000001,
		FiltersDepth:     1024,
		FiltersM:         8192, // m 1k
		FiltersK:         6,    // hashes
	}

	filters := newFilters(cfg)
	if uint32(len(filters)) != cfg.FiltersDepth {
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
	cfg := &Config{
		DefaultEventTTL:  1024,
		FilterFreshness:  180,
		PropagationDelay: 60000,
		HistoryLimit:     65536,
		HistoryAccuracy:  0.000001,
		FiltersDepth:     1024,
		FiltersM:         8192, // m 1k
		FiltersK:         6,    // hashes
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
