package crypto

import (
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// GenerateSalt creates a cryptographically secure random salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// DeriveKey derives a cryptographic key from a password using Argon2id
// This is the master key derivation function - used to derive the key that
// decrypts the actual encryption key
func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		Argon2Iterations,
		Argon2Memory,
		Argon2Parallelism,
		Argon2KeyLength,
	)
}

// GenerateEncryptionKey creates a random encryption key
// This is the actual key used to encrypt wallet data
func GenerateEncryptionKey() ([]byte, error) {
	key := make([]byte, EncryptionKeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}
	return key, nil
}
