package blockchain

import (
	"encoding/hex"

	"github.com/buuzcoin/go-buuzcoin/blockchain/trie"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

/*
	Account data must be saved in state trie. This file defines function
	for retrieving and saving account data from state trie
*/

// GetAccountState retrieves account state from local storage
// If it is not found, initial state will be returned
func GetAccountState(address []byte, stateRoot *trie.MerkleTrieNode, retrieve trie.LookupFn) (*AccountState, error) {
	accountData := &AccountState{
		Balance:      0,
		OutTxCounter: 0,
	}
	accountDataBytes, err := stateRoot.FindValue(hex.EncodeToString(address), retrieve)
	if err != nil {
		return nil, err
	}
	if accountDataBytes != nil {
		if err = proto.Unmarshal(accountDataBytes, accountData); err != nil {
			return nil, err
		}
	}
	return accountData, nil
}

// Save writes updated account state to trie passed. Returns updated trie leaf
func (accountState AccountState) Save(address []byte, stateRoot *trie.MerkleTrieNode, retrieve trie.LookupFn) (*trie.MerkleTrieNode, error) {
	accountStateBytes, err := proto.Marshal(&accountState)
	if err != nil {
		return nil, errors.Wrap(err, "Save: failed marshal account record")
	}
	return stateRoot.Put(hex.EncodeToString(address), accountStateBytes, retrieve)
}
