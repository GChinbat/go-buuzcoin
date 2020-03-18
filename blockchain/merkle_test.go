package blockchain

import (
	"encoding/hex"
	"testing"
)

func TestMerkleTreeEvenSize(t *testing.T) {
	leaf1, _ := hex.DecodeString("11")
	leaf2, _ := hex.DecodeString("12")

	treeRoot1 := hex.EncodeToString(CalculateMerkleRoot([][]byte{leaf1, leaf2}))
	expectedRoot1 := "214d0b0ef4f79bc1d0bd1be84d6add7218445be683f995d1e9c4c1a82a0a98ae"
	if treeRoot1 != expectedRoot1 {
		t.Errorf("Invalid merkle tree root hash: %s, expected %s", treeRoot1, expectedRoot1)
	}
}

func TestMerkleTreeOddSize(t *testing.T) {
	leaf1, _ := hex.DecodeString("11")
	leaf2, _ := hex.DecodeString("12")
	leaf3, _ := hex.DecodeString("13")

	treeRoot1 := hex.EncodeToString(CalculateMerkleRoot([][]byte{leaf1, leaf2, leaf3}))
	expectedRoot1 := "6cda0187a157d93a42617378127cab2e84d01652c66aa041b61473f60cba8766"
	if treeRoot1 != expectedRoot1 {
		t.Errorf("Invalid merkle tree root hash: %s, expected %s", treeRoot1, expectedRoot1)
	}
}
