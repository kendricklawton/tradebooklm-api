package encryption

import (
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
)

// defaultCipher is a package-level variable that holds the AES block.
// It's set once at startup by the InitEncryption function.
// This is a pragmatic way to make the cipher available to the
// Scan/Value methods, which have fixed signatures.
var defaultCipher cipher.Block

// InitEncryption initializes the encryption package with the given cipher.
// Call this from your main InitializeConfig function.
func InitEncryption(block cipher.Block) {
	if defaultCipher != nil {
		log.Println("Encryption already initialized")
		return
	}
	defaultCipher = block
}

// encrypt uses AES-GCM to encrypt plaintext.
// GCM is a modern, authenticated mode of operation.
func encrypt(plaintext []byte) ([]byte, error) {
	if defaultCipher == nil {
		return nil, fmt.Errorf("encryption cipher not initialized")
	}

	gcm, err := cipher.NewGCM(defaultCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// We need a unique nonce for each encryption
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to create nonce: %w", err)
	}

	// Seal will append the authenticated ciphertext to the nonce and return it.
	// We store nonce + ciphertext together in the DB.
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decrypt uses AES-GCM to decrypt ciphertext.
func decrypt(ciphertext []byte) ([]byte, error) {
	if defaultCipher == nil {
		return nil, fmt.Errorf("encryption cipher not initialized")
	}

	gcm, err := cipher.NewGCM(defaultCipher)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Split the nonce and the actual ciphertext
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Open will decrypt and authenticate
	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	return plaintext, nil
}
