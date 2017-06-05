package quasar

import (
	"encoding/hex"
	"testing"
)

func TestHash160(t *testing.T) {

	// decode topic
	topicHex := []byte("f483")
	topicBytes := make([]byte, hex.DecodedLen(len(topicHex)))
	topicBytesLen, err := hex.Decode(topicBytes, topicHex)
	if err != nil {
		t.Fatal(err)
	}
	topic := topicBytes[:topicBytesLen]

	// decode expected digest
	expectedHex := []byte("4e0123796bee558240c5945ac9aff553fcc6256d")
	expected := hash160digest{}
	expectedBytesLen, err := hex.Decode(expected[:], expectedHex)
	if err != nil {
		t.Fatal(err)
	}
	if expectedBytesLen != 20 {
		t.Errorf("Incorrect digest size! %i", expectedBytesLen)
	}

	digest := hash160(topic)
	if digest != expected {
		t.Errorf("Hash160 failed!")
	}
}
