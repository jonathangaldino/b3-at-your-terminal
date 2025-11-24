package crypto

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// VaultData represents the complete wallet data to be encrypted
type VaultData struct {
	Transactions interface{} `yaml:"transactions"`
	Assets       interface{} `yaml:"assets"`
}

// InitializeVault creates a new encrypted vault with the given password
// Returns the encryption key that should be kept in memory
func InitializeVault(dirPath, password string) ([]byte, error) {
	// Validate password strength
	if len(password) < 12 {
		return nil, fmt.Errorf("password must be at least 12 characters long")
	}

	// Ensure directory exists
	if err := os.MkdirAll(dirPath, 0700); err != nil {
		return nil, fmt.Errorf("failed to create wallet directory: %w", err)
	}

	// Generate salt
	salt, err := GenerateSalt()
	if err != nil {
		return nil, err
	}

	// Derive master key from password
	masterKey := DeriveKey(password, salt)
	defer ZeroBytes(masterKey)

	// Generate random encryption key
	encryptionKey, err := GenerateEncryptionKey()
	if err != nil {
		return nil, err
	}

	// Encrypt the encryption key with the master key
	encryptedKey, err := Encrypt(encryptionKey, masterKey)
	if err != nil {
		ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to encrypt encryption key: %w", err)
	}

	// Save salt
	saltPath := filepath.Join(dirPath, SaltFileName)
	if err := os.WriteFile(saltPath, salt, 0600); err != nil {
		ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to save salt: %w", err)
	}

	// Save encrypted encryption key
	keyPath := filepath.Join(dirPath, EncryptedKeyFileName)
	if err := os.WriteFile(keyPath, encryptedKey, 0600); err != nil {
		ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to save encrypted key: %w", err)
	}

	// Save metadata
	metadata := DefaultMetadata()
	metadataBytes, err := yaml.Marshal(metadata)
	if err != nil {
		ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	metadataPath := filepath.Join(dirPath, MetadataFileName)
	if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	// Create empty vault
	emptyVault := VaultData{
		Transactions: []interface{}{},
		Assets:       []interface{}{},
	}

	if err := SaveVault(dirPath, emptyVault, encryptionKey); err != nil {
		ZeroBytes(encryptionKey)
		return nil, err
	}

	return encryptionKey, nil
}

// UnlockVault decrypts the vault using the provided password
// Returns the encryption key that should be kept in memory
func UnlockVault(dirPath, password string) ([]byte, error) {
	// Load salt
	saltPath := filepath.Join(dirPath, SaltFileName)
	salt, err := os.ReadFile(saltPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load salt: %w", err)
	}

	// Derive master key from password
	masterKey := DeriveKey(password, salt)
	defer ZeroBytes(masterKey)

	// Load encrypted encryption key
	keyPath := filepath.Join(dirPath, EncryptedKeyFileName)
	encryptedKey, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load encrypted key: %w", err)
	}

	// Decrypt the encryption key
	encryptionKey, err := Decrypt(encryptedKey, masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock vault (incorrect password?): %w", err)
	}

	return encryptionKey, nil
}

// SaveVault encrypts and saves the vault data
func SaveVault(dirPath string, data VaultData, encryptionKey []byte) error {
	// Serialize to YAML
	yamlBytes, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal vault data: %w", err)
	}

	// Encrypt
	encryptedData, err := Encrypt(yamlBytes, encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault: %w", err)
	}

	// Save to file
	vaultPath := filepath.Join(dirPath, VaultFileName)
	if err := os.WriteFile(vaultPath, encryptedData, 0600); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

// LoadVault decrypts and loads the vault data
func LoadVault(dirPath string, encryptionKey []byte) (*VaultData, error) {
	// Load encrypted vault
	vaultPath := filepath.Join(dirPath, VaultFileName)
	encryptedData, err := os.ReadFile(vaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load vault: %w", err)
	}

	// Decrypt
	yamlBytes, err := Decrypt(encryptedData, encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt vault: %w", err)
	}

	// Deserialize
	var data VaultData
	if err := yaml.Unmarshal(yamlBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vault data: %w", err)
	}

	return &data, nil
}

// IsEncryptedWallet checks if a directory contains an encrypted wallet
func IsEncryptedWallet(dirPath string) bool {
	saltPath := filepath.Join(dirPath, SaltFileName)
	vaultPath := filepath.Join(dirPath, VaultFileName)

	_, saltErr := os.Stat(saltPath)
	_, vaultErr := os.Stat(vaultPath)

	return saltErr == nil && vaultErr == nil
}
