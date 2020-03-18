package consensus

import "github.com/buuzcoin/blockchain"

/*
	After block data in validated, it needs to be checked whether if block
	satisfies proof algorithm requirements. E.g: Proof-of-Work, Proof-of-Stake, etc.
*/

// ProofAlgorithm is abstract representation of block verification algorithm
type ProofAlgorithm interface {
	IsValidBlock(block blockchain.Block) (bool, error)
}
