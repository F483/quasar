package quasar

// Compressed public key: 33 bytes

// QUASAR PACKET
// * Public key hash: 20 bytes
// * Signature: 65 bytes
// * Nonce: 4 byte (for salting and adding work)
// * Type: 1 byte
// * Payload: 1024 bytes max

type networkOverlay interface {
	id() pubkey
	connectedPeers() []*pubkey
	isConnected(peerId *pubkey) bool
	receivedEventChannel() chan *event
	receivedUpdateChannel() chan *update
	sendEvent(peerId *pubkey, e *event)
	sendUpdate(peerId *pubkey, filters [][]byte)
	start()
	stop()
}
