package quasar

import (
//	"github.com/willf/bloom"
)

// Safeish (UDP unlikely to fragment as under MTU 1460) ?
// n: 1024 // number of elements
// m: 8192 // filter size = 1k
// k: 6    // number of hash functions = (m / n) log(2)

func newFilters(c *Config) [][]byte {
	m := c.FiltersM
	filters := make([][]byte, 0)
	for i := 0; uint32(i) < c.FiltersDepth; i++ {
		filters = append(filters, make([]byte, m/8, m/8))
	}
	return filters
}

func filterInsert(f []byte, data []byte, c *Config) []byte {
	return filterInsertDigest(f, hash160(data), c)
}

func filterInsertDigest(f []byte, d hash160digest, c *Config) []byte {
	// TODO implement
	return nil
}

func filterMerge(a []byte, b []byte, c *Config) []byte {
	return nil // TODO implement
}

func filterContains(f []byte, data []byte, c *Config) bool {
	return filterContainsDigest(f, hash160(data), c)
}

func filterContainsDigest(f []byte, d hash160digest, c *Config) bool {
	return false // TODO implement
}
