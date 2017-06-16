package quasar

import "time"

type peerData struct {
	filters    [][]byte
	timestamps []uint64 // unixtime
}

func makePeerTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}

func peerDataExpired(p *peerData, c *Config) bool {
	now := makePeerTimestamp()
	for _, timestamp := range p.timestamps {
		if timestamp >= (now - c.FilterFreshness) {
			return false
		}
	}
	return true
}

type peerUpdate struct {
	peer   *pubkey
	index  uint32
	filter []byte
}

func validUpdate(u *peerUpdate, c *Config) bool {
	return u != nil && c != nil && u.peer != nil &&
		u.index < (c.FiltersDepth-1) && // top filter never propagated
		uint64(len(u.filter)) == (c.FiltersM/8)
}

// func serializePeerUpdate(u *peerUpdate) []byte {
// 	return nil // TODO implement
// }
//
// func deserializePeerUpdate(data []byte) *peerUpdate {
// 	return nil // TODO implement
// }
