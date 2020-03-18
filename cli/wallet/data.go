package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"io"

	"golang.org/x/crypto/argon2"
)

/*
	EncryptedWalletData format:

	PasswordKDFSalt - 16 bytes
	EncryptionIV - 12 bytes
	CiphertextLength - 4 bytes, little-endian encoded number
	Ciphertext - CiphertextLength bytes
	AssociatedData - 20+32 bytes: Address || PublicKey
*/

// Data is represenation of wallet data in database
type Data struct {
	PrivateKey []byte
	PublicKey  []byte
	Address    []byte
}

// deriveKey uses password and salf provided and returns encryption key
func deriveKey(password, salt []byte) []byte {
	return argon2.IDKey(password, salt, 1, 64*1024, 4, 32)
}

// ErrInvalidPassword when wallet decryption failed
var (
	ErrInvalidPassword   = errors.New("wallet: invalid decryption password")
	ErrCorruptWalletData = errors.New("wallet: corrupt data")
)

// DecryptWallet decrypts EncryptedWalletData using AES256-GCM algorithm.
// Returns accociated wallet data
func DecryptWallet(encryptedWalletData []byte, password string) (*Data, error) {
	// We recieve binary and in EncryptedWalletData format specified above
	/*
		1. Read salt
		2. Read IV
		3. Read ciphertext length (binary.LittleEndian.Uint32)
		4. Allocate byte array called encryptedPayload and copy data from encryptedWalletData
		5. Derive encryption key using salt and password (deriveKey function)
		6. Decrypt ciphertext and get associatedData
		7. If encryption or decoding failed return one of errors defined above
	*/
	if len(encryptedWalletData) < 16+12+4 {
		return nil, ErrCorruptWalletData
	}

	salt := encryptedWalletData[:16]
	iv := encryptedWalletData[16 : 16+12]

	ciphertextLength := int(binary.LittleEndian.Uint32(encryptedWalletData[16+12 : 16+12+4]))
	if len(encryptedWalletData) != 16+12+4+ciphertextLength+20+32 {
		return nil, ErrCorruptWalletData
	}

	ciphertext := encryptedWalletData[16+12+4 : 16+12+4+ciphertextLength]

	// Generate encryption key using deriveKey function
	encryptionKey := deriveKey([]byte(password), salt)

	// Create AES block cipher instance
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create GCM cipher mode instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	associatedData := encryptedWalletData[len(encryptedWalletData)-(20+32):]
	plaintext, err := gcm.Open(nil, iv, ciphertext, associatedData)
	if err != nil {
		return nil, err
	}

	walletData := &Data{
		PrivateKey: plaintext,
		PublicKey:  associatedData[20 : 20+32],
		Address:    associatedData[:20],
	}

	return walletData, err
}

// EncryptWallet encrypts wallet using AES256-GCM algorithm.
// Returns encrypted data
func (walletData *Data) EncryptWallet(password string) ([]byte, error) {
	// Generate random salt
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}

	// Generate encryption key using deriveKey function
	encryptionKey := deriveKey([]byte(password), salt)

	// Generate random IV
	encryptionIV := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, encryptionIV); err != nil {
		return nil, err
	}

	plaintext := walletData.PrivateKey

	// Collect associatedData with wallet address and publicKey
	associatedData := make([]byte, 0, len(walletData.Address)+len(walletData.PublicKey))
	associatedData = append(associatedData, walletData.Address...)
	associatedData = append(associatedData, walletData.PublicKey...)

	// Create AES block cipher instance
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	// Create GCM cipher mode instance
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nil, encryptionIV, plaintext, associatedData)

	encryptedWalletData := make([]byte, 0, 16+12+4+len(ciphertext)+20+32)
	encryptedWalletData = append(encryptedWalletData, salt...)
	encryptedWalletData = append(encryptedWalletData, encryptionIV...)

	// Encode length of following ciphertext in little-endian format
	encryptedWalletData = append(encryptedWalletData, 0x00, 0x00, 0x00, 0x00)
	binary.LittleEndian.PutUint32(encryptedWalletData[len(encryptedWalletData)-4:], uint32(len(ciphertext)))

	encryptedWalletData = append(encryptedWalletData, ciphertext...)
	encryptedWalletData = append(encryptedWalletData, associatedData...)
	return encryptedWalletData, nil
}
