package trie

/*
	All MerkleTries have initial state called 'null trie':
	{Type: 0x00, ExtKey: '', Value: null, Children: {}}
*/

// NullTrie is initial tree state
var NullTrie = &MerkleTrieNode{
	Type:     0x00,
	ExtKey:   "",
	Value:    []byte{},
	Children: make(map[uint8][]byte),
}

// NullTrieHash is hash of null trie root
var NullTrieHash = NullTrie.CalculateHash()
