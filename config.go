package quasar

// Config holds the configuration for a quasar node.
type Config struct {
	DefaultEventTTL  uint32  // decremented every hop
	FilterFreshness  uint64  // in milliseconds
	PropagationDelay uint64  // in milliseconds
	HistoryLimit     uint32  // entries remembered
	HistoryAccuracy  float64 // chance of error
	FiltersDepth     uint32  // filter stack height
	FiltersM         uint64  // size in bits (multiple of 64)
	FiltersK         uint64  // number of hashes
}

// StandardConfig uses well agreed upon values, only deviate for testing.
var StandardConfig = Config{
	DefaultEventTTL:  32,              // 4 missed wells
	FilterFreshness:  50000,           // 50sec (>2.5 propagation)
	PropagationDelay: 20000,           // 20sec (~0.5M/min const traffic)
	HistoryLimit:     1048576,         // remember last 1G
	HistoryAccuracy:  1.0 / 1048576.0, // avg 1 false positive per 1G
	FiltersDepth:     8,               // reaches many many nodes
	FiltersM:         10240,           // keep packet under safe MTU 1400
	FiltersK:         7,               // =(m / n) log(2) with n=1024
}
