package protocol

import (
	"math/bits"
)

// KademliaTable is list of nodes in Buuzcoin network
type KademliaTable struct {
	// LocalID is ID of local node
	LocalID []byte

	ActiveNodes      [160]NodeList
	ReplacementCache [160]NodeList
}

// FindBucketIndex returns index of bucket corresponding to distance between LocalID and nodeID
func (routingTable *KademliaTable) FindBucketIndex(nodeID []byte) int {
	if len(routingTable.LocalID) != 20 {
		panic("protocol: KademliaTable.LocalID was not set")
	}

	distance := 160
	for i := 0; i < 20; i++ {
		dist := bits.LeadingZeros8(routingTable.LocalID[i] ^ nodeID[i])
		distance -= dist
		if dist != 8 {
			break
		}
	}

	return distance
}

// Insert adds node record in routing table
// If bucket in ActiveNodes is full, it will be inserted
// If bucket in ReplacementCache is full, nodeRecord will be ignored
// Returns bool, whether nodeRecord was ignored
func (routingTable *KademliaTable) Insert(nodeRecord *NodeRecord) bool {
	bucketIndex := routingTable.FindBucketIndex(nodeRecord.NodeID)

	bucket := routingTable.ActiveNodes[bucketIndex]
	if bucket.Size < NodeListMaxSize {
		bucket.Insert(nodeRecord.NodeID, nodeRecord)
		return false
	}

	replaceBucket := routingTable.ReplacementCache[bucketIndex]
	// Bucket in ActiveNodes is full, check least-recently communicated nodes
	if replaceBucket.Size < NodeListMaxSize {
		replaceBucket.Insert(nodeRecord.NodeID, nodeRecord)
		return false
	}
	return true
}

// LookupNearestNodes looks for nodes closest to targetID
func (routingTable *KademliaTable) LookupNearestNodes(targetID []byte) []*NodeRecord {
	bucket := routingTable.ActiveNodes[routingTable.FindBucketIndex(targetID)]

	results := make([]*NodeRecord, 0, bucket.Size)
	node := bucket.Head
	for node != nil {
		results = append(results, node.Record)
		node = node.Next
	}

	return results
}

// PingFn is function sending Ping message and waiting and verifying Pong message
type PingFn = func(nodeRecord *NodeRecord) error

// FlushReplacementCache checks ActiveNodes whether they are active, if they are
// not, these nodes are replaced with alive nodes in ReplacementCache
// Warning: this function is blocking
func (routingTable *KademliaTable) FlushReplacementCache(ping PingFn) error {
	for i := 0; i < 160; i++ {
		currentBucket := routingTable.ReplacementCache[i]
		if currentBucket.Size == 0 {
			continue
		}

		node := currentBucket.Head
		for node != nil {
			nextNode := node.Next
			if err := ping(node.Record); err != nil {
				currentBucket.Remove(node)
			}
			node = nextNode
		}
	}

	for i := 0; i < 160; i++ {
		currentBucket := routingTable.ActiveNodes[i]
		replacementBucket := routingTable.ReplacementCache[i]
		if replacementBucket.Size == 0 {
			continue
		}

		node := currentBucket.Head
		for node != nil {
			nextNode := node.Next
			if err := ping(node.Record); err != nil {
				currentBucket.Remove(node)
				node.Key = replacementBucket.Head.Key
				node.Record = replacementBucket.Head.Record
				replacementBucket.Remove(replacementBucket.Head)
			}
			node = nextNode
		}
	}
	return nil
}
