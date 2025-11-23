# Melhoria 10: Multi-Corretora e Consolidação

**Prioridade:** P3 (Baixa)
**Complexidade:** Média
**Impacto:** Médio

---

## 📋 Visão Geral

Gerenciar investimentos em múltiplas corretoras com tracking de custos, taxas e consolidação automática de posições.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Posições fragmentadas**
   - Mesma ação em 3 corretoras diferentes
   - Impossível ter visão unificada
   - Difícil saber posição real total

2. **Custos ocultos**
   - Não sabe quanto paga de taxa em cada corretora
   - Comparação impossível
   - Pode estar pagando mais que deveria

3. **Transferências não rastreadas**
   - Moveu ações entre corretoras
   - Perdeu histórico
   - Preço médio bagunçado

---

## 🏗️ Implementação

```go
type Broker struct {
    ID              string
    Name            string  // "Clear", "Rico", "Inter"
    FeeStructure    FeeStructure
}

type FeeStructure struct {
    TradeFee        decimal.Decimal  // Taxa de corretagem
    CustodyFee      decimal.Decimal  // Custódia mensal
    ISSPercentage   decimal.Decimal  // ISS sobre corretagem
}

type BrokerPosition struct {
    Ticker     string
    Broker     string
    Quantity   int
    AvgPrice   decimal.Decimal
}

// AddTransactionWithBroker registra transação com corretora
func (w *Wallet) AddTransactionWithBroker(tx *Transaction, brokerID string) error

// GetPositionsByBroker retorna posições por corretora
func (w *Wallet) GetPositionsByBroker(brokerID string) []BrokerPosition

// ConsolidatePositions consolida posições de todas as corretoras
func (w *Wallet) ConsolidatePositions() map[string]ConsolidatedPosition

// TransferCustody transfere ativos entre corretoras
func (w *Wallet) TransferCustody(ticker string, quantity int, fromBroker, toBroker string, date time.Time) error

// CalculateBrokerCosts calcula custos por corretora
func (w *Wallet) CalculateBrokerCosts(brokerID string, year int) *BrokerCosts

type BrokerCosts struct {
    TradeFees    decimal.Decimal
    CustodyFees  decimal.Decimal
    ISSFees      decimal.Decimal
    B3Fees       decimal.Decimal
    TotalCosts   decimal.Decimal
}
```

---

## 🎨 Comandos CLI

```bash
# Gerenciar corretoras
b3cli brokers add clear --trade-fee=0 --custody-fee=0
b3cli brokers list

# Ver posições por corretora
b3cli brokers positions --broker=clear
b3cli brokers overview

# Transferir custódia
b3cli brokers transfer ITSA4 100 --from=clear --to=rico --date=2024-11-20

# Análise de custos
b3cli brokers costs --year=2024
b3cli brokers compare
```

---

## 📊 TUI - Comparação de Custos:

```
╔══════════════════════════════════════════════════════════════════════╗
║              💰 CUSTOS POR CORRETORA - 2024                          ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Corretora   Corretagem  Custódia   ISS     B3      Total           ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  Clear       R$    0     R$    0    R$  15  R$ 120  R$  135  ⭐     ║
║  Rico        R$  150     R$   60    R$  25  R$ 120  R$  355         ║
║  Inter       R$  120     R$  240    R$  20  R$ 120  R$  500         ║
║                                                                      ║
║  TOTAL       R$  270     R$  300    R$  60  R$ 360  R$  990         ║
║                                                                      ║
║  💡 Sugestão: Consolidar operações na Clear (R$ 0 corretagem)       ║
║                                                                      ║
║  [C] Consolidar  [q] Sair                                           ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 📊 Valor para o Usuário

- 💰 **Economia:** Identificar corretora mais barata
- 📊 **Visão única:** Consolidação de posições
- 🔄 **Histórico:** Rastrear transferências
- 💡 **Otimização:** Sugestões de consolidação

---

**Estimativa de implementação:** 1-2 semanas
**ROI para usuários:** Médio (para quem usa múltiplas corretoras)
