package api

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/buuzcoin/go-buuzcoin/cli/chain"
)

// GetAccountState retrieves account state for current block.
// API endpoint: /api/v1/account?address=<account address>
func GetAccountState(w http.ResponseWriter, r *http.Request) {
	accountAddress := r.URL.Query().Get("address")

	// len(accountAddress) == len('0x') + len(address)*2 as address is hex-encoded
	if len(accountAddress) != 2+40 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	address, err := hex.DecodeString(accountAddress[2:])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accountState, err := chain.BlockchainDispatcher.GetAccountState(address)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("BlockchainDispatcher.GetAccountState failed: %+v\n", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonObject{
		"balance":      accountState.Balance,
		"outTxCounter": accountState.OutTxCounter,
	})
}
