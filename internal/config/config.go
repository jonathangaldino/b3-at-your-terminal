package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config representa a configuração do B3CLI
type Config struct {
	CurrentWallet string `yaml:"current_wallet"`
}

// configDir retorna o diretório de configuração do B3CLI
func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".b3cli"), nil
}

// configFile retorna o caminho completo do arquivo de configuração
func configFile() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load carrega a configuração do arquivo
func Load() (*Config, error) {
	filePath, err := configFile()
	if err != nil {
		return nil, err
	}

	// Se o arquivo não existir, retornar config vazia
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	// Ler arquivo
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Deserializar YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save salva a configuração no arquivo
func (c *Config) Save() error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	// Criar diretório se não existir
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	filePath, err := configFile()
	if err != nil {
		return err
	}

	// Serializar para YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Salvar arquivo
	return os.WriteFile(filePath, data, 0644)
}

// SetCurrentWallet define a wallet atual
func SetCurrentWallet(walletPath string) error {
	// Converter para caminho absoluto
	absPath, err := filepath.Abs(walletPath)
	if err != nil {
		return err
	}

	cfg := &Config{
		CurrentWallet: absPath,
	}

	return cfg.Save()
}

// GetCurrentWallet retorna o caminho da wallet atual
func GetCurrentWallet() (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	if cfg.CurrentWallet == "" {
		return "", fmt.Errorf("nenhuma wallet está aberta. Use 'b3cli wallet open <diretório>' primeiro")
	}

	return cfg.CurrentWallet, nil
}

// ClearCurrentWallet limpa a wallet atual
func ClearCurrentWallet() error {
	cfg := &Config{
		CurrentWallet: "",
	}
	return cfg.Save()
}

// HasCurrentWallet verifica se há uma wallet atual configurada
func HasCurrentWallet() bool {
	cfg, err := Load()
	if err != nil {
		return false
	}
	return cfg.CurrentWallet != ""
}
