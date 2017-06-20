package quasar

import (
	"bytes"
	"encoding/binary"
	"github.com/willf/bloom"
)

// FIXME always binary.BigEndian

func serializeFilter(f *bloom.BloomFilter) []byte {
	data, err := f.GobEncode()
	mustNotError(err)
	return data[24:] // remove known header
}

func deserializeFilter(d []byte, c *Config) *bloom.BloomFilter {
	m := c.FiltersM
	k := c.FiltersK

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

func newFilters(c *Config) [][]byte {
	m := c.FiltersM
	filters := make([][]byte, 0)
	for i := 0; uint32(i) < c.FiltersDepth; i++ {
		filters = append(filters, make([]byte, m/8, m/8))
	}
	return filters
}

func filterInsert(f []byte, c *Config, datas ...[]byte) []byte {
	ds := make([]hash160digest, len(datas), len(datas))
	for i, data := range datas {
		ds[i] = hash160(data)
	}
	return filterInsertDigest(f, c, ds...)
}

func filterInsertDigest(f []byte, c *Config, ds ...hash160digest) []byte {
	bf := deserializeFilter(f, c)
	for _, d := range ds {
		bf.Add(d[:])
	}
	return serializeFilter(bf)
}

func mergeFilters(fs ...[]byte) []byte {
	if len(fs) == 0 {
		return nil
	}
	size := len(fs[0])
	result := make([]byte, size, size)
	errMsg := "Filter m missmatch: %d != %d"
	for _, f := range fs {
		mustBeTrue(size == len(f), errMsg, size, len(f))

		m := make([]byte, size, size)
		for i := 0; i < len(m); i++ {
			m[i] = result[i] | f[i]
		}
		result = m
	}
	return result
}

func filterContains(f []byte, c *Config, data []byte) bool {
	return filterContainsDigest(f, c, hash160(data))
}

func filterContainsDigest(f []byte, c *Config, d hash160digest) bool {
	bf := deserializeFilter(f, c)
	return bf.Test(d[:])
}
