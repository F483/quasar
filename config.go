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
	FiltersM            uint64        // filter size in bits
	FiltersK            uint64        // number of hashes
}

func validConfig(c *Config) bool {
	if c.FiltersM%64 != 0 {
		return false
	}
	// FIXME review types and check everything
	return true
}
