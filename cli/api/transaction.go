package api

import (
	"net/http"
)

// GetTransactionData retrieves account state for current block.
// API endpoint: /api/v1/transaction?txn=<transaction address>
func GetTransactionData(w http.ResponseWriter, r *http.Request) {
}
