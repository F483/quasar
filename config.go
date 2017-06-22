package quasar

// Config for a quasar node.
type Config struct {
	DefaultEventTTL  uint32  // decremented every hop
	FilterFreshness  uint64  // in ms (older filters ignored)
	PropagationDelay uint64  // in ms (when filters sent to peers)
	HistoryLimit     uint32  // nr of event entries probably remembered
	HistoryAccuracy  float64 // false positive chance between 0.0 and 1.0
	FiltersDepth     uint32  // attenuated bloom filter depth
	FiltersM         uint64  // bloom filter m bits (multiple of 64)
	FiltersK         uint64  // bloom filter k hashes
}

// StandardConfig uses well agreed upon values, only deviate for testing.
var StandardConfig = Config{
	DefaultEventTTL:  32,              // 4 missed wells
	FilterFreshness:  50000,           // 50sec (>2.5 propagation)
	PropagationDelay: 20000,           // 20sec (~0.5M/min const traffic)
	HistoryLimit:     1048576,         // remember last 1G events
	HistoryAccuracy:  1.0 / 1048576.0, // avg 1 false positive per 1G
	FiltersDepth:     8,               // covers many many nodes ...
	FiltersM:         10240,           // keep packet under safe MTU 1400
	FiltersK:         7,               // = (m / n) log(2) with n=1024
}
