package validation

import (
	"github.com/buuzcoin/blockchain"
	"github.com/buuzcoin/blockchain/trie"
	"github.com/buuzcoin/network"
)

// applyTxInMemory evaluates transaction in memory and validates transaction's nonce
func applyTxInMemory(tx blockchain.TX, stateRoot *trie.MerkleTrieNode, retrieve trie.LookupFn) ([]*trie.MerkleTrieNode, uint64, bool, error) {
	if valid, err := CheckTx(tx, func(address []byte) (*blockchain.AccountState, error) {
		return blockchain.GetAccountState(address, stateRoot, retrieve)
	}); !valid || err != nil {
		return nil, 0, valid, err
	}

	recipientAccountState, err := blockchain.GetAccountState(tx.To, stateRoot, retrieve)
	if err != nil {
		return nil, 0, false, err
	}

	benefactorAccountState, err := blockchain.GetAccountState(tx.From, stateRoot, retrieve)
	if err != nil {
		return nil, 0, false, err
	}

	beneficiaryAmount := tx.Fee + network.GetGasAmount(tx)*tx.GasPrice

	if benefactorAccountState.Balance < tx.Amount+beneficiaryAmount {
		return nil, 0, false, ErrInsufficientFunds
	}

	benefactorAccountState.Balance -= tx.Amount + beneficiaryAmount
	benefactorAccountState.OutTxCounter++
	recipientAccountState.Balance += tx.Amount

	updatedChildren := make([]*trie.MerkleTrieNode, 0, 2)

	updatedChild, err := benefactorAccountState.Save(tx.From, stateRoot, retrieve)
	if err != nil {
		return nil, 0, false, err
	}
	updatedChildren = append(updatedChildren, updatedChild)

	updatedChild, err = recipientAccountState.Save(tx.To, stateRoot, retrieve)
	if err != nil {
		return nil, 0, false, err
	}
	updatedChildren = append(updatedChildren, updatedChild)

	return updatedChildren, beneficiaryAmount, true, nil
}

// ApplyBlockInMemory tries to evaluate block's transactions in memory
// Returns changed children on success, whether if block is valid or error.
// This function assumes that blocks were previously validated.
func ApplyBlockInMemory(prevStateRoot []byte, block blockchain.Block, transactions []blockchain.TX, retrieve trie.LookupFn) ([]*trie.MerkleTrieNode, bool, error) {
	var beneficiaryAmount uint64 = network.BlockReward(block.Index)
	updatedChildren := make([]*trie.MerkleTrieNode, 0, len(transactions)*2+1)

	stateRootBytes, err := retrieve(prevStateRoot)
	if err != nil {
		return nil, false, err
	}
	if stateRootBytes == nil {
		return nil, false, trie.ErrCorruptDataSource
	}
	stateRoot := new(trie.MerkleTrieNode)
	if err = stateRoot.SetBytes(stateRootBytes); err != nil {
		return nil, false, err
	}

	for _, tx := range transactions {
		txUpdatedChildren, txBeneficiaryAmount, valid, err := applyTxInMemory(tx, stateRoot, retrieve)
		if err != nil {
			return nil, false, err
		}
		if !valid {
			return nil, false, nil
		}
		updatedChildren = append(updatedChildren, txUpdatedChildren...)
		beneficiaryAmount += txBeneficiaryAmount
	}

	beneficiaryAccountState, err := blockchain.GetAccountState(block.Beneficiary, stateRoot, retrieve)
	if err != nil {
		return nil, false, err
	}

	beneficiaryAccountState.Balance += beneficiaryAmount
	updatedChild, err := beneficiaryAccountState.Save(block.Beneficiary, stateRoot, retrieve)
	if err != nil {
		return nil, false, err
	}

	updatedChildren = append(updatedChildren, updatedChild)
	return updatedChildren, true, nil
}
