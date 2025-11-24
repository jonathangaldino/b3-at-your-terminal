package crypto

// Cryptographic parameters based on Bitwarden's approach
const (
	// Argon2id parameters
	Argon2Memory      = 64 * 1024 // 64 MB
	Argon2Iterations  = 3
	Argon2Parallelism = 4
	Argon2KeyLength   = 32 // 256 bits

	// Salt and key sizes
	SaltSize          = 32 // 256 bits
	EncryptionKeySize = 32 // 256 bits for AES-256
	NonceSize         = 12 // 96 bits for GCM

	// File names
	VaultFileName        = "vault.enc"
	SaltFileName         = "salt.bin"
	EncryptedKeyFileName = "encrypted_key.bin"
	MetadataFileName     = "metadata.yaml"
)

// Metadata stores non-sensitive information about the encrypted wallet
type Metadata struct {
	Version   string `yaml:"version"`
	Algorithm string `yaml:"algorithm"`
	KDF       string `yaml:"kdf"`
}

// DefaultMetadata returns the current encryption metadata
func DefaultMetadata() Metadata {
	return Metadata{
		Version:   "1.0",
		Algorithm: "AES-256-GCM",
		KDF:       "Argon2id",
	}
}
