package chain

import (
	"bytes"
	"encoding/hex"
	"log"
	"sync"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/buuzcoin/blockchain"
	"github.com/buuzcoin/blockchain/trie"
	"github.com/buuzcoin/network/validation"
	"github.com/buuzcoin/node/db"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

/*
	As blockchain updates may be dispatched from multiple goroutines,
	we use single-thread update dispatcher for blockchain and transaction state.
*/

var (
	// ErrCorruptDatabase is returned if data in local storage is corrupt
	ErrCorruptDatabase = errors.New("dispatcher: Corrupt database")
	// ErrDifferentRoots is returned by ApplyBlock function if current block's hash is not
	// equal to to-be-applied block's prevHash
	ErrDifferentRoots = errors.New("dispatcher: block cannot be applied: different roots")
)

type blockchainDispatcher struct {
	localStorage  *db.LocalStorage
	genesisBlock  *blockchain.Block
	currentChain  *blockchain.Blockchain
	stateTrieRoot *trie.MerkleTrieNode

	lock *sync.RWMutex
}

// BlockchainDispatcher is object applying updates in blockchain's local state
var BlockchainDispatcher *blockchainDispatcher

// InitBlockchainDispatcher creates new instance and runs listener loop in separate goroutine
func initBlockchainDispatcher(localStorage *db.LocalStorage) {
	if BlockchainDispatcher != nil {
		panic("InitBlockchainDispatcher is called twice")
	}
	BlockchainDispatcher = &blockchainDispatcher{
		localStorage: localStorage,
		currentChain: &blockchain.Blockchain{
			LastBlockHash:   bytes.Repeat([]byte{0x00}, 32),
			StateMerkleRoot: trie.NullTrieHash,
			LastBlockIndex:  0,
		},
		stateTrieRoot: trie.NullTrie.Clone(),
		lock:          &sync.RWMutex{},
	}
}

// GetBlockchainState retrieves current state of blockchain from dispatcher
func (dispatcher *blockchainDispatcher) GetBlockchainState() blockchain.Blockchain {
	dispatcher.lock.RLock()
	defer dispatcher.lock.RUnlock()
	return *dispatcher.currentChain
}

// GetBlockchainState retrieves genesis block from dispatcher
func (dispatcher *blockchainDispatcher) GetGenesisBlock() blockchain.Block {
	dispatcher.lock.RLock()
	defer dispatcher.lock.RUnlock()
	return *dispatcher.genesisBlock
}

// GetAccountState retrieves account state for current block
func (dispatcher *blockchainDispatcher) GetAccountState(address []byte) (blockchain.AccountState, error) {
	dispatcher.lock.RLock()
	defer dispatcher.lock.RUnlock()

	var (
		accountState *blockchain.AccountState
		err          error
	)
	if err := dispatcher.localStorage.Env.View(func(txn *lmdb.Txn) error {
		retrieve := db.RetrieveFn(txn, dispatcher.localStorage.State)
		accountState, err = blockchain.GetAccountState(address, dispatcher.stateTrieRoot, retrieve)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return blockchain.AccountState{}, err
	}
	return *accountState, err
}

// ApplyBlock applies block's transactions to state trie.
func (dispatcher *blockchainDispatcher) ApplyBlock(block blockchain.Block, transactions []blockchain.TX) error {
	dispatcher.lock.Lock()
	defer dispatcher.lock.Unlock()

	if bytes.Compare(block.PrevBlockHash, dispatcher.currentChain.LastBlockHash) != 0 {
		return ErrDifferentRoots
	}

	if err := dispatcher.localStorage.Env.Update(func(txn *lmdb.Txn) error {
		retrieve := db.RetrieveFn(txn, dispatcher.localStorage.State)
		updatedChildren, isValidBlock, err := validation.ApplyBlockInMemory(dispatcher.currentChain.StateMerkleRoot, block, transactions, retrieve)
		if err != nil {
			return errors.Wrap(err, "ApplyBlock: failed to evaluate transaction in memory")
		}
		if !isValidBlock {
			return validation.ErrMalformedBlock
		}

		// Save updated trie up to root node
		for _, updatedChild := range updatedChildren {
			if err = db.SaveTrie(updatedChild, dispatcher.localStorage.State, txn); err != nil {
				return errors.Wrap(err, "applyBlock: failed to save state trie")
			}
		}

		// Update blockchain
		extendedChain := new(blockchain.Blockchain)
		extendedChain.LastBlockHash = block.CalculateHash()
		extendedChain.LastBlockIndex = block.Index
		extendedChain.StateMerkleRoot = dispatcher.stateTrieRoot.CalculateHash()

		blockchainData, err := proto.Marshal(extendedChain)
		if err != nil {
			return errors.Wrap(err, "applyBlock: failed to encode blockchain data")
		}
		if err = txn.Put(dispatcher.localStorage.Blockchain, []byte("chainState"), blockchainData, 0); err != nil {
			return errors.Wrap(err, "applyBlock: failed to save blockchain data")
		}

		if len(updatedChildren) > 0 {
			dispatcher.stateTrieRoot = updatedChildren[0].Root()
		}
		dispatcher.currentChain = extendedChain
		return nil
	}); err != nil {
		return err
	}

	log.Printf("Applied block %s", hex.EncodeToString(block.CalculateHash()))
	return dispatcher.localStorage.SaveBlock(block)
}
