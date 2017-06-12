package quasar

type event struct {
	topicDigest *hash160digest
	message     []byte
	publishers  []pubkey
	ttl         uint32
}

func newEvent(topic []byte, message []byte, ttl uint32) *event {
	if topic == nil {
		topic = []byte{}
	}
	if message == nil {
		message = []byte{}
	}
	digest := hash160(topic)
	return &event{
		topicDigest: &digest,
		message:     message,
		publishers:  []pubkey{},
		ttl:         ttl,
	}
}

func validEvent(e *event) bool {
	return e != nil && e.message != nil &&
		e.publishers != nil && e.topicDigest != nil
}

// func serializeEvent(e *event) []byte {
// 	return nil // TODO implement
// }
//
// func deserializeEvent(data []byte) *event {
// 	return nil // TODO implement
// }
