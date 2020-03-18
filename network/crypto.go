package network

import (
	"crypto/ed25519"

	"golang.org/x/crypto/sha3"
)

/*
	This file specifies cryptographic primitives used in Buuzcoin blockchain.

	Hash algorithm: SHA3-256
	ECDSA: ed25519
	Address format: SHA3-256/160
*/

// AddressSize is size of address in bytes
const AddressSize = 20

// DeriveAddress calculates address using ECDSA public key
func DeriveAddress(pubKey ed25519.PublicKey) []byte {
	hash := sha3.New256()
	hash.Write(pubKey)
	// Take last AddressSize bytes of hash
	return hash.Sum(nil)[32-AddressSize:]
}
