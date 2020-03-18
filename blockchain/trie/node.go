package trie

import (
	"encoding/binary"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/sha3"
)

const (
	hasChildren byte = 0b1
	hasExtKey   byte = 0b10
	hasValue    byte = 0b100
)

const (
	// TypeBranchNode is type of Merkle trie branch
	TypeBranchNode byte = hasChildren
	// TypeExtBranchNode is type of Merkle trie branch with extended key
	TypeExtBranchNode byte = hasChildren | hasExtKey
	// TypeLeafNode is type of Merkle trie leaf
	TypeLeafNode byte = hasValue
	// TypeExtLeafNode is type of Merkle trie leaf with extended key
	TypeExtLeafNode byte = hasValue | hasExtKey
)

// MerkleTrieNode representation of Merkle trie node
type MerkleTrieNode struct {
	Type byte

	// ExtKey exists only in extended nodes.
	ExtKey string
	// Value is specified only for leaf nodes
	Value []byte

	// Children is list of children hashes
	Children map[uint8][]byte

	Parent    *MerkleTrieNode
	ParentKey uint8
}

// Root returns root of MerkleTrie
func (node *MerkleTrieNode) Root() *MerkleTrieNode {
	root := node
	for root.Parent != nil {
		root = root.Parent
	}
	return root
}

// Clone performs deep-copy of MerkleTrieNode and returns new tree reference
func (node *MerkleTrieNode) Clone() *MerkleTrieNode {
	newNode := new(MerkleTrieNode)
	newNode.Type = node.Type
	newNode.Value = node.Value
	newNode.ExtKey = node.ExtKey

	newNode.Children = make(map[uint8][]byte)
	for k, v := range node.Children {
		newVal := make([]byte, len(v))
		copy(newVal, v)
		newNode.Children[k] = newVal
	}

	newNode.ParentKey = node.ParentKey
	if node.Parent != nil {
		newNode.Parent = node.Parent.Clone()
	}

	return newNode
}

/*
	MerkleTrieNode binary encoding:
	Type - 1 byte
	ExtKey - 1 byte length in nibbles + (varlength/2 + varlength%2) bytes
	Value - 4 bytes length + varlength bytes
	Children - 2 bytes bitmap + (children * 32) bytes
*/

var (
	// ErrCorruptData is returned if node's binary representation is malformed
	ErrCorruptData = errors.New("merkleTrie: corrupt node data")
)

// TODO: Add tests for all functions

func (node *MerkleTrieNode) parseChildren(data []byte, offset *int) error {
	if len(data) <= *offset+2 {
		return ErrCorruptData
	}

	bitmap := (uint16(data[*offset]) << 8) | uint16(data[*offset+1])
	*offset += 2

	node.Children = make(map[uint8][]byte)
	for i := uint8(0); i < 16; i++ {
		if (bitmap & (1 << (15 - i))) > 0 {
			if len(data) < *offset+32 {
				return ErrCorruptData
			}
			node.Children[i] = data[*offset : *offset+32]
			*offset += 32
		}
	}
	return nil
}
func (node *MerkleTrieNode) parseExtKey(data []byte, offset *int) error {
	if len(data) <= *offset {
		return ErrCorruptData
	}

	extKeyLen := data[*offset]
	*offset++

	extKeyLenInBytes := int(extKeyLen/2) + int(extKeyLen%2)
	if len(data) < *offset+extKeyLenInBytes {
		return ErrCorruptData
	}
	node.ExtKey = hex.EncodeToString(data[*offset : *offset+extKeyLenInBytes])
	if extKeyLen%2 == 1 {
		node.ExtKey = node.ExtKey[:extKeyLen]
	}
	*offset += extKeyLenInBytes
	return nil
}
func (node *MerkleTrieNode) parseValue(data []byte, offset *int) error {
	if len(data) <= *offset+4 {
		return ErrCorruptData
	}

	valueLen := binary.LittleEndian.Uint32(data[*offset : *offset+4])
	*offset += 4

	node.Value = data[*offset : *offset+int(valueLen)]
	return nil
}

// SetBytes decodes binary representation of MerkleTrieNode
func (node *MerkleTrieNode) SetBytes(data []byte) error {
	if len(data) == 0 {
		return ErrCorruptData
	}

	node.Type = data[0]
	offset := 1

	if node.Type&hasExtKey > 0 {
		if err := node.parseExtKey(data, &offset); err != nil {
			return err
		}
	}
	if node.Type&hasChildren > 0 {
		if err := node.parseChildren(data, &offset); err != nil {
			return err
		}
	}
	if node.Type&hasValue > 0 {
		if err := node.parseValue(data, &offset); err != nil {
			return err
		}
	}

	return nil
}

// ToBytes encodes MerkleTrieNode in binary format
func (node MerkleTrieNode) ToBytes() []byte {
	dataLength := 1
	if node.Type&hasExtKey > 0 {
		dataLength += 1 + len(node.ExtKey)/2 + len(node.ExtKey)%2
	}
	if node.Type&hasChildren > 0 {
		dataLength += 2 + len(node.Children)*32
	}
	if node.Type&hasValue > 0 {
		dataLength += 4 + len(node.Value)
	}

	data := make([]byte, 1, dataLength)
	data[0] = node.Type

	if node.Type&hasExtKey > 0 {
		data = append(data, byte(len(node.ExtKey)))

		var extKeyBytes []byte
		if len(node.ExtKey)%2 == 1 {
			extKeyBytes, _ = hex.DecodeString(node.ExtKey + "F")
		} else {
			extKeyBytes, _ = hex.DecodeString(node.ExtKey)
		}
		data = append(data, extKeyBytes...)
	}
	if node.Type&hasChildren > 0 {
		var (
			bitmap        uint16 = 0
			childrenBytes        = make([]byte, 0, len(node.Children)*32)
		)
		for index, childHash := range node.Children {
			bitmap |= 1 << (15 - index)
			childrenBytes = append(childrenBytes, childHash...)
		}

		data = append(data, byte(bitmap>>8), byte(bitmap&0xFF))
		data = append(data, childrenBytes...)
	}
	if node.Type&hasValue > 0 {
		data = append(data, 0x00, 0x00, 0x00, 0x00)
		binary.LittleEndian.PutUint32(data[len(data)-4:], uint32(len(node.Value)))

		data = append(data, node.Value...)
	}

	return data
}

// CalculateHash calculates hash of current node
func (node MerkleTrieNode) CalculateHash() []byte {
	hash := sha3.New256()
	hash.Write(node.ToBytes())
	return hash.Sum(nil)
}

func (node *MerkleTrieNode) createChild(parentKey uint8, data []byte) (*MerkleTrieNode, error) {
	childNode := new(MerkleTrieNode)
	childNode.Parent = node
	childNode.ParentKey = parentKey
	if err := childNode.SetBytes(data); err != nil {
		return nil, ErrCorruptDataSource
	}
	return childNode, nil
}
