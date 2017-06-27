package quasar

import (
	"github.com/vmihailenco/msgpack"
)

type update struct {
	NodeId  *pubkey
	Filters [][]byte
}

func newUpdate(nodeId *pubkey, filters [][]byte) *update {
	return &update{NodeId: nodeId, Filters: filters}
}

func (u *update) valid(c *Config) bool {
	if u == nil || c == nil || u.NodeId == nil {
		return false
	}
	// TODO check nodeid length
	if uint32(len(u.Filters)) != c.FiltersDepth {
		return false
	}
	for _, f := range u.Filters {
		if uint64(len(f)) != (c.FiltersM / 8) {
			return false
		}
	}
	return true
}

func (u *update) marshal() []byte {
	// TODO better more compact serialization
	b, err := msgpack.Marshal(u)
	if err != nil {
		panic(err)
	}
	return b
}

func unmarshalUpdate(data []byte) (*update, error) {
	var u update
	err := msgpack.Unmarshal(data, &u)
	if err != nil {
		return nil, err // TODO can any valid input update actually error?
	}
	return &u, nil
}
