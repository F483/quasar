package quasar

import "github.com/btcsuite/btcutil"

type hash160digest [20]byte // ripemd160(sha256(data))
type pubkey [33]byte        // compressed secp256k1 public key
type signature [65]byte

func hash160(data []byte) hash160digest {
	digest := hash160digest{}
	copy(digest[:], btcutil.Hash160(data))
	return digest
}
