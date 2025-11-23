# Melhoria 05: Importação e Exportação Avançada

**Prioridade:** P1 (Alta)
**Complexidade:** Alta
**Impacto:** Muito Alto

---

## 📋 Visão Geral

Facilitar integração com outras plataformas através de importação automatizada do CEI (Canal Eletrônico do Investidor), suporte a múltiplas corretoras, e exportação de dados para contabilidade e backup.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Trabalho manual excessivo**
   - Baixar Excel de cada corretora todo mês
   - Converter formatos diferentes
   - Importar arquivo por arquivo

2. **Falta de integração**
   - Dados fragmentados entre corretoras
   - Impossível ter visão consolidada rápida
   - Retrabalho para cada fonte

3. **Risco de perda de dados**
   - Sem backup automático
   - Impossível restaurar se perder arquivo
   - Dados presos no formato próprio

### Benefícios mensuráveis:

- ⏱️ **Economia de tempo:** 2h/mês → 5min/mês (importação)
- 🔄 **Sincronização automática:** Dados sempre atualizados
- 💾 **Segurança:** Backups automáticos e versionados
- 🔓 **Portabilidade:** Exportar para qualquer formato

---

## 🏗️ Arquitetura Proposta

### Componentes principais:

```
internal/importer/
├── cei.go               # Integração CEI B3
├── clear.go             # Parser Clear/XP
├── rico.go              # Parser Rico
├── inter.go             # Parser Inter
├── btg.go               # Parser BTG
├── nubank.go            # Parser Nubank
├── detector.go          # Auto-detectar formato
└── normalizer.go        # Normalizar dados

internal/exporter/
├── pdf.go               # Exportar PDFs
├── csv.go               # Exportar CSV
├── json.go              # Exportar JSON
├── irpf.go              # Formato IRPF
└── backup.go            # Sistema de backup

cmd/b3cli/
├── import.go            # Comandos de importação
└── export.go            # Comandos de exportação
```

---

## 💡 Implementação Técnica

### 1. Integração CEI (Canal Eletrônico do Investidor)

**Visão geral:**
- CEI é o sistema oficial da B3
- Consolida dados de todas as corretoras
- Acesso via web scraping ou API (se disponível)

**Estrutura:**

```go
type CEIClient struct {
    Username string  // CPF
    Password string
    Session  *http.Client
}

type CEIData struct {
    Transactions []parser.Transaction
    Earnings     []parser.Earning
    Positions    []Position
    Period       DateRange
}

type Position struct {
    Ticker          string
    Quantity        int
    AveragePrice    decimal.Decimal
    CurrentPrice    decimal.Decimal
    Institution     string
}
```

**Implementação (pseudo-código):**

```go
// LoginCEI faz login no CEI
func (c *CEIClient) Login() error {
    // 1. GET https://cei.b3.com.br
    // 2. POST credenciais
    // 3. Salvar cookies de sessão
    // 4. Validar login bem-sucedido

    resp, err := c.Session.Post("https://cei.b3.com.br/login",
        url.Values{
            "username": {c.Username},
            "password": {c.Password},
        },
    )

    if err != nil {
        return fmt.Errorf("erro ao fazer login: %w", err)
    }

    // Verificar se login foi bem-sucedido
    if !strings.Contains(resp.Body, "Bem-vindo") {
        return fmt.Errorf("credenciais inválidas")
    }

    return nil
}

// FetchTransactions busca transações do CEI
func (c *CEIClient) FetchTransactions(startDate, endDate time.Time) ([]parser.Transaction, error) {
    // 1. Navegar para página de negociações
    // 2. Filtrar por data
    // 3. Fazer download ou parsear HTML
    // 4. Converter para formato interno

    url := fmt.Sprintf(
        "https://cei.b3.com.br/negociacao?dataInicio=%s&dataFim=%s",
        startDate.Format("02/01/2006"),
        endDate.Format("02/01/2006"),
    )

    resp, err := c.Session.Get(url)
    if err != nil {
        return nil, err
    }

    // Parsear HTML ou CSV retornado
    transactions := parseHTMLTransactions(resp.Body)

    return transactions, nil
}

// FetchEarnings busca proventos do CEI
func (c *CEIClient) FetchEarnings(startDate, endDate time.Time) ([]parser.Earning, error) {
    // Similar a FetchTransactions
    url := fmt.Sprintf(
        "https://cei.b3.com.br/proventos?dataInicio=%s&dataFim=%s",
        startDate.Format("02/01/2006"),
        endDate.Format("02/01/2006"),
    )

    resp, err := c.Session.Get(url)
    if err != nil {
        return nil, err
    }

    earnings := parseHTMLEarnings(resp.Body)

    return earnings, nil
}

// SyncFromCEI sincroniza wallet com dados do CEI
func (w *Wallet) SyncFromCEI(username, password string, period DateRange) (*SyncResult, error) {
    client := &CEIClient{
        Username: username,
        Password: password,
        Session:  &http.Client{},
    }

    // Login
    if err := client.Login(); err != nil {
        return nil, fmt.Errorf("erro ao conectar ao CEI: %w", err)
    }

    // Buscar transações
    transactions, err := client.FetchTransactions(period.Start, period.End)
    if err != nil {
        return nil, err
    }

    // Buscar proventos
    earnings, err := client.FetchEarnings(period.Start, period.End)
    if err != nil {
        return nil, err
    }

    // Importar na wallet
    txAdded, txDup, _ := w.AddTransactions(transactions)
    earnAdded, earnDup, _ := w.AddEarnings(earnings)

    return &SyncResult{
        TransactionsAdded:   txAdded,
        TransactionsDup:     txDup,
        EarningsAdded:       earnAdded,
        EarningsDup:         earnDup,
    }, nil
}
```

**Considerações de segurança:**
- NUNCA salvar senha em plain text
- Usar keyring do sistema operacional
- Ou pedir senha a cada sync
- Suportar 2FA se CEI exigir

### 2. Suporte Multi-Corretoras

**Auto-detecção de formato:**

```go
type BrokerFormat string

const (
    FormatB3      BrokerFormat = "b3"
    FormatClear   BrokerFormat = "clear"
    FormatRico    BrokerFormat = "rico"
    FormatInter   BrokerFormat = "inter"
    FormatBTG     BrokerFormat = "btg"
    FormatNubank  BrokerFormat = "nubank"
    FormatUnknown BrokerFormat = "unknown"
)

// DetectBrokerFormat detecta automaticamente o formato do arquivo
func DetectBrokerFormat(filePath string) (BrokerFormat, error) {
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return FormatUnknown, err
    }
    defer f.Close()

    sheets := f.GetSheetList()
    if len(sheets) == 0 {
        return FormatUnknown, fmt.Errorf("arquivo sem sheets")
    }

    // Ler primeira linha (cabeçalho)
    rows, err := f.GetRows(sheets[0])
    if err != nil {
        return FormatUnknown, err
    }

    if len(rows) == 0 {
        return FormatUnknown, fmt.Errorf("arquivo vazio")
    }

    header := rows[0]

    // Detectar por padrões no cabeçalho
    switch {
    case containsAll(header, []string{"Data do Negócio", "Código de Negociação", "Instituição"}):
        return FormatB3, nil

    case containsAll(header, []string{"Data", "Ativo", "Operação", "Corretora"}):
        return FormatClear, nil

    case containsAll(header, []string{"Data Neg.", "Ticker", "Tipo", "Quantidade"}):
        return FormatRico, nil

    case containsAll(header, []string{"Data/Hora", "Papel", "C/V", "Qtde"}):
        return FormatInter, nil

    // ... outros formatos
    }

    return FormatUnknown, nil
}

// ImportFromBroker importa de qualquer corretora
func (w *Wallet) ImportFromBroker(filePath string) (*ImportResult, error) {
    // Auto-detectar formato
    format, err := DetectBrokerFormat(filePath)
    if err != nil {
        return nil, err
    }

    if format == FormatUnknown {
        return nil, fmt.Errorf("formato não reconhecido")
    }

    // Delegar para parser específico
    var transactions []parser.Transaction
    var earnings []parser.Earning

    switch format {
    case FormatB3:
        transactions, earnings, err = parser.ParseB3File(filePath)
    case FormatClear:
        transactions, earnings, err = parser.ParseClearFile(filePath)
    case FormatRico:
        transactions, earnings, err = parser.ParseRicoFile(filePath)
    // ... outros casos
    }

    if err != nil {
        return nil, err
    }

    // Importar
    txAdded, txDup, _ := w.AddTransactions(transactions)
    earnAdded, earnDup, _ := w.AddEarnings(earnings)

    return &ImportResult{
        Format:              string(format),
        TransactionsAdded:   txAdded,
        TransactionsDup:     txDup,
        EarningsAdded:       earnAdded,
        EarningsDup:         earnDup,
    }, nil
}
```

### 3. Sistema de Backup

**Estrutura:**

```go
type Backup struct {
    ID          string
    Timestamp   time.Time
    WalletPath  string
    FilePath    string
    Size        int64
    Compressed  bool
    Hash        string  // SHA256 do arquivo
}

// CreateBackup cria backup completo da wallet
func CreateBackup(walletPath, backupDir string) (*Backup, error) {
    backup := &Backup{
        ID:         generateUUID(),
        Timestamp:  time.Now(),
        WalletPath: walletPath,
        Compressed: true,
    }

    // Nome do arquivo: wallet-2024-11-23-153045.zip
    filename := fmt.Sprintf("wallet-%s.zip", backup.Timestamp.Format("2006-01-02-150405"))
    backup.FilePath = filepath.Join(backupDir, filename)

    // Criar arquivo ZIP
    zipFile, err := os.Create(backup.FilePath)
    if err != nil {
        return nil, err
    }
    defer zipFile.Close()

    zipWriter := zip.NewWriter(zipFile)
    defer zipWriter.Close()

    // Adicionar wallet.yaml
    if err := addFileToZip(zipWriter, filepath.Join(walletPath, "wallet.yaml"), "wallet.yaml"); err != nil {
        return nil, err
    }

    // Adicionar metadata
    metadata := map[string]interface{}{
        "id":        backup.ID,
        "timestamp": backup.Timestamp,
        "version":   "1.0",
    }

    metadataBytes, _ := yaml.Marshal(metadata)
    w, _ := zipWriter.Create("metadata.yaml")
    w.Write(metadataBytes)

    // Calcular hash
    zipFile.Seek(0, 0)
    hash := sha256.New()
    io.Copy(hash, zipFile)
    backup.Hash = fmt.Sprintf("%x", hash.Sum(nil))

    // Obter tamanho
    stat, _ := zipFile.Stat()
    backup.Size = stat.Size()

    return backup, nil
}

// RestoreBackup restaura wallet de um backup
func RestoreBackup(backupPath, targetPath string) error {
    // Abrir ZIP
    r, err := zip.OpenReader(backupPath)
    if err != nil {
        return err
    }
    defer r.Close()

    // Extrair arquivos
    for _, f := range r.File {
        rc, err := f.Open()
        if err != nil {
            return err
        }

        path := filepath.Join(targetPath, f.Name)

        if f.FileInfo().IsDir() {
            os.MkdirAll(path, os.ModePerm)
        } else {
            os.MkdirAll(filepath.Dir(path), os.ModePerm)
            outFile, err := os.Create(path)
            if err != nil {
                return err
            }
            _, err = io.Copy(outFile, rc)
            outFile.Close()
            if err != nil {
                return err
            }
        }

        rc.Close()
    }

    return nil
}

// AutoBackup cria backups automáticos periódicos
func (w *Wallet) AutoBackup() error {
    backupDir := filepath.Join(w.Path, "backups")
    os.MkdirAll(backupDir, os.ModePerm)

    backup, err := CreateBackup(w.Path, backupDir)
    if err != nil {
        return err
    }

    // Limpar backups antigos (manter últimos 10)
    cleanOldBackups(backupDir, 10)

    log.Printf("Backup criado: %s (%d bytes)", backup.FilePath, backup.Size)

    return nil
}
```

### 4. Exportação para Contabilidade

**Formatos de exportação:**

```go
// ExportToIRPFFormat exporta para formato IRPF
func (w *Wallet) ExportToIRPFFormat(year int, outputPath string) error {
    report := w.GenerateAnnualIRPFReport(year)

    // Gerar CSV para importação em programas de IRPF
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Cabeçalho
    writer.Write([]string{
        "Código", "Tipo", "Discriminação", "Quantidade", "Valor Aquisição",
    })

    // Bens e Direitos
    for _, pos := range report.AssetPositions {
        code := "31" // Ações
        if pos.SubType == "fundos imobiliários" {
            code = "73" // FII
        }

        writer.Write([]string{
            code,
            pos.Ticker,
            fmt.Sprintf("%d cotas de %s", pos.Quantity, pos.Ticker),
            fmt.Sprintf("%d", pos.Quantity),
            pos.TotalInvested.StringFixed(2),
        })
    }

    return nil
}

// ExportToJSON exporta wallet completa em JSON
func (w *Wallet) ExportToJSON(outputPath string) error {
    data, err := json.MarshalIndent(w, "", "  ")
    if err != nil {
        return err
    }

    return os.WriteFile(outputPath, data, 0644)
}

// ExportToCSV exporta transações em CSV
func (w *Wallet) ExportToCSV(outputPath string) error {
    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    // Cabeçalho
    writer.Write([]string{
        "Data", "Tipo", "Ticker", "Quantidade", "Preço", "Valor", "Instituição",
    })

    // Todas as transações
    for _, asset := range w.Assets {
        for _, tx := range asset.Negotiations {
            writer.Write([]string{
                tx.Date.Format("02/01/2006"),
                tx.Type,
                tx.Ticker,
                tx.Quantity.String(),
                tx.Price.StringFixed(2),
                tx.Amount.StringFixed(2),
                tx.Institution,
            })
        }
    }

    return nil
}
```

---

## 🎨 Interface do Usuário (CLI)

### Comandos propostos:

```bash
# Importação CEI
b3cli import cei --user=12345678900
# Solicita senha de forma segura

b3cli import cei --from=2024-01-01 --to=2024-12-31

# Importação de arquivo (auto-detecta)
b3cli import broker transacoes.xlsx

# Importação forçando formato
b3cli import broker --format=clear transacoes.xlsx

# Backup
b3cli backup create ./backup/
b3cli backup list
b3cli backup restore ./backup/wallet-2024-11-23.zip

# Exportação
b3cli export irpf 2024 --output=irpf-2024.csv
b3cli export json --output=wallet-completo.json
b3cli export csv --output=transacoes.csv
b3cli export pdf --year=2024 --output=relatorio-2024.pdf
```

---

## 🔄 Fluxo de Trabalho do Usuário

### Setup inicial:
1. `b3cli import cei`
2. Informar CPF e senha
3. Importar todo histórico
4. Wallet populada automaticamente

### Mensal:
1. `b3cli import cei --from=last-month`
2. Sincronização automática
3. Pronto!

### Backup periódico (automático):
1. Após cada importação → backup automático
2. Ou: `b3cli backup create` manual

---

## 📊 Métricas de Sucesso

- ✅ 95%+ dos formatos de corretoras suportados
- ✅ Importação CEI com <1% erro
- ✅ Redução de 90% no tempo de importação
- ✅ Zero perda de dados com backups

---

## 🚧 Desafios Técnicos

1. **CEI pode mudar layout** → Testes automatizados, alertas
2. **Captcha no login** → Suporte a 2FA, manual quando necessário
3. **Rate limiting** → Respeitar limites, retry com backoff

---

**Estimativa de implementação:** 3-4 semanas (CEI é complexo)
**ROI para usuários:** Altíssimo (elimina trabalho manual repetitivo)
