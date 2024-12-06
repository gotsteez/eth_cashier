package ethcashier

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
)

type wallet struct {
	EncryptedPrivateKey string
	PublicKey           string
}

type User struct {
	ID      string
	Wallet  wallet
	Balance float64
}

func NewUser() *User {
	// Generate UUID for user ID
	userID := uuid.New().String()

	// Generate Ethereum private key
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil
	}

	// Get public key
	publicKey := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	// Convert private key to bytes and hex string
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hex.EncodeToString(privateKeyBytes)

	// Create and return new user
	return &User{
		ID: userID,
		Wallet: wallet{
			EncryptedPrivateKey: privateKeyHex,
			PublicKey:           publicKey,
		},
		Balance: 0,
	}
}

// encryptPrivateKey encrypts the private key using AES-GCM
func encryptPrivateKey(privateKey string, secretKey []byte) (string, error) {
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext := aesgcm.Seal(nil, nonce, []byte(privateKey), nil)

	// Combine nonce and ciphertext for storage
	encryptedData := append(nonce, ciphertext...)
	return hex.EncodeToString(encryptedData), nil
}

func decryptPrivateKey(encryptedKey string, secretKey []byte) (string, error) {
	// Decode the hex string back to bytes
	encryptedData, err := hex.DecodeString(encryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode hex string: %v", err)
	}

	// Extract nonce and ciphertext
	if len(encryptedData) <= 12 {
		return "", fmt.Errorf("encrypted data too short")
	}
	nonce := encryptedData[:12]
	ciphertext := encryptedData[12:]

	// Create cipher
	block, err := aes.NewCipher(secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}

	// Create GCM
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}

	// Decrypt
	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}

	return string(plaintext), nil
}
