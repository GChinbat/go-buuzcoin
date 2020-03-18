package db

import (
	"github.com/bmatsuo/lmdb-go/lmdb"
	"github.com/buuzcoin/go-buuzcoin/blockchain/trie"
)

// SaveTrie saves updated trie up to the root and returns hash of the root
func SaveTrie(node *trie.MerkleTrieNode, dbi lmdb.DBI, txn *lmdb.Txn) error {
	nodeHash := node.CalculateHash()
	if _, err := txn.Get(dbi, nodeHash); lmdb.IsNotFound(err) {
		if err = txn.Put(dbi, nodeHash, node.ToBytes(), 0); err != nil {
			return err
		}
	}

	if node.Parent == nil {
		return nil
	}
	return SaveTrie(node.Parent, dbi, txn)
}
