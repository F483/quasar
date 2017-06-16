package quasar

import "github.com/btcsuite/btcutil"

// TODO move to own lib
// TODO be functional for simplicity/security
// ECIES secp256k1

type hash160digest [20]byte // ripemd160(sha256(data))
type pubkey [33]byte        // compressed secp256k1 public key
type signature [65]byte

func hash160(data []byte) hash160digest {
	// TODO handles nil data correctly?
	digest := hash160digest{}
	copy(digest[:], btcutil.Hash160(data))
	return digest
}

// GENERATION
// TODO RandomBytes(int) -> []byte
// TODO GenPrivKey() -> privkey
// TODO GetPubKey(privkey) -> pubkey

// SIGNING
// TODO Sign(privkey, data) -> signature
// TODO Validate(pubkeyDigest, data, signature) -> (ok, pubkey)

// ENCRYPTION
// TODO Encrypt(data, pubkey) -> encryptedData
// TODO Decrypt(encryptedData, privkey) -> data

// ENCODING: BTC WIF
// TODO WifFromPrivKey(privkey, netcode) -> wif
// TODO WifToPrivkey(wif) -> privkey

// ENCODING: BTC ADDRESS
// TODO AddressFromDigest(digest, netcode) -> address
// TODO AddressFromPubkey(pubkey, netcode) -> address
// TODO AddressFromWif(wif) -> address
// TODO AddressToDigest(address) -> digest

// ENCODING: DER
// TODO DerFromPrivKey(privkey) -> der
// TODO DerToPrivKey(der) -> privkey

// ENCODING: PEM
// TODO PemFromPrivKey(privkey) -> pem
// TODO PemToPrivKey(pem) -> privkey
