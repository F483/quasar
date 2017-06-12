package quasar

type config struct {
	defaultEventTTL  uint32  // decremented every hop
	filterFreshness  uint64  // in seconds TODO change to ms
	propagationDelay uint64  // in seconds TODO change to ms
	historyLimit     uint32  // entries remembered
	historyAccuracy  float64 // chance of error
	filtersDepth     uint32  // filter stack height
	filtersM         uint64  // size in bits (multiple of 64)
	filtersK         uint64  // number of hashes
}

/*
Safeish Filter (UDP unlikely to fragment as under MTU 1460)
n: 1024 // number of elements
m: 8192 // filter size = 1k
k: 6    // number of hash functions = (m / n) log(2)
*/
var defaultConfig = config{
	defaultEventTTL:  32,              // 4 missed wells
	filterFreshness:  50,              // 50sec (>2.5 propagation)
	propagationDelay: 20,              // 20sec (~0.5M/min const traffic)
	historyLimit:     1048576,         // remember last 1G
	historyAccuracy:  1.0 / 1048576.0, // avg 1 false positive per 1G
	filtersDepth:     8,               // reaches many many nodes
	filtersM:         8192,            // 1K (packet under safe MTU 1400)
	filtersK:         6,               // (m / n) log(2)
}
