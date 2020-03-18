package api

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/buuzcoin/node/chain"
)

// GetBlockchain retrieves blockchain data from BlockchainDispatcher.
// API endpoint: /api/v1/blockchain
func GetBlockchain(w http.ResponseWriter, r *http.Request) {
	blockchain := chain.BlockchainDispatcher.GetBlockchainState()
	response := jsonObject{
		"lastBlockHash":   hex.EncodeToString(blockchain.LastBlockHash),
		"stateMerkleRoot": hex.EncodeToString(blockchain.StateMerkleRoot),
		"lastBlockIndex":  blockchain.LastBlockIndex,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetBlockData retrieves block data from local storage
// API endpoint: /api/v1/block?hash=<block hash>
// Query parameters: hash - hex-encoded hash of block
func GetBlockData(w http.ResponseWriter, r *http.Request) {
	hashString := r.URL.Query().Get("hash")
	decodeHash, err := hex.DecodeString(hashString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	block, err := localStorage.GetBlock(decodeHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("GetBlock failed: %+v", err)
		return
	}
	if block == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)

	txHashes := make([]string, len(block.TxHashes))
	for i, txHash := range block.TxHashes {
		txHashes[i] = hex.EncodeToString(txHash)
	}

	json.NewEncoder(w).Encode(jsonObject{
		"version":         block.Version,
		"index":           block.Index,
		"timestamp":       block.Timestamp,
		"prevBlockHash":   hex.EncodeToString(block.PrevBlockHash),
		"txMerkleRoot":    hex.EncodeToString(block.TxMerkleRoot),
		"stateMerkleRoot": hex.EncodeToString(block.StateMerkleRoot),
		"beneficiary":     "0x" + hex.EncodeToString(block.Beneficiary),
		"additionalData":  hex.EncodeToString(block.AdditionalData),
		"proofData":       hex.EncodeToString(block.ProofData),
		"signature":       hex.EncodeToString(block.Signature),
		"txHashes":        txHashes,
	})
}
