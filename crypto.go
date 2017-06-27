package quasar

import (
	"crypto/sha256"
	"github.com/btcsuite/btcutil"
)

// TODO move to own lib
// TODO be functional for simplicity/security
// ECIES secp256k1

type pubkey [33]byte // compressed secp256k1 public key
type signature [65]byte

//////////////////////////////////////////////
// BITCOIN HASH160: RIPEMD160(SHA256(data)) //
//////////////////////////////////////////////

type hash160digest [20]byte

func hash160sum(data []byte) hash160digest {
	// TODO handles nil data correctly?
	digest := hash160digest{}
	copy(digest[:], btcutil.Hash160(data))
	return digest
}

////////////
// SHA256 //
////////////

// wrap for more readable code

type sha256digest [sha256.Size]byte

func sha256sum(data []byte) sha256digest {
	return sha256.Sum256(data)
}

////////////////
// GENERATION //
////////////////

// TODO RandomBytes(int) -> []byte
// TODO GenPrivKey() -> privkey
// TODO GetPubKey(privkey) -> pubkey

/////////////
// SIGNING //
/////////////

// TODO Sign(privkey, data) -> signature
// TODO Validate(pubkeyDigest, data, signature) -> (ok, pubkey)

////////////////
// ENCRYPTION //
////////////////

// TODO Encrypt(data, pubkey) -> encryptedData
// TODO Decrypt(encryptedData, privkey) -> data

///////////////////////
// ENCODING: BTC WIF //
///////////////////////

// TODO WifFromPrivKey(privkey, netcode) -> wif
// TODO WifToPrivkey(wif) -> privkey

///////////////////////////
// ENCODING: BTC ADDRESS //
///////////////////////////

// TODO AddressFromDigest(digest, netcode) -> address
// TODO AddressFromPubkey(pubkey, netcode) -> address
// TODO AddressFromWif(wif) -> address
// TODO AddressToDigest(address) -> digest

///////////////////
// ENCODING: DER //
///////////////////

// TODO DerFromPrivKey(privkey) -> der
// TODO DerToPrivKey(der) -> privkey

///////////////////
// ENCODING: PEM //
///////////////////

// TODO PemFromPrivKey(privkey) -> pem
// TODO PemToPrivKey(pem) -> privkey
