package quasar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/willf/bloom"
)

// Safeish (UDP unlikely to fragment as under MTU 1460) ?
// n: 1024 // number of elements
// m: 8192 // filter size = 1k
// k: 6    // number of hash functions = (m / n) log(2)

func filterEncode(f *bloom.BloomFilter) ([]byte, error) {
	data, err := f.GobEncode()
	if err != nil {
		return nil, err
	}
	return data[24:], nil // remove known header
}

func filterDecode(d []byte, c *Config) (*bloom.BloomFilter, error) {
	m := c.FiltersM
	k := c.FiltersK
	if m%64 != 0 {
		return nil, fmt.Errorf("M not multiple of 64: %d!", m)
	}
	if uint64(len(d)) != (m / 8) {
		return nil, fmt.Errorf("Data len '%d' != '%d'!", len(d), m/8)
	}

	// recreate header
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, m)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.BigEndian, k)
	if err != nil {
		return nil, err
	}
	err = binary.Write(&buf, binary.BigEndian, m)
	if err != nil {
		return nil, err
	}
	header := buf.Bytes()

	// recreate and decode gob
	gob := append(header, d...)
	f := bloom.New(uint(m), uint(k))
	err = f.GobDecode(gob)
	if err != nil {
		return nil, err
	}
	return f, nil
}

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
	bf, err := filterDecode(f, c)
	if err != nil {
		// TODO panic, shouldn't happen after validation
	}
	bf.Add(d[:])
	updated, err := filterEncode(bf)
	if err != nil {
		// TODO panic, shouldn't happen after input validation
	}
	return updated
}

func filterMerge(a []byte, b []byte, c *Config) []byte {
	abf, err := filterDecode(a, c)
	if err != nil {
		// TODO panic, shouldn't happen after input validation
	}
	bbf, err := filterDecode(a, c)
	if err != nil {
		// TODO panic, shouldn't happen after input validation
	}
	err = abf.Merge(bbf)
	if err != nil {
		// TODO panic, shouldn't happen after input validation
	}
	merged, err := filterEncode(abf)
	if err != nil {
		// TODO panic, shouldn't happen after input validation
	}
	return merged
}

func filterContains(f []byte, data []byte, c *Config) bool {
	return filterContainsDigest(f, hash160(data), c)
}

func filterContainsDigest(f []byte, d hash160digest, c *Config) bool {
	bf, err := filterDecode(f, c)
	if err != nil {
		// TODO panic, shouldn't happen after validation
	}
	return bf.Test(d[:])
}
