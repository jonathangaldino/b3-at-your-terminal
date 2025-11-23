# Melhoria 09: Transações Avançadas

**Prioridade:** P2 (Média)
**Complexidade:** Alta
**Impacto:** Médio

---

## 📋 Visão Geral

Suporte a tipos avançados de transações: day trade, opções, BDRs, ETFs e renda fixa.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Limitação a ações e FIIs**
   - Usuário investe em BDRs mas não consegue rastrear
   - Opera opções e não tem registro
   - Day trader não consegue separar operações

2. **Cálculo de IR incorreto**
   - Day trade tem IR diferente (20%)
   - BDRs têm regras específicas
   - Opções têm tratamento fiscal próprio

3. **Visão incompleta da carteira**
   - Renda fixa ignorada
   - ETFs não separados
   - Diversificação real desconhecida

---

## 🏗️ Implementação

### 1. Day Trade

```go
type DayTrade struct {
    Date        time.Time
    Ticker      string
    BuyPrice    decimal.Decimal
    SellPrice   decimal.Decimal
    Quantity    decimal.Decimal
    Profit      decimal.Decimal  // Lucro ou prejuízo
    TaxDue      decimal.Decimal  // IR 20%
    BuyHash     string
    SellHash    string
}

// IdentifyDayTrades identifica day trades automaticamente
func (w *Wallet) IdentifyDayTrades() []DayTrade {
    dayTrades := make([]DayTrade, 0)

    // Agrupar transações por data
    byDate := w.groupTransactionsByDate()

    for date, txs := range byDate {
        // Identificar pares compra-venda do mesmo ativo no mesmo dia
        buys := filterByType(txs, "Compra")
        sells := filterByType(txs, "Venda")

        for _, buy := range buys {
            for _, sell := range sells {
                if buy.Ticker == sell.Ticker {
                    // É day trade!
                    profit := sell.Amount.Sub(buy.Amount)
                    taxDue := profit.Mul(decimal.NewFromFloat(0.20))

                    dayTrades = append(dayTrades, DayTrade{
                        Date:      date,
                        Ticker:    buy.Ticker,
                        BuyPrice:  buy.Price,
                        SellPrice: sell.Price,
                        Quantity:  buy.Quantity,
                        Profit:    profit,
                        TaxDue:    taxDue,
                        BuyHash:   buy.Hash,
                        SellHash:  sell.Hash,
                    })
                }
            }
        }
    }

    return dayTrades
}
```

### 2. Opções

```go
type Option struct {
    ID           string
    Ticker       string      // Ex: PETR290
    UnderlyingTicker string  // PETR4
    Type         OptionType  // "call" ou "put"
    Strike       decimal.Decimal
    Expiration   time.Time
    Premium      decimal.Decimal  // Prêmio pago/recebido
    Quantity     int
    Status       OptionStatus
}

type OptionType string
const (
    OptionCall OptionType = "call"
    OptionPut  OptionType = "put"
)

type OptionStatus string
const (
    OptionActive   OptionStatus = "active"
    OptionExercised OptionStatus = "exercised"
    OptionExpired   OptionStatus = "expired"
    OptionSold      OptionStatus = "sold"
)

// BuyOption registra compra de opção
func (w *Wallet) BuyOption(ticker string, optionType OptionType, strike decimal.Decimal, expiration time.Time, premium decimal.Decimal, quantity int) (*Option, error)

// SellOption registra venda de opção
func (w *Wallet) SellOption(optionID string, premium decimal.Decimal) error

// ExerciseOption exerce opção
func (w *Wallet) ExerciseOption(optionID string) error

// ExpireOptions marca opções vencidas
func (w *Wallet) ExpireOptions() []Option
```

### 3. BDRs (Brazilian Depositary Receipts)

```go
type BDR struct {
    Ticker          string  // AAPL34, MSFT34
    UnderlyingTicker string  // AAPL, MSFT
    Currency        string  // USD
    Ratio           int     // 1 BDR = X ações (geralmente 1:10)
    ExchangeRate    decimal.Decimal
}

// AddBDR registra transação de BDR
func (w *Wallet) AddBDR(ticker string, quantity decimal.Decimal, priceBRL decimal.Decimal, exchangeRate decimal.Decimal) error {
    asset, exists := w.Assets[ticker]
    if !exists {
        asset = &Asset{
            ID:      ticker,
            Type:    "renda variável",
            SubType: "BDR",
        }
        w.Assets[ticker] = asset
    }

    // Armazenar taxa de câmbio para cálculo de IR
    // BDRs têm variação cambial tributada
    // ...
}

// CalculateBDRTax calcula IR sobre BDR (15% sobre ganho + variação cambial)
func (w *Wallet) CalculateBDRTax(sale *Transaction) decimal.Decimal
```

### 4. Renda Fixa

```go
type FixedIncome struct {
    ID              string
    Type            FixedIncomeType
    Issuer          string  // Emissor (Tesouro, Banco, etc)
    IndexType       string  // "prefixado", "IPCA+", "Selic"
    Rate            decimal.Decimal
    MaturityDate    time.Time
    InvestedAmount  decimal.Decimal
    CurrentValue    decimal.Decimal
}

type FixedIncomeType string
const (
    TesouroDireto FixedIncomeType = "tesouro"
    CDB           FixedIncomeType = "cdb"
    LCI           FixedIncomeType = "lci"
    LCA           FixedIncomeType = "lca"
    Debenture     FixedIncomeType = "debenture"
)

// AddFixedIncome adiciona investimento em renda fixa
func (w *Wallet) AddFixedIncome(fiType FixedIncomeType, amount decimal.Decimal, rate decimal.Decimal, maturity time.Time) (*FixedIncome, error)

// CalculateFixedIncomeValue calcula valor atualizado
func (w *Wallet) CalculateFixedIncomeValue(fi *FixedIncome) decimal.Decimal
```

---

## 🎨 Comandos CLI

```bash
# Day Trade
b3cli trade daytrade list
b3cli trade daytrade report --month=2024-11

# Opções
b3cli options buy call PETR4 --strike=40 --expiry=2024-12-20 --premium=2.50
b3cli options list
b3cli options exercise OPTION123

# BDRs
b3cli bdr add AAPL34 100 --price=145.50 --usd=4.95
b3cli bdr overview

# Renda Fixa
b3cli fixed add cdb --amount=10000 --rate=110% --maturity=2026-01-01
b3cli fixed overview
```

---

## 📊 TUI - Opções:

```
╔══════════════════════════════════════════════════════════════════════╗
║              📈 OPÇÕES ATIVAS                                        ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Ticker     Tipo   Strike   Venc.        Prêmio    Qtd   Status     ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  PETR290    CALL   R$ 40    20/12/2024   R$ 2.50   10    Ativa     ║
║  VALE295    PUT    R$ 65    15/01/2025   R$ 3.80    5    Ativa     ║
║  ITSA150    CALL   R$ 11    20/12/2024   R$ 0.45   50    Vence 27d ║
║                                                                      ║
║  Total investido em prêmios: R$ 1,725.00                             ║
║                                                                      ║
║  [E] Exercer  [V] Vender  [q] Sair                                  ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 📊 Métricas de Sucesso

- ✅ Suporte a 95%+ dos tipos de investimentos B3
- ✅ Cálculo correto de IR para cada tipo
- ✅ Identificação automática de day trades
- ✅ Tracking completo de opções

---

**Estimativa de implementação:** 3-4 semanas
**ROI para usuários:** Médio (para investidores avançados)
