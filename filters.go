package quasar

import (
	"bytes"
	"encoding/binary"
	"github.com/willf/bloom"
)

func serializeFilter(f *bloom.BloomFilter) []byte {
	data, err := f.GobEncode()
	mustNotError(err)
	return data[24:] // remove known header
}

func deserializeFilter(d []byte, c config) *bloom.BloomFilter {
	m := c.filtersM
	k := c.filtersK

	// recreate header
	var buf bytes.Buffer
	mustNotError(binary.Write(&buf, binary.BigEndian, m))
	mustNotError(binary.Write(&buf, binary.BigEndian, k))
	mustNotError(binary.Write(&buf, binary.BigEndian, m))
	header := buf.Bytes()

	// recreate and decode gob
	gob := append(header, d...)
	f := bloom.New(uint(m), uint(k))
	mustNotError(f.GobDecode(gob))
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

func filterInsert(f []byte, c config, datas ...[]byte) []byte {
	ds := make([]hash160digest, len(datas), len(datas))
	for i, data := range datas {
		ds[i] = hash160(data)
	}
	return filterInsertDigest(f, c, ds...)
}

func filterInsertDigest(f []byte, c config, ds ...hash160digest) []byte {
	bf := deserializeFilter(f, c)
	for _, d := range ds {
		bf.Add(d[:])
	}
	return serializeFilter(bf)
}

func mergeFilters(a []byte, b []byte) []byte {

	errMsg := "Filter m missmatch: %d != %d"
	mustBeTrue(len(a) == len(b), errMsg, len(a), len(b))

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
