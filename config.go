package quasar

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

/*
Safeish Filter (UDP unlikely to fragment as under MTU 1460)
n: 1024 // number of elements
m: 8192 // filter size = 1k
k: 6    // number of hash functions = (m / n) log(2)
*/
var DefaultConfig = Config{
	DefaultEventTTL:  32,              // 4 missed wells
	FilterFreshness:  50000,           // 50sec (>2.5 propagation)
	PropagationDelay: 20000,           // 20sec (~0.5M/min const traffic)
	HistoryLimit:     1048576,         // remember last 1G
	HistoryAccuracy:  1.0 / 1048576.0, // avg 1 false positive per 1G
	FiltersDepth:     8,               // reaches many many nodes
	FiltersM:         8192,            // 1K (packet under safe MTU 1400)
	FiltersK:         6,               // (m / n) log(2)
}
