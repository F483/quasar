package quasar

var testConfig = Config{
	DefaultEventTTL:  32,
	FilterFreshness:  64,
	PropagationDelay: 24,
	HistoryLimit:     1000000,
	HistoryAccuracy:  0.000001,
	FiltersDepth:     8,
	FiltersM:         10240, // m 1k
	FiltersK:         7,     // hashes
}
