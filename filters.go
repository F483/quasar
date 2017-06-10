package quasar

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/willf/bloom"
)

/*
Safeish (UDP unlikely to fragment as under MTU 1460)
n: 1024 // number of elements
m: 8192 // filter size = 1k
k: 6    // number of hash functions = (m / n) log(2)
*/

func serializeFilter(f *bloom.BloomFilter) []byte {
	data, err := f.GobEncode()
	if err != nil { // shouldn't happen after input validation
		panic(fmt.Sprintf("Enexpected error: %s", err.Error()))
	}
	return data[24:] // remove known header
}

func deserializeFilter(d []byte, c config) *bloom.BloomFilter {
	m := c.filtersM
	k := c.filtersK

	// recreate header
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.BigEndian, m)
	if err != nil { // shouldn't happen after input validation
		panic(fmt.Sprintf("Enexpected error: %s", err.Error()))
	}
	err = binary.Write(&buf, binary.BigEndian, k)
	if err != nil { // shouldn't happen after input validation
		panic(fmt.Sprintf("Enexpected error: %s", err.Error()))
	}
	err = binary.Write(&buf, binary.BigEndian, m)
	if err != nil { // shouldn't happen after input validation
		panic(fmt.Sprintf("Enexpected error: %s", err.Error()))
	}
	header := buf.Bytes()

	// recreate and decode gob
	gob := append(header, d...)
	f := bloom.New(uint(m), uint(k))
	err = f.GobDecode(gob)
	if err != nil { // shouldn't happen after input validation
		panic(fmt.Sprintf("Enexpected error: %s", err.Error()))
	}
	return f
}

func newFilters(c config) [][]byte {
	m := c.filtersM
	filters := make([][]byte, 0)
	for i := 0; uint32(i) < c.filtersDepth; i++ {
		filters = append(filters, make([]byte, m/8, m/8))
	}
	return filters
}

func filterInsert(f []byte, data []byte, c config) []byte {
	// TODO use variable arguments instead
	return filterInsertDigest(f, hash160(data), c)
}

func filterInsertDigest(f []byte, d hash160digest, c config) []byte {
	// TODO use variable arguments instead
	bf := deserializeFilter(f, c)
	bf.Add(d[:])
	return serializeFilter(bf)
}

func mergeFilters(a []byte, b []byte) []byte {
	if len(a) != len(b) { // shouldn't happen after input validation
		panic(fmt.Sprintf("Filter m missmatch: %i != %i", len(a), len(b)))
	}
	c := make([]byte, len(a), len(a))
	for i := 0; i < len(c); i++ {
		c[i] = a[i] | b[i]
	}
	return c
}

func filterContains(f []byte, data []byte, c config) bool {
	return filterContainsDigest(f, hash160(data), c)
}

func filterContainsDigest(f []byte, d hash160digest, c config) bool {
	bf := deserializeFilter(f, c)
	return bf.Test(d[:])
}
