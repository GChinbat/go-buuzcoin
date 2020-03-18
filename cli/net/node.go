package net

import (
	"bytes"
	"crypto/ed25519"
	"crypto/tls"
	"fmt"
	"os"
	"sync"

	"github.com/buuzcoin/go-buuzcoin/cli/db"
	"github.com/buuzcoin/go-buuzcoin/network/consensus"
	"github.com/buuzcoin/go-buuzcoin/network/protocol"
	"golang.org/x/crypto/sha3"
)

// NetworkNode is responsible for managing incoming and outgoing connections
type NetworkNode struct {
	done         chan interface{}
	tlsConfig    *tls.Config
	localStorage *db.LocalStorage

	nodeAddress             []byte
	nodePubKey, nodePrivKey []byte

	nodeRecord       *protocol.NodeRecord
	nodeRecordLock   *sync.RWMutex
	sealedNodeRecord *protocol.SealedNodeRecord
}

// InitNodeOptions are options passed to InitNode function
type InitNodeOptions struct {
	Port int
	// Whether to regenerate TLS certificates
	ForceRegenerate bool
	IPNetwork       byte
	LocalStorage    *db.LocalStorage
	ProofAlgorithm  consensus.ProofAlgorithm
}

// InitNode initializes node and returns new ConnectionnetNode instance
func InitNode(options *InitNodeOptions) *NetworkNode {
	netNode := &NetworkNode{
		done:           make(chan interface{}),
		localStorage:   options.LocalStorage,
		nodeRecordLock: &sync.RWMutex{},
	}

	address := fmt.Sprintf("0.0.0.0:%d", options.Port)
	if err := netNode.LoadNodeKeys(options.ForceRegenerate); err != nil {
		fmt.Fprintf(os.Stderr, "[fatal] Failed to load node keys: %+v\n", err)
		os.Exit(1)
	}

	netNode.CreateInitialNodeRecord(options.IPNetwork)
	if err := netNode.LoadTLSConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "[fatal] Failed to init listener on %s: %+v\n", address, err)
		os.Exit(1)
	}
	go func() {
		if err := netNode.Listen(address); err != nil {
			fmt.Fprintf(os.Stderr, "[fatal] Failed to init listener on %s: %s", address, err)
			os.Exit(1)
		}
	}()
	return netNode
}

// Close terminates all connections and exits all listeners
func (netNode *NetworkNode) Close() {
	close(netNode.done)
}

// CreateInitialNodeRecord initializes NodeRecord structure
func (netNode *NetworkNode) CreateInitialNodeRecord(ipNetwork byte) {
	netNode.nodeRecordLock.Lock()

	netNode.nodeRecord = new(protocol.NodeRecord)
	netNode.nodeRecord.Port = 0
	netNode.nodeRecord.SeqID = 0
	netNode.nodeRecord.NodeID = netNode.nodeAddress
	netNode.nodeRecord.Network = ipNetwork
	netNode.nodeRecord.IPAddress = bytes.Repeat([]byte{0x00}, int(ipNetwork))
	netNode.nodeRecord.PublicKey = netNode.nodePubKey

	netNode.nodeRecordLock.Unlock()

	netNode.sealedNodeRecord = new(protocol.SealedNodeRecord)
	netNode.SealNodeRecord()
}

// SealNodeRecord signs node record and writes resulting data to NetworkNode structure
func (netNode *NetworkNode) SealNodeRecord() {
	netNode.nodeRecordLock.Lock()
	defer netNode.nodeRecordLock.Unlock()

	netNode.sealedNodeRecord.NodeRecord = netNode.nodeRecord.ToBytes()
	hash := sha3.Sum256(netNode.sealedNodeRecord.NodeRecord)
	netNode.sealedNodeRecord.Signature = ed25519.Sign(netNode.nodePrivKey, hash[:])
}
