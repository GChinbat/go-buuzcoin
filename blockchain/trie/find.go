package trie

func hexChar(c byte) uint8 {
	if c >= '0' && c <= '9' {
		return c - '0'
	}
	return c - 'a' + 10
}

func (node *MerkleTrieNode) findClosest(key string, lookup LookupFn) (*MerkleTrieNode, string, error) {
	if len(key) == 0 {
		return node, "", nil
	}

	// This node has no ExtKey, do one level deeper
	if len(node.ExtKey) == 0 {
		nextNodeHash, exists := node.Children[hexChar(key[0])]
		if !exists {
			return node, key, nil
		}

		nextNodeData, err := lookup(nextNodeHash)
		if err != nil {
			return nil, "", err
		}
		if nextNodeData == nil {
			return nil, "", ErrCorruptDataSource
		}

		nextNode, err := node.createChild(hexChar(key[0]), nextNodeData)
		if err != nil {
			return nil, "", ErrCorruptDataSource
		}
		return nextNode.findClosest(key[1:], lookup)
	}

	lcpLength := longestCommonPrefixLength(key, node.ExtKey)
	if lcpLength == 0 {
		return node, key, nil
	}

	if lcpLength >= len(key) {
		return node, key, nil
	}

	nextNodeHash, exists := node.Children[hexChar(key[lcpLength])]
	if !exists {
		return node, key, nil
	}

	nextNodeData, err := lookup(nextNodeHash)
	if err != nil {
		return nil, "", err
	}
	if nextNodeData == nil {
		return nil, "", ErrCorruptDataSource
	}

	nextNode, err := node.createChild(hexChar(key[lcpLength]), nextNodeData)
	if err != nil {
		return nil, "", ErrCorruptDataSource
	}
	return nextNode.findClosest(key[lcpLength+1:], lookup)
}

// FindValue searches MerkleTrie looking for leaf with specific key. Returns nil if node is not found
func (node *MerkleTrieNode) FindValue(key string, lookup LookupFn) ([]byte, error) {
	closestNode, remainingKeyPart, err := node.findClosest(key, lookup)
	if err != nil {
		return nil, err
	}
	if remainingKeyPart != closestNode.ExtKey {
		return nil, nil
	}
	return closestNode.Value, nil
}
