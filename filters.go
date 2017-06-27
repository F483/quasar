package quasar

import (
	"bytes"
	"encoding/binary"
	"github.com/willf/bloom"
)

// FIXME always binary.BigEndian?

// TODO define types and use methods for better namespacing
// type filters [][]byte
// type filter []byte

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

func clearFilters(filters [][]byte) {
	d := len(filters) // depth
	for i := 0; i < d; i++ {
		w := len(filters[i]) // width
		for j := 0; j < w; j++ {
			filters[i][j] = 0
		}
	}
}

func newFilters(c *Config) [][]byte {
	m := c.FiltersM
	filters := make([][]byte, c.FiltersDepth, c.FiltersDepth)
	for i := 0; uint32(i) < c.FiltersDepth; i++ {
		filters[i] = make([]byte, m/8, m/8)
	}
	return filters
}

func newFilterFromDigests(c *Config, digests ...*sha256digest) []byte {
	m := c.FiltersM
	k := c.FiltersK
	bf := bloom.New(uint(m), uint(k))
	for _, digest := range digests {
		bf.Add(digest[:])
	}
	return serializeFilter(bf)
}

func filterInsert(f []byte, c *Config, datas ...[]byte) []byte {
	ds := make([]*sha256digest, len(datas), len(datas))
	for i, data := range datas {
		digest := sha256sum(data)
		ds[i] = &digest
	}
	return filterInsertDigest(f, c, ds...)
}

func filterInsertDigest(f []byte, c *Config, ds ...*sha256digest) []byte {
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
	m := make([]byte, size, size)
	for _, f := range fs {
		for i := 0; i < len(m); i++ {
			m[i] = m[i] | f[i]
		}
	}
	return m
}

func filterContains(f []byte, c *Config, data []byte) bool {
	digest := sha256sum(data)
	return filterContainsDigest(f, c, &digest)
}

func filterContainsDigest(f []byte, c *Config, d *sha256digest) bool {
	bf := deserializeFilter(f, c)
	return bf.Test(d[:])
}
