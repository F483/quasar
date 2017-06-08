package quasar

// Compressed public key: 33 bytes

// QUASAR PACKET
// * Public key hash: 20 bytes
// * Signature: 65 bytes
// * Nonce: 4 byte (for salting and adding work)
// * Type: 1 byte
// * Payload: 1024 bytes max

type overlayNetwork interface {
	Id() pubkey
	ConnectedPeers() []pubkey
	ReceivedEventChannel() chan *event
	ReceivedUpdateChannel() chan *update
	SendEvent(*pubkey, *event)
	SendUpdate(receiver *pubkey, index uint, filter []byte)
	Start()
	Stop()
}
