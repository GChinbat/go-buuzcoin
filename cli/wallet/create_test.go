package wallet

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/buuzcoin/network"
	"golang.org/x/crypto/sha3"
)

func TestGenerateWallet(t *testing.T) {
	/*
		1. Generate wallet data
		2. Check address
		3. Create data to be signed: sha3.Sum256("testdata")
		4. Sign hash created above (ed25519.Sign)
		5. Verify signature using public key (ed25519.Verify)
	*/
	wallet, err := GenerateWallet()
	if err != nil {
		t.Fatalf("GenerateWallet failed: %+v", err)
	}

	expectedAddress := network.DeriveAddress(wallet.PublicKey)
	if bytes.Compare(expectedAddress, wallet.Address) != 0 {
		t.Errorf("Unexpected address: returned %s, expected %s", hex.EncodeToString(wallet.Address), hex.EncodeToString(expectedAddress))
	}

	hash := sha3.Sum256([]byte("testdata"))
	signature := ed25519.Sign(wallet.PrivateKey, hash[:])
	if !ed25519.Verify(wallet.PublicKey, hash[:], signature) {
		t.Error("Signature verification failed")
	}
}
