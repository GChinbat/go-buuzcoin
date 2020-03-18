package network

import "github.com/buuzcoin/go-buuzcoin/blockchain"

// This file decribes rules for transaction in Buuzcoin network

// Every transaction consumes gas. Gas costs are defined below
const (
	// GasPerTxByte is gas cost for every byte in encoded transaction
	GasPerTxByte uint64 = 1
	// GasPerTxAdditionalDataByte is gas cost for every byte in lockingData.Data
	GasPerTxAdditionalDataByte uint64 = 5
	// GasParTxToUserAccount is gas cost for sending coins to user account
	GasPerTxToUserAccount uint64 = 21000
)

// MinimalGasFee is minimal price for 1 gas unit
const MinimalGasFee = 2

// CurrentTxVersion is current version of transaction supported by this implementation
const CurrentTxVersion = 1

// GetGasAmount returns amount of gas consumed by transaction
func GetGasAmount(tx blockchain.TX) uint64 {
	var gasCost uint64 = 0
	gasCost += uint64(len(tx.OptionalData)) * GasPerTxAdditionalDataByte
	gasCost += uint64(len(tx.GetBinaryData())) * GasPerTxByte
	return gasCost
}
