package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"

	"github.com/buuzcoin/go-buuzcoin/cli/api"
	"github.com/buuzcoin/go-buuzcoin/cli/chain"
	"github.com/buuzcoin/go-buuzcoin/cli/db"
	"github.com/buuzcoin/go-buuzcoin/cli/net"
	"github.com/buuzcoin/go-buuzcoin/network/consensus"
)

func main() {
	dbpath := flag.String("db", "cli/db/", "Path for local storage")
	apiPort := flag.Int("apiPort", 14010, "Port for WebAPI")
	netPort := flag.Int("netPort", 14000, "Port for network node listener")
	ipNetwork := flag.String("ipNetwork", "IPv4", "IP network")
	genesisBlock := flag.String("genesisBlock", "cli/genesis-blocks/testnet-v1-genesis.block", "Genesis block file")
	regenerateCert := flag.Bool("regenerateCert", false, "Whether to regenerate node keys")
	// entrypoint := flag.String("entrypoint", "", "Address of network entrypoint")
	flag.Parse()

	if *ipNetwork != "IPv4" && *ipNetwork != "IPv6" {
		fmt.Fprintln(os.Stderr, "[Fatal] invalid ipNetwork value. Expected 'IPv4' or 'IPv6'")
		os.Exit(1)
	}
	var network byte
	if *ipNetwork == "IPv4" {
		network = 0x04
	}
	if *ipNetwork == "IPv6" {
		network = 0x06
	}

	localStorage, err := db.InitDB(*dbpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[Fatal] failed to initialize database: %+v\n", err)
		os.Exit(1)
	}
	defer localStorage.Env.Close()

	if err = chain.InitNullState(localStorage); err != nil {
		fmt.Fprintf(os.Stderr, "[Fatal] failed to initialize null state of account tree: %+v\n", err)
		os.Exit(1)
	}

	authorityPublicKey, _ := base64.StdEncoding.DecodeString("IaC8c+IDS+zQ6owoanDj744syG+QBEdqMNjndRFzyp0=")
	poa := &consensus.ProofOfAuthority{
		AuthorityPublicKey: authorityPublicKey,
	}

	if err = chain.InitBlockchainState(*genesisBlock, localStorage, poa); err != nil {
		fmt.Fprintf(os.Stderr, "[Fatal] failed to initialize blockchain state: %+v\n", err)
		os.Exit(1)
	}

	netNode := net.InitNode(&net.InitNodeOptions{
		Port:            *netPort,
		ForceRegenerate: *regenerateCert,
		IPNetwork:       network,
		LocalStorage:    localStorage,
		ProofAlgorithm:  poa,
	})
	defer netNode.Close()

	apiErrChan := api.InitAPI(*apiPort, localStorage)
	defer api.Close()

	panic(<-apiErrChan)
}
