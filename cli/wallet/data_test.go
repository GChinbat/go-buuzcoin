package wallet

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestWalletEncryption(t *testing.T) {
	/*
		1. Generate wallet
		2. Encrypt wallet
		3. Check EncryptedWallet format
		4. Decrypt wallet with invalid password
		5. Decrypt wallet with correct password
		5. Verify decrypted data
	*/
	wallet, err := GenerateWallet()
	if err != nil {
		t.Fatalf("GenerateWallet failed: %+v", err)
	}

	walletPassword := "testpass"
	invalidPass := "invalidpass"

	encryptedWalletData, err := wallet.EncryptWallet(walletPassword)
	if err != nil {
		t.Fatalf("EncryptWallet failed %s", hex.EncodeToString(encryptedWalletData))
	}

	if len(encryptedWalletData) < 16+12+4+20+32 {
		t.Fatalf("Invalid encryptedWalletData format: invalid length")
	}

	decryptedWalletData, err := DecryptWallet(encryptedWalletData, invalidPass)
	if err == nil {
		t.Errorf("DecryptWallet succeeded with wrong password, got private key: %s", hex.EncodeToString(decryptedWalletData.PrivateKey))
	}

	decryptedWalletData, err = DecryptWallet(encryptedWalletData, walletPassword)
	if err != nil {
		t.Fatalf("DecryptWallet failed with correct password: %+v", err)
	}

	expectedPrivatekey := wallet.PrivateKey
	if bytes.Compare(expectedPrivatekey, decryptedWalletData.PrivateKey) != 0 {
		t.Errorf("Unexpected private key: returned %s, expected %s", hex.EncodeToString(decryptedWalletData.PrivateKey), hex.EncodeToString(expectedPrivatekey))
	}

	expectedPubicKey := wallet.PublicKey
	if bytes.Compare(expectedPubicKey, decryptedWalletData.PublicKey) != 0 {
		t.Errorf("Unexpected public key: returned %s, expected %s", hex.EncodeToString(decryptedWalletData.PublicKey), hex.EncodeToString(expectedPubicKey))
	}

	expectedAddess := wallet.Address
	if bytes.Compare(expectedAddess, decryptedWalletData.Address) != 0 {
		t.Errorf("Unexpected address key: returned %s, expected %s", hex.EncodeToString(decryptedWalletData.Address), hex.EncodeToString(expectedAddess))
	}
}
