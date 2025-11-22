package wallet

import (
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/john/b3-project/internal/parser"
	"github.com/shopspring/decimal"
	"gopkg.in/yaml.v3"
)

// WalletFile representa a estrutura do arquivo YAML da wallet
type WalletFile struct {
	Assets       []AssetYAML       `yaml:"assets"`
	Transactions []TransactionYAML `yaml:"transactions"`
}

// AssetYAML representa um ativo simplificado para serialização YAML
// Valores monetários são armazenados como strings para manter precisão decimal
// Quantity é int pois representa quantidade inteira de papéis
type AssetYAML struct {
	Ticker             string `yaml:"ticker"`
	Type               string `yaml:"type"`
	SubType            string `yaml:"subtype,omitempty"`
	Segment            string `yaml:"segment,omitempty"`
	AveragePrice       string `yaml:"average_price"`
	TotalInvestedValue string `yaml:"total_invested_value"`
	Quantity           int    `yaml:"quantity"`
	IsSubscription     bool   `yaml:"is_subscription,omitempty"`
	SubscriptionOf     string `yaml:"subscription_of,omitempty"`
}

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

// Save salva a wallet em um arquivo YAML
func (w *Wallet) Save(dirPath string) error {
	// Criar diretório se não existir
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	// Converter Wallet para WalletFile
	walletFile := w.toYAML()

	// Serializar para YAML
	data, err := yaml.Marshal(walletFile)
	if err != nil {
		return err
	}

	// Salvar arquivo
	filePath := filepath.Join(dirPath, "wallet.yaml")
	return os.WriteFile(filePath, data, 0644)
}

// Load carrega uma wallet de um arquivo YAML
func Load(dirPath string) (*Wallet, error) {
	filePath := filepath.Join(dirPath, "wallet.yaml")

	// Ler arquivo
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Deserializar YAML
	var walletFile WalletFile
	if err := yaml.Unmarshal(data, &walletFile); err != nil {
		return nil, err
	}

	// Converter WalletFile para Wallet
	return walletFile.toWallet(), nil
}

// Exists verifica se existe um arquivo wallet.yaml no diretório
func Exists(dirPath string) bool {
	filePath := filepath.Join(dirPath, "wallet.yaml")
	_, err := os.Stat(filePath)
	return err == nil
}

// toYAML converte Wallet para WalletFile (estrutura YAML)
func (w *Wallet) toYAML() WalletFile {
	wf := WalletFile{
		Assets:       make([]AssetYAML, 0, len(w.Assets)),
		Transactions: make([]TransactionYAML, 0, len(w.Transactions)),
	}

	// Converter Assets (ordenar por ticker)
	tickers := make([]string, 0, len(w.Assets))
	for ticker := range w.Assets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	for _, ticker := range tickers {
		asset := w.Assets[ticker]
		wf.Assets = append(wf.Assets, AssetYAML{
			Ticker:             asset.ID,
			Type:               asset.Type,
			SubType:            asset.SubType,
			Segment:            asset.Segment,
			AveragePrice:       asset.AveragePrice.StringFixed(4),
			TotalInvestedValue: asset.TotalInvestedValue.StringFixed(4),
			Quantity:           asset.Quantity,
			IsSubscription:     asset.IsSubscription,
			SubscriptionOf:     asset.SubscriptionOf,
		})
	}

	// Converter Transactions (ordenar por data, mais antigo primeiro)
	transactions := make([]parser.Transaction, len(w.Transactions))
	copy(transactions, w.Transactions)
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	for _, t := range transactions {
		wf.Transactions = append(wf.Transactions, TransactionYAML{
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

	return wf
}

// toWallet converte WalletFile (estrutura YAML) para Wallet
func (wf *WalletFile) toWallet() *Wallet {
	// Converter TransactionYAML para Transaction
	transactions := make([]parser.Transaction, 0, len(wf.Transactions))
	for _, ty := range wf.Transactions {
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

	// Criar Wallet a partir das transações
	// Isso automaticamente recalcula todos os campos derivados
	w := NewWallet(transactions)

	// Restaurar metadados dos assets (IsSubscription, SubscriptionOf)
	// que não são recalculados automaticamente
	for _, ay := range wf.Assets {
		if asset, exists := w.Assets[ay.Ticker]; exists {
			asset.IsSubscription = ay.IsSubscription
			asset.SubscriptionOf = ay.SubscriptionOf
		}
	}

	return w
}
