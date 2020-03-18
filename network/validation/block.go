package validation

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"time"

	"github.com/buuzcoin/blockchain"
	"github.com/buuzcoin/network"
)

// This file decribes how block should be checked in Buuzcoin network

// CurrentBlockVersion is current version of block supported by this implementation
const CurrentBlockVersion = 1

var (
	// ErrMalformedBlock is returned if block data is malformed
	ErrMalformedBlock = errors.New("validation: malformed block data")
	// ErrBlockVersionUnsupported is returned on validation if block version is not supported
	ErrBlockVersionUnsupported = errors.New("vaildation: unsupported block version")
	// ErrInvalidBlockHash is returned on validation if block hash isn't equal to hash calculated over block header
	ErrInvalidBlockHash = errors.New("vaildation: invalid block hash")
	// ErrInvalidBlockTimestamp is returned on validation if block time is invalid
	ErrInvalidBlockTimestamp = errors.New("vaildation: invalid timestamp")
	// ErrInvalidMerkleRoot is returned on validation if block's MerkleRoot doesn't match calculated one
	ErrInvalidMerkleRoot = errors.New("vaildation: invalid merkle root")
	// ErrInvalidBlockIndex is returned on validation if block's index isn't previous block's next one
	ErrInvalidBlockIndex = errors.New("vaildation: invalid block index")
)

// CheckGenesisBlock checks if genesis block satisfies network requirements.
// This function doesn't verify StateMerkleRoot value
func CheckGenesisBlock(block blockchain.Block) (bool, error) {
	if block.Version > CurrentBlockVersion {
		return false, ErrBlockVersionUnsupported
	}
	if block.Index != 0 {
		return false, ErrInvalidBlockIndex
	}

	if len(block.Beneficiary) != network.AddressSize {
		return false, ErrMalformedBlock
	}
	if len(block.AdditionalData) > 32 {
		return false, ErrMalformedBlock
	}

	if len(block.TxHashes) != 0 {
		return false, ErrMalformedBlock
	}

	// Block cannot be created later than 15 minutes in future
	if block.Timestamp > time.Now().Add(15*time.Minute).Unix() {
		return false, ErrInvalidBlockTimestamp
	}

	if bytes.Compare(block.PrevBlockHash, bytes.Repeat([]byte{0x00}, 32)) != 0 {
		return false, ErrInvalidBlockHash
	}

	if bytes.Compare(blockchain.CalculateMerkleRoot(block.TxHashes), block.TxMerkleRoot) != 0 {
		return false, ErrInvalidMerkleRoot
	}

	blockHash := block.CalculateHash()
	pubKey := block.Signature[ed25519.SignatureSize:]
	if bytes.Compare(network.DeriveAddress(pubKey), block.Beneficiary) != 0 {
		return false, ErrMalformedBlock
	}
	if !ed25519.Verify(pubKey, blockHash, block.Signature[:ed25519.SignatureSize]) {
		return false, ErrMalformedBlock
	}

	return true, nil
}

// CheckBlock checks if block satisfies network requirements.
// This function doesn't verify StateMerkleRoot value
func CheckBlock(block, prevBlock blockchain.Block) (bool, error) {
	if block.Version > CurrentBlockVersion {
		return false, ErrBlockVersionUnsupported
	}
	if block.Index-1 != prevBlock.Index {
		return false, ErrInvalidBlockIndex
	}

	if len(block.Beneficiary) != network.AddressSize {
		return false, ErrMalformedBlock
	}
	if len(block.AdditionalData) > 32 {
		return false, ErrMalformedBlock
	}
	if len(block.ProofData) == 0 {
		return false, ErrMalformedBlock
	}

	for _, txHash := range block.TxHashes {
		if len(txHash) != 32 {
			return false, ErrMalformedBlock
		}
	}

	// Block cannot be created later than 15 minutes in future
	if block.Timestamp > time.Now().Add(15*time.Minute).Unix() {
		return false, ErrInvalidBlockTimestamp
	}
	// Block cannot be created before previous block
	if block.Timestamp < prevBlock.Timestamp {
		return false, ErrInvalidBlockTimestamp
	}

	if bytes.Compare(block.PrevBlockHash, prevBlock.CalculateHash()) != 0 {
		return false, ErrInvalidBlockHash
	}

	if bytes.Compare(blockchain.CalculateMerkleRoot(block.TxHashes), block.TxMerkleRoot) != 0 {
		return false, ErrInvalidMerkleRoot
	}

	blockHash := block.CalculateHash()
	pubKey := block.Signature[ed25519.SignatureSize:]
	if bytes.Compare(network.DeriveAddress(pubKey), block.Beneficiary) != 0 {
		return false, ErrMalformedBlock
	}
	if !ed25519.Verify(pubKey, blockHash, block.Signature[:ed25519.SignatureSize]) {
		return false, ErrMalformedBlock
	}

	return true, nil
}
