package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/buuzcoin/go-buuzcoin/cli/db"
)

var server = http.Server{}
var localStorage *db.LocalStorage

type jsonObject map[string]interface{}

// Close terminates API server
func Close() error {
	return server.Close()
}

// InitAPI creates HTTP JSON API listener on specific port
func InitAPI(port int, storage *db.LocalStorage) chan error {
	if localStorage != nil {
		panic("InitAPI is called twice")
	}
	localStorage = storage

	server.Addr = fmt.Sprintf("0.0.0.0:%d", port)

	// TODO: register API listeners
	http.HandleFunc("/api/v1/blockchain", GetBlockchain)
	http.HandleFunc("/api/v1/block", GetBlockData)

	resultChan := make(chan error)
	go func() {
		resultChan <- server.ListenAndServe()
	}()
	log.Printf("Started API server on port %d", port)
	return resultChan
}
