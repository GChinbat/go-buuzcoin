package db

import (
	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/buuzcoin/go-buuzcoin/blockchain/trie"
	"github.com/pkg/errors"
)

// LocalStorage represents local data storage
type LocalStorage struct {
	Env *lmdb.Env

	Node         lmdb.DBI
	State        lmdb.DBI
	Wallets      lmdb.DBI
	Blockchain   lmdb.DBI
	Transactions lmdb.DBI
}

// RetrieveFn creates trie.RetrieveFn function using dbi provided
func RetrieveFn(txn *lmdb.Txn, dbi lmdb.DBI) trie.LookupFn {
	return func(key []byte) ([]byte, error) {
		data, err := txn.Get(dbi, key)
		if err != nil {
			if lmdb.IsNotFound(err) {
				return nil, nil
			}
			return nil, err
		}
		return data, err
	}
}

// InitDB creates local database in specific path
func InitDB(path string) (*LocalStorage, error) {
	db := new(LocalStorage)

	env, err := lmdb.NewEnv()
	if err != nil {
		return nil, errors.Wrap(err, "InitDB: creating NewEnv failed")
	}
	db.Env = env

	if err = env.SetMaxDBs(5); err != nil {
		return nil, errors.Wrap(err, "InitDB: SetMaxDBs call failed")
	}
	if err = env.SetMapSize(1 << 30); err != nil {
		return nil, errors.Wrap(err, "InitDB: SetMapSize call failed")
	}
	if env.Open(path, 0, 0600); err != nil {
		return nil, errors.Wrap(err, "InitDB: opening local storage path failed")
	}

	if _, err = env.ReaderCheck(); err != nil {
		return nil, errors.Wrap(err, "InitDB: ReaderCheck failed")
	}

	if err = env.Update(func(txn *lmdb.Txn) error {
		if db.Transactions, err = txn.CreateDBI("transactions"); err != nil {
			return errors.Wrap(err, "InitDB: creating DBI 'transactions' failed")
		}
		if db.Blockchain, err = txn.CreateDBI("blockchain"); err != nil {
			return errors.Wrap(err, "InitDB: creating DBI 'blockchain' failed")
		}
		if db.Wallets, err = txn.CreateDBI("wallets"); err != nil {
			return errors.Wrap(err, "InitDB: creating DBI 'wallets' failed")
		}
		if db.State, err = txn.CreateDBI("state"); err != nil {
			return errors.Wrap(err, "InitDB: creating DBI 'state' failed")
		}
		if db.Node, err = txn.CreateDBI("node"); err != nil {
			return errors.Wrap(err, "InitDB: creating DBI 'node' failed")
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return db, nil
}
