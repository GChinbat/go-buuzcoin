package trie

import "errors"

var (
	// ErrCorruptDataSource is returned if fetching data with specific key from DataSource haven't returned anything
	ErrCorruptDataSource = errors.New("merkleTrie: corrupt data source")
)

// LookupFn is function for retrieving MerkleTrie node data
type LookupFn = func(key []byte) ([]byte, error)
