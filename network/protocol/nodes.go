package protocol

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"

	"github.com/buuzcoin/go-buuzcoin/network"
	"golang.org/x/crypto/sha3"
)

const (
	networkIPv4 = 0x04
	networkIPv6 = 0x06
)

// NodeRecord is structure containing node data
type NodeRecord struct {
	PublicKey []byte
	SeqID     uint32
	NodeID    []byte
	Network   byte
	Port      uint16
	IPAddress []byte
}

// FromBytes parses binary encoded NodeRecord structure.
// Returns nil if parsing failed.
func (nodeRecord *NodeRecord) FromBytes(data []byte) *NodeRecord {
	if len(data) > 60 {
		return nil
	}

	offset := 0

	if len(data) <= offset+32 {
		return nil
	}
	nodeRecord.PublicKey = data[offset : offset+32]
	offset += 32

	if len(data) <= offset+4 {
		return nil
	}
	nodeRecord.SeqID = binary.LittleEndian.Uint32(data[offset : offset+2])
	offset += 4

	if len(data) <= offset+32 {
		return nil
	}
	nodeRecord.NodeID = data[offset : offset+32]
	offset += 32

	if len(data) <= offset+1 {
		return nil
	}
	nodeRecord.Network = data[offset]
	if nodeRecord.Network != networkIPv4 &&
		nodeRecord.Network != networkIPv6 {
		return nil
	}
	offset++

	if len(data) <= offset+2 {
		return nil
	}
	nodeRecord.Port = binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2

	if nodeRecord.Network == networkIPv4 {
		if len(data) <= offset+4 {
			return nil
		}
		nodeRecord.IPAddress = data[offset : offset+4]
		offset += 4
	} else if nodeRecord.Network == networkIPv6 {
		if len(data) <= offset+16 {
			return nil
		}
		nodeRecord.IPAddress = data[offset : offset+16]
		offset += 16
	}

	return nodeRecord
}

// ToBytes encodes NodeRecord to binary format
func (nodeRecord NodeRecord) ToBytes() []byte {
	dataLen := 4 + 32 + 1 + 2
	if nodeRecord.Network == networkIPv4 {
		dataLen += 4
	} else if nodeRecord.Network == networkIPv6 {
		dataLen += 16
	}

	data := make([]byte, dataLen)
	binary.LittleEndian.PutUint32(data[0:4], nodeRecord.SeqID)
	copy(data[4:4+32], nodeRecord.NodeID)
	data[4+32] = nodeRecord.Network
	binary.LittleEndian.PutUint16(data[4+32+1:4+32+1+2], nodeRecord.Port)
	copy(data[4+32+1+2:dataLen], nodeRecord.IPAddress)

	return data
}

// Unseal verifies signature of nodeRecord and parses its data
func (sealedNodeRecord SealedNodeRecord) Unseal() *NodeRecord {
	pubKey := sealedNodeRecord.Signature[ed25519.SignatureSize:]
	hash := sha3.Sum256(sealedNodeRecord.NodeRecord)
	if !ed25519.Verify(pubKey, hash[:], sealedNodeRecord.Signature[:ed25519.SignatureSize]) {
		return nil
	}

	nodeRecord := new(NodeRecord).FromBytes(sealedNodeRecord.NodeRecord)
	if nodeRecord == nil {
		return nil
	}

	if bytes.Compare(network.DeriveAddress(pubKey), nodeRecord.NodeID) != 0 {
		return nil
	}

	return nodeRecord
}

// NodeListItem is item in nodes bucket
type NodeListItem struct {
	Key    []byte
	Record *NodeRecord

	Next *NodeListItem
	Prev *NodeListItem
}

// NodeListMaxSize is maximum size of NodeList
const NodeListMaxSize = 16

// NodeList is called k-bucket in Kademlia terminology,
// structure containing list of nodes
type NodeList struct {
	Size int
	Head *NodeListItem
	Tail *NodeListItem
}

// Remove deletes node from list
func (list *NodeList) Remove(node *NodeListItem) {
	if node == nil {
		return
	}

	if node == list.Head {
		list.Head = node.Next
	}
	if node == list.Tail {
		list.Tail = node.Prev
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	}
	if node.Next != nil {
		node.Next.Prev = node.Prev
	}
	list.Size--
}

// Lookup finds list item with specific key
func (list NodeList) Lookup(key []byte) *NodeListItem {
	node := list.Head
	for node != nil {
		if bytes.Compare(node.Key, key) == 0 {
			return node
		}
		node = node.Next
	}
	return nil
}

// Insert adds new record in front of list
func (list *NodeList) Insert(key []byte, record *NodeRecord) {
	newNode := new(NodeListItem)
	newNode.Key = key
	newNode.Record = record

	if list.Head != nil {
		// Shift previous head
		list.Head.Prev = newNode
		newNode.Next = list.Head
	}
	list.Head = newNode

	list.Size++
}

// BringToFront moves NodeListItem with specific key to front list NodeList.
// Returns false if node isn't found.
func (list *NodeList) BringToFront(key []byte) bool {
	node := list.Lookup(key)
	if node == nil {
		return false
	}

	if list.Size == 1 || node == list.Head {
		return true
	}

	// Insert node in front of list
	list.Insert(node.Key, node.Record)
	// Remove node from middle of list
	list.Remove(node)
	return true
}
