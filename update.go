package quasar

type update struct {
	peer   *pubkey
	index  uint32
	filter []byte
}

func validUpdate(u *update, c *Config) bool {
	return u != nil && u.peer != nil && u.index < c.FiltersDepth &&
		uint64(len(u.filter)) == (c.FiltersM/8)
}

func serializeUpdate(u *update) []byte {
	return nil // TODO implement
}

func deserializeUpdate(data []byte) *update {
	return nil // TODO implement
}
