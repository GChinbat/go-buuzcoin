package blockchain

import (
	"fmt"

	"golang.org/x/crypto/sha3"
)

const leafNodePrefix byte = 0x00
const internalNodePrefix byte = 0x01

// CalculateMerkleRoot calculates root of Merkle tree
func CalculateMerkleRoot(data [][]byte) []byte {
	if len(data) == 0 {
		hash := sha3.Sum256([]byte{})
		return hash[:]
	}

	tree := make([][]byte, len(data))

	for i, leafData := range data {
		leafDataHash := sha3.Sum256(leafData)

		tree[i] = make([]byte, 1+32)
		tree[i][0] = leafNodePrefix
		copy(tree[i][1:], leafDataHash[:])
	}

	treeSize := len(tree)
	for treeSize > 1 {
		fmt.Println(tree[:treeSize])
		for i := 0; i < treeSize; i += 2 {
			if i+1 == treeSize {
				tree[i/2] = tree[i]
				continue
			}

			hash := sha3.New256()
			hash.Write(tree[i])
			hash.Write(tree[i+1])
			tree[i/2][0] = internalNodePrefix
			copy(tree[i/2][1:], hash.Sum(nil))
		}

		treeSize = treeSize/2 + treeSize%2
		fmt.Println(tree[:treeSize])
	}

	// Omit internalNodePrefix
	return tree[0][1:]
}
