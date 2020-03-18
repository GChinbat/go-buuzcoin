package trie

func (node *MerkleTrieNode) putNewChild(subkey string, value []byte) *MerkleTrieNode {
	newLeaf := new(MerkleTrieNode)
	newLeaf.Parent = node
	newLeaf.Type = hasValue
	newLeaf.Value = value

	if len(subkey) > 1 {
		newLeaf.Type |= hasExtKey
		newLeaf.ExtKey = subkey[1:]
	}

	newLeafHash := newLeaf.CalculateHash()

	node.Children[hexChar(subkey[0])] = newLeafHash

	newLeaf.Parent = node
	newLeaf.ParentKey = hexChar(subkey[0])

	return node
}

// Put adds new or modifies existing node with specific key and setting its value.
// Returns deepest updated node in copy of previous tree
func (node *MerkleTrieNode) Put(key string, value []byte, lookup LookupFn) (*MerkleTrieNode, error) {
	closestNode, remainingKeyPart, err := node.findClosest(key, lookup)
	if err != nil {
		return nil, err
	}

	if remainingKeyPart == closestNode.ExtKey {
		closestNode.Value = value
		return closestNode, nil
	}

	closestNode.Type |= hasChildren
	if closestNode.Children == nil {
		closestNode.Children = make(map[uint8][]byte)
	}

	if len(closestNode.ExtKey) == 0 {
		if len(closestNode.Children) > 0 {
			closestNode.ExtKey = remainingKeyPart
			closestNode.Value = value
			return closestNode, nil
		}

		return closestNode.putNewChild(remainingKeyPart, value), nil
	}

	// Example case: remainingKeyPart=[faba], node ExtKey=[fab]
	lcpLength := longestCommonPrefixLength(remainingKeyPart, closestNode.ExtKey)
	if lcpLength == len(closestNode.ExtKey) {
		return closestNode.putNewChild(remainingKeyPart[lcpLength:], value), nil
	}

	/*
		Split node with extended key: create new child with key remainingKeyPart[lcpLength],
		move children to splitted node
		Set node.ExtNode to node.ExtKey[:lcpLength], insert new leaf and update hashes
		Example case: remainingKeyPart=[faba], node ExtKey=[fac0]
	*/
	splittedNodeKey := hexChar(closestNode.ExtKey[lcpLength])

	splittedNode := new(MerkleTrieNode)
	splittedNode.Parent = closestNode
	splittedNode.Type = hasChildren
	if closestNode.Type&hasValue > 0 {
		splittedNode.Value = closestNode.Value
		closestNode.Type ^= hasValue
	}
	if len(closestNode.ExtKey) > lcpLength+1 {
		splittedNode.Type |= hasExtKey
		splittedNode.ExtKey = closestNode.ExtKey[lcpLength+1:]
	}
	splittedNode.Children = closestNode.Children

	closestNode.ExtKey = closestNode.ExtKey[:lcpLength]
	closestNode.Children = make(map[uint8][]byte)
	closestNode.Children[splittedNodeKey] = splittedNode.CalculateHash()

	if lcpLength == len(remainingKeyPart) {
		closestNode.Value = value
		return closestNode, nil
	}

	return closestNode.putNewChild(remainingKeyPart[lcpLength:], value), nil
}
