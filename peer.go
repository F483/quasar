package quasar

import "time"

type peerData struct {
	filters   [][]byte
	timestamp uint64 // unixtime
}

func makePeerTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) / uint64(time.Millisecond)
}

func peerDataExpired(p *peerData, c *Config) bool {
	return (makePeerTimestamp() - c.FilterFreshness) > p.timestamp
}

type peerUpdate struct {
	peer    *pubkey
	filters [][]byte
}

func validUpdate(u *peerUpdate, c *Config) bool {
	if u == nil || c == nil || u.peer == nil {
		return false
	}
	if uint32(len(u.filters)) != c.FiltersDepth {
		return false
	}
	for _, f := range u.filters {
		if uint64(len(f)) != (c.FiltersM / 8) {
			return false
		}
	}
	return true
}

// func serializePeerUpdate(u *peerUpdate) []byte {
// 	return nil // TODO implement
// }
//
// func deserializePeerUpdate(data []byte) *peerUpdate {
// 	return nil // TODO implement
// }
