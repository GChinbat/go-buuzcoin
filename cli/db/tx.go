package db

import (
	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/buuzcoin/blockchain"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// SaveTx saves transaction data to local storage
func (storage *LocalStorage) SaveTx(tx blockchain.TX) error {
	txHash := tx.CalculateHash()
	txData, err := proto.Marshal(&tx)
	if err != nil {
		return errors.Wrap(err, "SaveTx: tx marshalling failed")
	}

	return storage.Env.Update(func(txn *lmdb.Txn) error {
		return txn.Put(storage.Transactions, txHash, txData, 0)
	})
}

// GetTX retrieves transaction from local storage.
// Returns nil pointer if tx was not found.
func (storage *LocalStorage) GetTX(hash []byte) (*blockchain.TX, error) {
	var txData []byte
	if err := storage.Env.View(func(txn *lmdb.Txn) error {
		var err error
		txData, err = txn.Get(storage.Transactions, hash)
		return err
	}); err != nil {
		if lmdb.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	tx := new(blockchain.TX)
	if err := proto.Unmarshal(txData, tx); err != nil {
		return nil, err
	}
	return tx, nil
}
