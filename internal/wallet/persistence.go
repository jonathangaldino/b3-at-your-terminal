package wallet

import (
	"fmt"
	"sort"
	"time"

	"github.com/john/b3-project/internal/parser"
	wcrypto "github.com/john/b3-project/internal/wallet/crypto"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// TransactionYAML representa uma transação simplificada para serialização YAML
// Valores numéricos são armazenados como strings para manter precisão decimal
type TransactionYAML struct {
	Date        string `yaml:"date"`
	Type        string `yaml:"type"`
	Institution string `yaml:"institution"`
	Ticker      string `yaml:"ticker"`
	Quantity    string `yaml:"quantity"`
	Price       string `yaml:"price"`
	Amount      string `yaml:"amount"`
	Hash        string `yaml:"hash"`
}

// AssetYAML representa um ativo simplificado para serialização YAML
// Valores monetários são armazenados como strings para manter precisão decimal
type AssetYAML struct {
	Ticker             string          `yaml:"ticker"`
	Type               string          `yaml:"type"`
	SubType            string          `yaml:"subtype,omitempty"`
	Segment            string          `yaml:"segment,omitempty"`
	AveragePrice       string          `yaml:"average_price"`
	TotalInvestedValue string          `yaml:"total_invested_value"`
	TotalEarnings      string          `yaml:"total_earnings"`
	Quantity           int             `yaml:"quantity"`
	IsSubscription     bool            `yaml:"is_subscription,omitempty"`
	SubscriptionOf     string          `yaml:"subscription_of,omitempty"`
	Earnings           []EarningYAML   `yaml:"earnings,omitempty"`
}

// EarningYAML representa um provento simplificado para serialização YAML
// Valores numéricos são armazenados como strings para manter precisão decimal
type EarningYAML struct {
	Date        string `yaml:"date"`
	Type        string `yaml:"type"`
	Ticker      string `yaml:"ticker"`
	Quantity    string `yaml:"quantity"`
	UnitPrice   string `yaml:"unit_price"`
	TotalAmount string `yaml:"total_amount"`
	Hash        string `yaml:"hash"`
}

// VaultData representa os dados completos da wallet que serão criptografados
type VaultData struct {
	Transactions []TransactionYAML `yaml:"transactions"`
	Assets       []AssetYAML       `yaml:"assets"`
}

// Save encrypts and saves the wallet to disk
// The wallet must have an encryption key set (unlocked) to be saved
func (w *Wallet) Save(dirPath string) error {
	// Check if wallet is locked
	if w.IsLocked() {
		return fmt.Errorf("wallet is locked - cannot save without encryption key")
	}

	// Prepare vault data
	vaultData := w.prepareVaultData()

	// Convert to crypto.VaultData format
	cryptoVaultData := wcrypto.VaultData{
		Transactions: vaultData.Transactions,
		Assets:       vaultData.Assets,
	}

	// Save encrypted vault
	if err := wcrypto.SaveVault(dirPath, cryptoVaultData, w.encryptionKey); err != nil {
		return fmt.Errorf("failed to save wallet: %w", err)
	}

	// Update stored dirPath
	w.dirPath = dirPath

	return nil
}

// prepareVaultData converts wallet data to VaultData for serialization
func (w *Wallet) prepareVaultData() VaultData {
	vaultData := VaultData{
		Transactions: make([]TransactionYAML, 0, len(w.Transactions)),
		Assets:       make([]AssetYAML, 0, len(w.Assets)),
	}

	// Convert transactions
	transactions := make([]parser.Transaction, len(w.Transactions))
	copy(transactions, w.Transactions)
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	for _, t := range transactions {
		vaultData.Transactions = append(vaultData.Transactions, TransactionYAML{
			Date:        t.Date.Format("2006-01-02"),
			Type:        t.Type,
			Institution: t.Institution,
			Ticker:      t.Ticker,
			Quantity:    t.Quantity.StringFixed(4),
			Price:       t.Price.StringFixed(4),
			Amount:      t.Amount.StringFixed(4),
			Hash:        t.Hash,
		})
	}

	// Convert assets (sorted by ticker)
	tickers := make([]string, 0, len(w.Assets))
	for ticker := range w.Assets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	for _, ticker := range tickers {
		asset := w.Assets[ticker]

		// Convert earnings
		earnings := make([]EarningYAML, 0, len(asset.Earnings))
		for _, e := range asset.Earnings {
			earnings = append(earnings, EarningYAML{
				Date:        e.Date.Format("2006-01-02"),
				Type:        e.Type,
				Ticker:      e.Ticker,
				Quantity:    e.Quantity.StringFixed(4),
				UnitPrice:   e.UnitPrice.StringFixed(4),
				TotalAmount: e.TotalAmount.StringFixed(4),
				Hash:        e.Hash,
			})
		}

		assetYAML := AssetYAML{
			Ticker:             asset.ID,
			Type:               asset.Type,
			SubType:            asset.SubType,
			Segment:            asset.Segment,
			AveragePrice:       asset.AveragePrice.StringFixed(4),
			TotalInvestedValue: asset.TotalInvestedValue.StringFixed(4),
			TotalEarnings:      asset.TotalEarnings.StringFixed(4),
			Quantity:           asset.Quantity,
			IsSubscription:     asset.IsSubscription,
			SubscriptionOf:     asset.SubscriptionOf,
			Earnings:           earnings,
		}

		vaultData.Assets = append(vaultData.Assets, assetYAML)
	}

	return vaultData
}

// Create creates a new encrypted wallet with the given password
// Returns the unlocked wallet ready to use
func Create(dirPath, password string) (*Wallet, error) {
	// Initialize encrypted vault
	encryptionKey, err := wcrypto.InitializeVault(dirPath, password)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create empty wallet
	w := NewWallet([]parser.Transaction{})
	w.SetEncryptionKey(encryptionKey)
	w.SetDirPath(dirPath)

	return w, nil
}

// Load loads and decrypts a wallet from disk using the provided password
// Returns the unlocked wallet ready to use
func Load(dirPath, password string) (*Wallet, error) {
	// Unlock vault with password
	encryptionKey, err := wcrypto.UnlockVault(dirPath, password)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock wallet: %w", err)
	}

	// Load encrypted vault
	cryptoVaultData, err := wcrypto.LoadVault(dirPath, encryptionKey)
	if err != nil {
		wcrypto.ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to load wallet: %w", err)
	}

	// Convert from interface{} to proper types
	var vaultData VaultData

	// Marshal and unmarshal to convert interface{} to typed structs
	yamlBytes, err := yaml.Marshal(cryptoVaultData)
	if err != nil {
		wcrypto.ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to process vault data: %w", err)
	}

	if err := yaml.Unmarshal(yamlBytes, &vaultData); err != nil {
		wcrypto.ZeroBytes(encryptionKey)
		return nil, fmt.Errorf("failed to parse vault data: %w", err)
	}

	// Convert transactions
	transactions := make([]parser.Transaction, 0, len(vaultData.Transactions))
	for _, ty := range vaultData.Transactions {
		date, _ := time.Parse("2006-01-02", ty.Date)
		quantity, _ := decimal.NewFromString(ty.Quantity)
		price, _ := decimal.NewFromString(ty.Price)
		amount, _ := decimal.NewFromString(ty.Amount)

		transactions = append(transactions, parser.Transaction{
			Date:        date,
			Type:        ty.Type,
			Institution: ty.Institution,
			Ticker:      ty.Ticker,
			Quantity:    quantity,
			Price:       price,
			Amount:      amount,
			Hash:        ty.Hash,
		})
	}

	// Create wallet from transactions
	w := NewWallet(transactions)

	// Restore asset metadata and earnings
	for _, ay := range vaultData.Assets {
		asset, exists := w.Assets[ay.Ticker]
		if !exists {
			// Create asset if it doesn't exist (e.g., has only earnings)
			asset = &Asset{
				ID:           ay.Ticker,
				Negotiations: make([]parser.Transaction, 0),
				Earnings:     make([]parser.Earning, 0),
				Type:         ay.Type,
				SubType:      ay.SubType,
				Segment:      ay.Segment,
			}
			w.Assets[ay.Ticker] = asset
		}

		// Restore metadata
		asset.SubType = ay.SubType
		asset.Segment = ay.Segment
		asset.IsSubscription = ay.IsSubscription
		asset.SubscriptionOf = ay.SubscriptionOf

		// Restore earnings
		for _, ey := range ay.Earnings {
			date, _ := time.Parse("2006-01-02", ey.Date)
			quantity, _ := decimal.NewFromString(ey.Quantity)
			unitPrice, _ := decimal.NewFromString(ey.UnitPrice)
			totalAmount, _ := decimal.NewFromString(ey.TotalAmount)

			earning := parser.Earning{
				Date:        date,
				Type:        ey.Type,
				Ticker:      ey.Ticker,
				Quantity:    quantity,
				UnitPrice:   unitPrice,
				TotalAmount: totalAmount,
				Hash:        ey.Hash,
			}

			asset.Earnings = append(asset.Earnings, earning)
		}
	}

	// Recalculate derived fields
	w.RecalculateAssets()

	// Set encryption key and path
	w.SetEncryptionKey(encryptionKey)
	w.SetDirPath(dirPath)

	return w, nil
}

// Exists checks if an encrypted wallet exists at the given directory
func Exists(dirPath string) bool {
	return wcrypto.IsEncryptedWallet(dirPath)
}
