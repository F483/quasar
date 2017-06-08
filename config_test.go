package quasar

import "testing"

func TestConfig(t *testing.T) {

	// valid config
	valid := validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     180,
		PropagationInterval: 60,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.000001,
		FiltersDepth:        1024,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if !valid {
		t.Errorf("Expected valid config!")
	}

	// nil config
	valid = validConfig(nil)
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid event ttl
	valid = validConfig(&Config{
		DefaultEventTTL:     0,
		DispatcherDelay:     1,
		FilterFreshness:     180,
		PropagationInterval: 60,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.000001,
		FiltersDepth:        1024,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid dispatcher delay
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     0,
		FilterFreshness:     180,
		PropagationInterval: 60,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.000001,
		FiltersDepth:        1024,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid freshness
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     0,
		PropagationInterval: 60,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.000001,
		FiltersDepth:        1024,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid propagation interval
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     1,
		PropagationInterval: 0,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.000001,
		FiltersDepth:        1024,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid history accuracy
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     1,
		PropagationInterval: 0,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.0,
		FiltersDepth:        1024,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid history accuracy
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     1,
		PropagationInterval: 0,
		HistoryLimit:        65536,
		HistoryAccuracy:     1.0,
		FiltersDepth:        1024,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid filter depth
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     1,
		PropagationInterval: 0,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.1,
		FiltersDepth:        0,
		FiltersM:            8192,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid filter m
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     1,
		PropagationInterval: 0,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.1,
		FiltersDepth:        1,
		FiltersM:            0,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid filter m
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     1,
		PropagationInterval: 0,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.1,
		FiltersDepth:        1,
		FiltersM:            13,
		FiltersK:            6,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}

	// invalid filter k
	valid = validConfig(&Config{
		DefaultEventTTL:     1024,
		DispatcherDelay:     1,
		FilterFreshness:     1,
		PropagationInterval: 0,
		HistoryLimit:        65536,
		HistoryAccuracy:     0.1,
		FiltersDepth:        1,
		FiltersM:            8192,
		FiltersK:            0,
	})
	if valid {
		t.Errorf("Expected invalid config!")
	}
}
