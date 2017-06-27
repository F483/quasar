package quasar

import (
	"github.com/vmihailenco/msgpack"
)

type event struct {
	TopicDigest *sha256digest
	Message     []byte
	Publishers  []*sha256digest
	Ttl         uint32
}

func newEvent(topic []byte, message []byte, ttl uint32) *event {
	if topic == nil {
		topic = []byte{}
	}
	if message == nil {
		message = []byte{}
	}
	digest := sha256sum(topic)
	return &event{
		TopicDigest: &digest,
		Message:     message,
		Publishers:  []*sha256digest{},
		Ttl:         ttl,
	}
}

func (e *event) addPublisher(nodeId *pubkey) {
	nodeIdDigest := sha256sum(nodeId[:])
	e.Publishers = append(e.Publishers, &nodeIdDigest)
}

func (e *event) inPublishers(peerId *pubkey) bool {
	peerIdDigest := sha256sum(peerId[:])
	for _, publisherIdDigest := range e.Publishers {
		if *publisherIdDigest == peerIdDigest {
			return true
		}
	}
	return false
}

func (e *event) valid() bool {
	// TODO check for nil publishers and topic digest?
	return e != nil && e.Message != nil &&
		e.Publishers != nil && e.TopicDigest != nil
}

func (e *event) marshal() []byte {
	// TODO better more compact serialization
	b, err := msgpack.Marshal(e)
	if err != nil {
		panic(err) // TODO can any valid input event actually error?
	}
	return b
}

func unmarshalEvent(data []byte) (*event, error) {
	var e event
	err := msgpack.Unmarshal(data, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}
