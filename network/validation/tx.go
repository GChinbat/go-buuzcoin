package validation

import (
	"bytes"
	"crypto/ed25519"
	"errors"

	"github.com/buuzcoin/go-buuzcoin/blockchain"
	"github.com/buuzcoin/go-buuzcoin/network"
)

// This file decribes rules for valid transaction in Buuzcoin network

/*
	Buuzcoin blockchain transaction consists of inputs and outputs.
	Inputs reference to previous transactions' outputs, this prevents outputs from being spent twice.

	Block's first transaction can be without inputs, it means this transaction uses block reward.
	Other transactions should spend less coins than tx inputs reference to.
*/

var (
	// ErrTxUnsupported is returned on validation if tx is not supported
	ErrTxUnsupported = errors.New("vaildation: unsupported transaction")
	// ErrInvalidTxHash is returned on validation if transaction hash is invalid
	ErrInvalidTxHash = errors.New("vaildation: invalid transaction hash")
	// ErrAlreadyUsedOutput is returned on validation if tx's input uses already used output
	ErrAlreadyUsedOutput = errors.New("vaildation: output is already used")
	// ErrInsufficientFunds is returned on validation if tx total amount is less than sender's balance
	ErrInsufficientFunds = errors.New("validation: insufficient funds")
	// ErrMalformedTx is returned if block data is malformed
	ErrMalformedTx = errors.New("validation: malformed transaction data")
	// ErrInsufficientGas is returned if evaluating transaction requires more gas than provided
	ErrInsufficientGas = errors.New("validation: insufficient gas")
	//ErrRejectedTx is returned if transaction is not used in chain and is rejected
	ErrRejectedTx = errors.New("validation: rejected transaction")
)

// GetAccountDataFn is function retriveving account data from data source using address specified
type GetAccountDataFn = func(address []byte) (*blockchain.AccountState, error)

// CheckTx checks whether if transaction data is valid
func CheckTx(tx blockchain.TX, getAccountData GetAccountDataFn) (bool, error) {
	if tx.Version > network.CurrentTxVersion {
		return false, ErrTxUnsupported
	}
	if tx.GasPrice < network.MinimalGasFee {
		return false, ErrMalformedTx
	}

	if len(tx.From) != network.AddressSize {
		return false, ErrMalformedTx
	}
	if len(tx.To) != network.AddressSize {
		return false, ErrMalformedTx
	}
	if len(tx.Signature) != ed25519.PublicKeySize+ed25519.SignatureSize {
		return false, ErrMalformedTx
	}

	account, err := getAccountData(tx.From)
	if err != nil {
		return false, err
	}

	if bytes.Compare(tx.Hash, tx.CalculateHash()) != 0 {
		return false, ErrInvalidTxHash
	}
	if tx.Nonce <= account.OutTxCounter {
		return false, ErrRejectedTx
	}

	gasCost := network.GetGasAmount(tx)
	if gasCost > tx.GasLimit {
		return false, ErrInsufficientGas
	}

	var txCost uint64 = tx.Amount + tx.Fee + gasCost*tx.GasPrice
	if account.Balance < txCost {
		return false, ErrInsufficientFunds
	}

	pubKey := tx.Signature[ed25519.SignatureSize:]
	if bytes.Compare(network.DeriveAddress(pubKey), tx.From) != 0 {
		return false, ErrMalformedTx
	}
	if !ed25519.Verify(pubKey, tx.Hash, tx.Signature[:ed25519.SignatureSize]) {
		return false, ErrMalformedTx
	}

	return true, nil
}
