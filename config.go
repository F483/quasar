package quasar

import "time"

type Config struct {
	DefaultEventTTL     uint32        // decremented every hop
	DispatcherDelay     time.Duration // in ms FIXME uint64
	FilterFreshness     uint64        // in seconds
	PropagationInterval uint64        // in seconds
	HistoryLimit        uint32        // entries remembered
	HistoryAccuracy     float64       // chance of error
	FiltersDepth        uint32        // filter stack height
	FiltersM            uint64        // size in bits (multiple of 64)
	FiltersK            uint64        // number of hashes
}

func validConfig(c *Config) bool {
	return c != nil && c.DefaultEventTTL != 0 &&
		c.DispatcherDelay != 0 && c.FilterFreshness != 0 &&
		c.PropagationInterval != 0 && c.HistoryAccuracy > 0.0 &&
		c.HistoryAccuracy < 1.0 && c.FiltersDepth != 0 &&
		c.FiltersM != 0 && c.FiltersM%64 == 0 && c.FiltersK != 0
}
