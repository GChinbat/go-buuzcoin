package db

import (
	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/buuzcoin/go-buuzcoin/blockchain"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// SaveBlock saves block to local storage
func (storage *LocalStorage) SaveBlock(block blockchain.Block) error {
	blockHash := block.CalculateHash()
	blockData, err := proto.Marshal(&block)
	if err != nil {
		return errors.Wrap(err, "SaveBlock: block marshalling failed")
	}

	return storage.Env.Update(func(txn *lmdb.Txn) error {
		return txn.Put(storage.Blockchain, blockHash, blockData, 0)
	})
}

// GetBlock retrieves block from local storage.
// Returns nil pointer if block was not found.
func (storage *LocalStorage) GetBlock(hash []byte) (*blockchain.Block, error) {
	var blockData []byte
	if err := storage.Env.View(func(txn *lmdb.Txn) error {
		var err error
		blockData, err = txn.Get(storage.Blockchain, hash)
		return err
	}); err != nil {
		if lmdb.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	block := new(blockchain.Block)
	if err := proto.Unmarshal(blockData, block); err != nil {
		return nil, err
	}
	return block, nil
}
