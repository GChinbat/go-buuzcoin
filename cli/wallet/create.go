package wallet

import (
	"crypto/ed25519"
	"crypto/rand"

	"github.com/buuzcoin/network"
)

// GenerateWallet generates new wallet
func GenerateWallet() (*Data, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	walletAddress := network.DeriveAddress(pubKey)
	return &Data{
		Address:    walletAddress,
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}, nil
}
