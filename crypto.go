package quasar

import "github.com/btcsuite/btcutil"

type Hash160Digest [20]byte     // ripemd160(sha256(data))
type PubKey [33]byte            // compressed secp256k1 public key
type PubKeyDigest Hash160Digest // ripemd160(sha256(PubKey))
type PeerID PubKey              // id for authentication/encryption
type Signature [65]byte

func Hash160(topic string) *Hash160Digest {
	digest := Hash160Digest{}
	copy(digest[:], btcutil.Hash160([]byte(topic)))
	return &digest
}
