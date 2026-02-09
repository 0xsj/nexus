package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// ============================================================================
// Token Encryption
// ============================================================================

// Encryptor handles encryption and decryption of sensitive data.
type Encryptor struct {
	key []byte // 32 bytes for AES-256
}

// NewEncryptor creates a new encryptor with the given key.
// The key must be 32 bytes for AES-256.
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(key))
	}

	return &Encryptor{
		key: key,
	}, nil
}

// NewEncryptorFromString creates an encryptor from a base64-encoded key string.
func NewEncryptorFromString(keyStr string) (*Encryptor, error) {
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	return NewEncryptor(key)
}

// Encrypt encrypts plaintext using AES-256-GCM.
// Returns the encrypted ciphertext (nonce + ciphertext).
func (e *Encryptor) Encrypt(plaintext string) ([]byte, error) {
	if plaintext == "" {
		return nil, nil // Empty string returns nil
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// GCM mode provides authenticated encryption
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	// The nonce is prepended to the ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM.
// Expects the ciphertext to include the nonce (nonce + ciphertext).
func (e *Encryptor) Decrypt(ciphertext []byte) (string, error) {
	if len(ciphertext) == 0 {
		return "", nil // Empty ciphertext returns empty string
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// ============================================================================
// Key Generation
// ============================================================================

// GenerateEncryptionKey generates a random 32-byte (256-bit) encryption key.
// Returns the key as a base64-encoded string for easy storage in env vars.
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	return base64.StdEncoding.EncodeToString(key), nil
}
