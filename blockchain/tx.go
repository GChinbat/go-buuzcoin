package blockchain

import (
	"bytes"
	"encoding/binary"

	"golang.org/x/crypto/sha3"
)

/*
	Hash of transaction is calculated over binary data in following format:
	Version - 4 bytes
	From - 20 bytes
	Nonce - 8 bytes
	To - 20 bytes
	Amount - 8 bytes
	Fee - 8 bytes
	OptionalData - 4 byte length + varlength bytes
	GasLimit - 8 bytes
	GasPrice - 8 bytes

	All numbers are encoded in little-endian format
*/

// GetBinaryData returns TX data in binary format as specified above
func (tx TX) GetBinaryData() []byte {
	dataLength := 4 + 20 + 8 + 20 + 8 + 8 + 4 + len(tx.OptionalData) + 8 + 8

	data := make([]byte, 4, dataLength)
	binary.LittleEndian.PutUint32(data[:4], tx.Version)

	data = append(data, tx.From...)

	data = append(data, bytes.Repeat([]byte{0x00}, 8)...)
	binary.LittleEndian.PutUint64(data[len(data)-8:], tx.Nonce)

	data = append(data, tx.To...)

	data = append(data, bytes.Repeat([]byte{0x00}, 8)...)
	binary.LittleEndian.PutUint64(data[len(data)-8:], tx.Amount)

	data = append(data, bytes.Repeat([]byte{0x00}, 8)...)
	binary.LittleEndian.PutUint64(data[len(data)-8:], tx.Fee)

	data = append(data, bytes.Repeat([]byte{0x00}, 4)...)
	binary.LittleEndian.PutUint32(data[len(data)-4:], uint32(len(tx.OptionalData)))

	data = append(data, tx.OptionalData...)

	data = append(data, bytes.Repeat([]byte{0x00}, 8)...)
	binary.LittleEndian.PutUint64(data[len(data)-8:], tx.GasLimit)

	data = append(data, bytes.Repeat([]byte{0x00}, 8)...)
	binary.LittleEndian.PutUint64(data[len(data)-8:], tx.GasPrice)
	return data
}

// CalculateHash returns hash of specific transaction calculated over its binary data
func (tx TX) CalculateHash() []byte {
	hash := sha3.Sum256(tx.GetBinaryData())
	return hash[:]
}
