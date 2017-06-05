package quasar

import (
//	"github.com/willf/bloom"
//	"sync"
)

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
