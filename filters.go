package quasar

import "fmt"

// import (
//	"github.com/willf/bloom"
//	"sync"
// )

// Safe (udp will not fragment)
// n: 1024 // number of elements
// m: 3408 // filter size = 426byte
// k: 2    // number of hash functions = (m / n) log(2)

// Safeish (UDP unlikely to fragment as under MTU 1460) ?
// n: 1024 // number of elements
// m: 8192 // filter size = 1k
// k: 6    // number of hash functions = (m / n) log(2)

// type Filters struct {
// 	filters []*bloom.BloomFilter
// 	mutex   *sync.RWMutex
// }

// func NewFilters(depth uint32) *Filters {
// 	filters := make([]*bloom.BloomFilter, depth)
// 	for i := 0; uint(i) < depth; i++ {
// 		filters[i] = bloom.New(m, k)
// 	}
// 	return &Filters{
// 		filters:  filters,
// 		depth:    depth,
// 		limit:    limit,
// 		accuracy: accuracy,
// 		mutex:    new(sync.RWMutex),
// 	}
// }
//
// func (f Filters) rebuild(topics []Hash160Digest, peerFilters []*Filters) {
//
// }

type filter struct {
	m uint32
	k uint32
	b []byte
}

func newFilterStack(cfg *Config) [][]byte {
	filters := make([][]byte, 0)
	for i := 0; uint32(i) < cfg.FiltersDepth; i++ {
		m := cfg.FiltersM
		if m%8 != 0 {
			panic(fmt.Sprintf("Invalid filter m: %v not multiple of 8!", m))
		}
		filters = append(filters, make([]byte, m/8, m/8))
	}
	return filters
}

func unmarshalFilter(b []byte, cfg *Config) *filter {
	if uint32(len(b)) != cfg.FiltersM {
		panic(fmt.Sprintf("Invalid filter m: %v != %v!", len(b), cfg.FiltersM))
	}
	return &filter{b: b, m: cfg.FiltersM, k: cfg.FiltersK}
}

func mergeFilters(a []byte, b []byte, cfg *Config) []byte {
	return nil // TODO implement
}

func fsInsert(fs [][]byte, index int, cfg *Config, data []byte) {
	fsInsertDigest(fs, index, cfg, hash160(data))
}

func fsInsertDigest(fs [][]byte, index int, cfg *Config, d hash160digest) {
	// TODO implement
}

func (f *filter) insert(data []byte) {
	f.insertDigest(hash160(data))
}

func (f *filter) insertDigest(digest hash160digest) {
	// TODO implement
}

func (f *filter) contains(data []byte) bool {
	return f.containsDigest(hash160(data))
}

func (f *filter) containsDigest(digest hash160digest) bool {
	return false // TODO implement
}
