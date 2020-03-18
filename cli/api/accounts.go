package api

import (
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/buuzcoin/go-buuzcoin/cli/chain"
)

// GetAccountState retrieves account state for current block.
// API endpoint: /api/v1/account?address=<account address>
func GetAccountState(w http.ResponseWriter, r *http.Request) {
	accountAdderss := r.URL.Query().Get("address")
	accountAdderss = accountAdderss[2:]
	passAddress, err := hex.DecodeString(accountAdderss)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	accountState, err := chain.BlockchainDispatcher.GetAccountState(passAddress)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(jsonObject{
		"balance":      accountState.Balance,
		"outTxCounter": accountState.OutTxCounter,
	})
}
