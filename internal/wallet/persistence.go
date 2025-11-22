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

// Save salva a wallet em arquivos YAML separados
// - assets.yaml: contém apenas ativos ativos (quantity != 0)
// - sold-assets.yaml: contém ativos vendidos completamente (quantity == 0)
// - transactions.yaml: contém a lista de transações
func (w *Wallet) Save(dirPath string) error {
	// Criar diretório se não existir
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return err
	}

	// Salvar assets ativos e vendidos
	if err := w.saveAssets(dirPath); err != nil {
		return err
	}

	// Salvar transactions
	if err := w.saveTransactions(dirPath); err != nil {
		return err
	}

	return nil
}

// saveAssets salva os ativos em dois arquivos separados:
// - assets.yaml: apenas ativos com quantity != 0 (carteira atual)
// - sold-assets.yaml: ativos com quantity == 0 (vendidos completamente)
func (w *Wallet) saveAssets(dirPath string) error {
	// Coletar e ordenar assets por ticker
	tickers := make([]string, 0, len(w.Assets))
	for ticker := range w.Assets {
		tickers = append(tickers, ticker)
	}
	sort.Strings(tickers)

	// Separar assets ativos dos vendidos
	activeAssets := make([]AssetYAML, 0)
	soldAssets := make([]AssetYAML, 0)

	for _, ticker := range tickers {
		asset := w.Assets[ticker]
		assetYAML := AssetYAML{
			Ticker:             asset.ID,
			Type:               asset.Type,
			SubType:            asset.SubType,
			Segment:            asset.Segment,
			AveragePrice:       asset.AveragePrice.StringFixed(4),
			TotalInvestedValue: asset.TotalInvestedValue.StringFixed(4),
			Quantity:           asset.Quantity,
			IsSubscription:     asset.IsSubscription,
			SubscriptionOf:     asset.SubscriptionOf,
		}

		if asset.Quantity == 0 {
			soldAssets = append(soldAssets, assetYAML)
		} else {
			activeAssets = append(activeAssets, assetYAML)
		}
	}

	// Salvar assets ativos
	if len(activeAssets) > 0 {
		data, err := yaml.Marshal(activeAssets)
		if err != nil {
			return err
		}
		filePath := filepath.Join(dirPath, "assets.yaml")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	} else {
		// Se não houver assets ativos, remover o arquivo
		filePath := filepath.Join(dirPath, "assets.yaml")
		os.Remove(filePath)
	}

	// Salvar assets vendidos
	if len(soldAssets) > 0 {
		data, err := yaml.Marshal(soldAssets)
		if err != nil {
			return err
		}
		filePath := filepath.Join(dirPath, "sold-assets.yaml")
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return err
		}
	} else {
		// Se não houver assets vendidos, remover o arquivo
		filePath := filepath.Join(dirPath, "sold-assets.yaml")
		os.Remove(filePath)
	}

	return nil
}

// saveTransactions salva apenas as transações em transactions.yaml
func (w *Wallet) saveTransactions(dirPath string) error {
	// Ordenar transactions por data (mais antigo primeiro)
	transactions := make([]parser.Transaction, len(w.Transactions))
	copy(transactions, w.Transactions)
	sort.Slice(transactions, func(i, j int) bool {
		return transactions[i].Date.Before(transactions[j].Date)
	})

	// Converter para TransactionYAML
	transactionsYAML := make([]TransactionYAML, 0, len(transactions))
	for _, t := range transactions {
		transactionsYAML = append(transactionsYAML, TransactionYAML{
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

	// Serializar para YAML
	data, err := yaml.Marshal(transactionsYAML)
	if err != nil {
		return err
	}

	// Salvar arquivo
	filePath := filepath.Join(dirPath, "transactions.yaml")
	return os.WriteFile(filePath, data, 0644)
}

// Load carrega uma wallet dos arquivos YAML separados
func Load(dirPath string) (*Wallet, error) {
	// Tentar migrar de formato antigo se necessário
	if err := migrateOldFormat(dirPath); err != nil {
		// Se falhar, continuar tentando carregar do novo formato
	}

	// Carregar transactions
	transactions, err := loadTransactions(dirPath)
	if err != nil {
		return nil, err
	}

	// Criar Wallet a partir das transações
	// Isso automaticamente recalcula todos os campos derivados
	w := NewWallet(transactions)

	// Carregar e restaurar metadados dos assets
	if err := loadAssetsMetadata(dirPath, w); err != nil {
		// Se não conseguir carregar assets.yaml, continuar mesmo assim
		// (pode ser wallet antiga ou recém-criada)
	}

	return w, nil
}

// loadTransactions carrega as transações do arquivo transactions.yaml
func loadTransactions(dirPath string) ([]parser.Transaction, error) {
	filePath := filepath.Join(dirPath, "transactions.yaml")

	// Ler arquivo
	data, err := os.ReadFile(filePath)
	if err != nil {
		// Se não existir, retornar lista vazia
		if os.IsNotExist(err) {
			return []parser.Transaction{}, nil
		}
		return nil, err
	}

	// Deserializar YAML
	var transactionsYAML []TransactionYAML
	if err := yaml.Unmarshal(data, &transactionsYAML); err != nil {
		return nil, err
	}

	// Converter para Transaction
	transactions := make([]parser.Transaction, 0, len(transactionsYAML))
	for _, ty := range transactionsYAML {
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

	return transactions, nil
}

// loadAssetsMetadata carrega metadados dos assets de assets.yaml e sold-assets.yaml
func loadAssetsMetadata(dirPath string, w *Wallet) error {
	// Carregar assets ativos
	activePath := filepath.Join(dirPath, "assets.yaml")
	if data, err := os.ReadFile(activePath); err == nil {
		var assetsYAML []AssetYAML
		if err := yaml.Unmarshal(data, &assetsYAML); err == nil {
			for _, ay := range assetsYAML {
				if asset, exists := w.Assets[ay.Ticker]; exists {
					asset.IsSubscription = ay.IsSubscription
					asset.SubscriptionOf = ay.SubscriptionOf
					asset.SubType = ay.SubType
					asset.Segment = ay.Segment
				}
			}
		}
	}

	// Carregar assets vendidos
	soldPath := filepath.Join(dirPath, "sold-assets.yaml")
	if data, err := os.ReadFile(soldPath); err == nil {
		var assetsYAML []AssetYAML
		if err := yaml.Unmarshal(data, &assetsYAML); err == nil {
			for _, ay := range assetsYAML {
				if asset, exists := w.Assets[ay.Ticker]; exists {
					asset.IsSubscription = ay.IsSubscription
					asset.SubscriptionOf = ay.SubscriptionOf
					asset.SubType = ay.SubType
					asset.Segment = ay.Segment
				}
			}
		}
	}

	return nil
}

// Exists verifica se existe uma wallet válida no diretório
// Uma wallet é considerada válida se existe o arquivo transactions.yaml ou wallet.yaml (formato antigo)
func Exists(dirPath string) bool {
	// Verificar novo formato
	transactionsPath := filepath.Join(dirPath, "transactions.yaml")
	if _, err := os.Stat(transactionsPath); err == nil {
		return true
	}

	// Verificar formato antigo
	oldWalletPath := filepath.Join(dirPath, "wallet.yaml")
	if _, err := os.Stat(oldWalletPath); err == nil {
		return true
	}

	return false
}

// WalletFile representa a estrutura do arquivo YAML antigo (wallet.yaml)
type WalletFile struct {
	Assets       []AssetYAML       `yaml:"assets"`
	Transactions []TransactionYAML `yaml:"transactions"`
}

// migrateOldFormat migra automaticamente do formato antigo (wallet.yaml) para o novo formato
// (assets.yaml + transactions.yaml) se necessário
func migrateOldFormat(dirPath string) error {
	oldWalletPath := filepath.Join(dirPath, "wallet.yaml")
	newTransactionsPath := filepath.Join(dirPath, "transactions.yaml")

	// Verificar se já existe no novo formato
	if _, err := os.Stat(newTransactionsPath); err == nil {
		// Já está no novo formato
		return nil
	}

	// Verificar se existe wallet.yaml antigo
	if _, err := os.Stat(oldWalletPath); os.IsNotExist(err) {
		// Não existe formato antigo
		return nil
	}

	// Carregar wallet.yaml antigo
	data, err := os.ReadFile(oldWalletPath)
	if err != nil {
		return err
	}

	var walletFile WalletFile
	if err := yaml.Unmarshal(data, &walletFile); err != nil {
		return err
	}

	// Converter para transações
	transactions := make([]parser.Transaction, 0, len(walletFile.Transactions))
	for _, ty := range walletFile.Transactions {
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

	// Criar wallet a partir das transações
	w := NewWallet(transactions)

	// Restaurar metadados dos assets
	for _, ay := range walletFile.Assets {
		if asset, exists := w.Assets[ay.Ticker]; exists {
			asset.IsSubscription = ay.IsSubscription
			asset.SubscriptionOf = ay.SubscriptionOf
			asset.SubType = ay.SubType
			asset.Segment = ay.Segment
		}
	}

	// Salvar no novo formato
	if err := w.Save(dirPath); err != nil {
		return err
	}

	// Renomear wallet.yaml antigo para wallet.yaml.bak
	backupPath := filepath.Join(dirPath, "wallet.yaml.bak")
	if err := os.Rename(oldWalletPath, backupPath); err != nil {
		// Se não conseguir renomear, não é crítico
	}

	return nil
}
