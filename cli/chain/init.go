package chain

import (
	"encoding/hex"
	"io/ioutil"
	"log"

	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/buuzcoin/go-buuzcoin/blockchain"
	"github.com/buuzcoin/go-buuzcoin/blockchain/trie"
	"github.com/buuzcoin/go-buuzcoin/cli/db"
	"github.com/buuzcoin/go-buuzcoin/network/consensus"
	"github.com/buuzcoin/go-buuzcoin/network/validation"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// ErrInvalidGenesisBlock is returned if genesis block validation failed
var ErrInvalidGenesisBlock = errors.New("validation: invalid genesis block")

func (dispatcher *blockchainDispatcher) loadStateTrie() error {
	dispatcher.lock.Lock()
	defer dispatcher.lock.Unlock()

	var (
		rootData []byte
		err      error
	)
	if err := dispatcher.localStorage.Env.View(func(txn *lmdb.Txn) error {
		rootData, err = txn.Get(dispatcher.localStorage.State, dispatcher.currentChain.StateMerkleRoot)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return errors.Wrap(err, "loadStateTrie: failed to get root state from local storage")
	}

	dispatcher.stateTrieRoot = new(trie.MerkleTrieNode)
	if err = dispatcher.stateTrieRoot.SetBytes(rootData); err != nil {
		return errors.Wrap(err, "loadStateTrie: failed to init root node")
	}
	return nil
}

func loadGenesisBlock(genesisBlockFile string, localStorage *db.LocalStorage,
	proofAlgo consensus.ProofAlgorithm) (*blockchain.Block, error) {
	blockBytes, err := ioutil.ReadFile(genesisBlockFile)
	if err != nil {
		return nil, errors.Wrap(err, "loadGenesisBlock: failed to read block from file")
	}

	block := new(blockchain.Block)
	if err := proto.Unmarshal(blockBytes, block); err != nil {
		return nil, errors.Wrap(err, "loadGenesisBlock: corrupt block data")
	}

	valid, err := validation.CheckGenesisBlock(*block)
	if err != nil {
		return nil, errors.Wrap(err, "loadGenesisBlock: block validation failed")
	}
	if !valid {
		return nil, ErrInvalidGenesisBlock
	}

	valid, err = proofAlgo.IsValidBlock(*block)
	if err != nil {
		return nil, errors.Wrap(err, "loadGenesisBlock: block proof-algorithm check failed")
	}
	if !valid {
		return nil, ErrInvalidGenesisBlock
	}
	log.Printf("Loaded genesis block %s\n", hex.EncodeToString(block.CalculateHash()))

	return block, nil
}

// InitBlockchainState loads current blockchain state from local storage.
// If it is not found, blockchain is initialized from genesisBlockFile
func InitBlockchainState(genesisBlockFile string, localStorage *db.LocalStorage,
	proofAlgo consensus.ProofAlgorithm) error {
	initBlockchainDispatcher(localStorage)

	genesisBlock, err := loadGenesisBlock(genesisBlockFile, localStorage, proofAlgo)
	if err != nil {
		return err
	}
	BlockchainDispatcher.genesisBlock = genesisBlock

	var chainState []byte
	if err := localStorage.Env.View(func(txn *lmdb.Txn) error {
		var err error
		chainState, err = txn.Get(localStorage.Blockchain, []byte("chainState"))
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		if lmdb.IsNotFound(err) {
			return initStateFromGenesisBlock(localStorage)
		}
		return errors.Wrap(err, "InitBlockchainState: failed to load blockchain state from database")
	}

	blockchainCurrentState := new(blockchain.Blockchain)
	if err := proto.Unmarshal(chainState, blockchainCurrentState); err != nil {
		return errors.New("InitBlockchainState: corrupt database, cannot load blockchain state")
	}
	BlockchainDispatcher.currentChain = blockchainCurrentState
	log.Printf("Loaded blockchain, last block is: %s\n", hex.EncodeToString(blockchainCurrentState.LastBlockHash))

	if err := BlockchainDispatcher.loadStateTrie(); err != nil {
		return errors.New("InitBlockchainState: could not load state trie")
	}
	log.Printf("Loaded state trie, hash is: %s\n", hex.EncodeToString(BlockchainDispatcher.stateTrieRoot.CalculateHash()))
	return nil
}

func initStateFromGenesisBlock(localStorage *db.LocalStorage) error {
	if err := BlockchainDispatcher.ApplyBlock(*BlockchainDispatcher.genesisBlock, []blockchain.TX{}); err != nil {
		return errors.Wrap(err, "initStateFromGenesisBlock: applying block failed")
	}
	log.Printf("State root: %s\n", hex.EncodeToString(BlockchainDispatcher.stateTrieRoot.CalculateHash()))
	return nil
}
