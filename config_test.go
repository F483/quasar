package quasar

var testConfig = Config{
	DefaultEventTTL:  32,
	FilterFreshness:  64,
	PropagationDelay: 24,
	HistoryLimit:     65536,
	HistoryAccuracy:  1.0 / 65536.0,
	FiltersDepth:     8,
	FiltersM:         10240, // m 1k
	FiltersK:         7,     // hashes
}
