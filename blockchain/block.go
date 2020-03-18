package blockchain

import (
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

/*
	Block hash is calculated using block header with following fields:
	Version - 4 bytes
	Index - 8 bytes
	Timestamp - 8 bytes
	PrevBlockHash - 32 bytes
	TxMerkleRoot - 32 bytes
	StateMerkleRoot - 32 bytes
	Beneficiary - 20 bytes
	AdditionalData - 1 byte length + varlength bytes

	Numbers are encoded in little-endian format.
*/

// GetBlockHeader returns block header in binary format
func (b Block) GetBlockHeader() []byte {
	blockHeaderLength := 4 + 8 + 8 + 32 + 32 + 32 + 20 + 1 + 4 + len(b.AdditionalData) + len(b.ProofData)

	blockHeader := make([]byte, 4+8+8, blockHeaderLength)
	binary.LittleEndian.PutUint32(blockHeader[:4], b.Version)
	binary.LittleEndian.PutUint64(blockHeader[4:4+8], b.Index)
	binary.LittleEndian.PutUint64(blockHeader[4+8:4+8+8], uint64(b.Timestamp))

	blockHeader = append(blockHeader, b.PrevBlockHash...)
	blockHeader = append(blockHeader, b.TxMerkleRoot...)
	blockHeader = append(blockHeader, b.StateMerkleRoot...)
	blockHeader = append(blockHeader, b.Beneficiary...)

	blockHeader = append(blockHeader, byte(len(b.AdditionalData)&0xFF))
	blockHeader = append(blockHeader, b.AdditionalData...)

	return blockHeader
}

// CalculateHash returns hash of specific block calculated over its header
func (b Block) CalculateHash() []byte {
	hash := sha3.Sum256(b.GetBlockHeader())
	return hash[:]
}
