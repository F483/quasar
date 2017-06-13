package quasar

import "time"

type peerData struct {
	filters    [][]byte
	timestamps []uint64 // unixtime
}

func makePeerTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}

func peerDataExpired(p *peerData, c config) bool {
	now := makePeerTimestamp()
	for _, timestamp := range p.timestamps {
		if timestamp >= (now - c.filterFreshness) {
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

func validUpdate(u *peerUpdate, c config) bool {
	return u != nil && u.peer != nil &&
		u.index < (c.filtersDepth-1) && // top filter never propagated
		uint64(len(u.filter)) == (c.filtersM/8)
}

// func serializePeerUpdate(u *peerUpdate) []byte {
// 	return nil // TODO implement
// }
//
// func deserializePeerUpdate(data []byte) *peerUpdate {
// 	return nil // TODO implement
// }
