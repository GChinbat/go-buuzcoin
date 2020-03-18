package consensus

import (
	"bytes"
	"crypto/ed25519"

	"github.com/buuzcoin/go-buuzcoin/blockchain"
	"github.com/buuzcoin/go-buuzcoin/network"
	"golang.org/x/crypto/sha3"
)

/*
	This file defines Proof-of-Authority algorithm used in testnet v1
	ProofData in block is calculated in following way:
	1. Get hash to be signed by authority: SHA3(Beneficiary || AdditionalData || BlockHash)
	2. Sign hash and append public key of beneficiary
*/

// ProofOfAuthority defines proof of authority algorithm used in testnet v1
type ProofOfAuthority struct {
	AuthorityPublicKey []byte
}

// IsValidBlock checks if block satisfies PoA algorithm requirements specified above
func (poa *ProofOfAuthority) IsValidBlock(block blockchain.Block) (bool, error) {
	if len(block.ProofData) != ed25519.PublicKeySize+ed25519.SignatureSize {
		return false, nil
	}

	signature := block.ProofData[:ed25519.SignatureSize]
	pubKey := block.ProofData[ed25519.SignatureSize:]
	if bytes.Compare(pubKey, poa.AuthorityPublicKey) != 0 {
		return false, nil
	}

	if bytes.Compare(network.DeriveAddress(pubKey), block.Beneficiary) != 0 {
		return false, nil
	}

	hash := sha3.New256()
	hash.Write(block.Beneficiary)
	hash.Write(block.AdditionalData)
	hash.Write(block.CalculateHash())
	if !ed25519.Verify(pubKey, hash.Sum(nil), signature) {
		return false, nil
	}
	return true, nil
}
