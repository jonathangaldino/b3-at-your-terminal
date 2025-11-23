# Melhoria 03: Eventos Corporativos (Corporate Events)

**Prioridade:** P1 (Alta)
**Complexidade:** Média-Alta
**Impacto:** Alto

---

## 📋 Visão Geral

Suporte completo a eventos corporativos que afetam quantidade, preço médio e estrutura dos ativos: desdobramentos, grupamentos, bonificações, direitos de subscrição, fusões e aquisições.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Cálculos manuais complexos**
   - Desdobramento 1:3 → usuário tem que recalcular tudo manualmente
   - Bonificação → como ajustar o preço médio?
   - Fusão de empresas → perda de histórico

2. **Dados incorretos na carteira**
   - Quantidade de ações errada após split
   - Preço médio inflado/deflado
   - Prejuízo no cálculo de IR

3. **Falta de rastreabilidade**
   - "Por que meu preço médio mudou?"
   - Impossível auditar histórico
   - Difícil explicar para a Receita Federal

### Benefícios mensuráveis:

- ✅ **Precisão 100%:** Elimina erros de cálculo manual
- ⏱️ **Economia de tempo:** 30min-2h por evento → automático
- 📊 **Histórico completo:** Auditável e rastreável
- 💰 **IR correto:** Evita problemas com Receita Federal

---

## 🏗️ Arquitetura Proposta

### Componentes principais:

```
internal/events/
├── split.go             # Desdobramento
├── merge.go             # Grupamento
├── bonus.go             # Bonificação
├── subscription.go      # Subscrição (já parcialmente implementado)
├── merger.go            # Fusões e aquisições
├── history.go           # Histórico de eventos
└── validators.go        # Validações

cmd/b3cli/
├── events.go            # Comando principal
├── events_split.go
├── events_merge.go
├── events_bonus.go
└── events_history.go
```

---

## 💡 Implementação Técnica

### 1. Desdobramento (Stock Split)

**Conceito:**
- Empresa aumenta o número de ações dividindo cada ação em N partes
- Ex: Split 1:2 → cada ação vira 2, preço cai pela metade
- Quantidade × N, Preço ÷ N

**Estrutura de dados:**

```go
type SplitEvent struct {
    ID          string          // UUID do evento
    Ticker      string          // Ativo afetado
    Date        time.Time       // Data do evento
    Ratio       SplitRatio      // Proporção (1:2, 1:3, etc)
    Description string          // "Desdobramento 1:2"

    // Estado antes
    QuantityBefore int
    PriceBefore    decimal.Decimal

    // Estado depois
    QuantityAfter  int
    PriceAfter     decimal.Decimal

    CreatedAt   time.Time
}

type SplitRatio struct {
    From int  // 1
    To   int  // 2, 3, 4...
}
```

**Implementação:**

```go
// ApplySplit aplica um desdobramento a um ativo
func (w *Wallet) ApplySplit(ticker string, ratio SplitRatio, eventDate time.Time) (*SplitEvent, error) {
    // Validar
    asset, exists := w.Assets[ticker]
    if !exists {
        return nil, fmt.Errorf("ativo %s não encontrado", ticker)
    }

    if ratio.From != 1 || ratio.To < 2 {
        return nil, fmt.Errorf("proporção inválida: %d:%d", ratio.From, ratio.To)
    }

    // Criar evento
    event := &SplitEvent{
        ID:             generateUUID(),
        Ticker:         ticker,
        Date:           eventDate,
        Ratio:          ratio,
        Description:    fmt.Sprintf("Desdobramento %d:%d", ratio.From, ratio.To),
        QuantityBefore: asset.Quantity,
        PriceBefore:    asset.AveragePrice,
        CreatedAt:      time.Now(),
    }

    // Aplicar split
    multiplier := decimal.NewFromInt(int64(ratio.To)).Div(decimal.NewFromInt(int64(ratio.From)))

    // Ajustar transações históricas (ANTES da data do evento)
    for i := range asset.Negotiations {
        if asset.Negotiations[i].Date.Before(eventDate) {
            // Quantidade × multiplier
            asset.Negotiations[i].Quantity = asset.Negotiations[i].Quantity.Mul(multiplier)

            // Preço ÷ multiplier
            asset.Negotiations[i].Price = asset.Negotiations[i].Price.Div(multiplier)

            // Amount permanece igual (Quantidade × Preço)
        }
    }

    // Recalcular ativo
    w.RecalculateAsset(asset)

    event.QuantityAfter = asset.Quantity
    event.PriceAfter = asset.AveragePrice

    // Adicionar ao histórico
    w.EventHistory = append(w.EventHistory, event)

    return event, nil
}
```

### 2. Grupamento (Reverse Split)

**Conceito:**
- Empresa reduz o número de ações agrupando N ações em 1
- Ex: Grupamento 10:1 → cada 10 ações viram 1, preço multiplica por 10
- Quantidade ÷ N, Preço × N

**Implementação:**

```go
type MergeEvent struct {
    ID          string
    Ticker      string
    Date        time.Time
    Ratio       MergeRatio      // Ex: 10:1, 5:1
    Description string

    QuantityBefore int
    PriceBefore    decimal.Decimal
    QuantityAfter  int
    PriceAfter     decimal.Decimal

    CreatedAt   time.Time
}

type MergeRatio struct {
    From int  // 10, 5, 2...
    To   int  // 1
}

// ApplyMerge aplica um grupamento a um ativo
func (w *Wallet) ApplyMerge(ticker string, ratio MergeRatio, eventDate time.Time) (*MergeEvent, error) {
    asset, exists := w.Assets[ticker]
    if !exists {
        return nil, fmt.Errorf("ativo %s não encontrado", ticker)
    }

    if ratio.To != 1 || ratio.From < 2 {
        return nil, fmt.Errorf("proporção inválida: %d:%d", ratio.From, ratio.To)
    }

    event := &MergeEvent{
        ID:             generateUUID(),
        Ticker:         ticker,
        Date:           eventDate,
        Ratio:          ratio,
        Description:    fmt.Sprintf("Grupamento %d:%d", ratio.From, ratio.To),
        QuantityBefore: asset.Quantity,
        PriceBefore:    asset.AveragePrice,
        CreatedAt:      time.Now(),
    }

    // Divisor (ex: 10:1 = divisor 10)
    divisor := decimal.NewFromInt(int64(ratio.From)).Div(decimal.NewFromInt(int64(ratio.To)))

    // Ajustar transações históricas (ANTES da data do evento)
    for i := range asset.Negotiations {
        if asset.Negotiations[i].Date.Before(eventDate) {
            // Quantidade ÷ divisor
            asset.Negotiations[i].Quantity = asset.Negotiations[i].Quantity.Div(divisor)

            // Preço × divisor
            asset.Negotiations[i].Price = asset.Negotiations[i].Price.Mul(divisor)
        }
    }

    // Recalcular
    w.RecalculateAsset(asset)

    event.QuantityAfter = asset.Quantity
    event.PriceAfter = asset.AveragePrice

    w.EventHistory = append(w.EventHistory, event)

    return event, nil
}
```

### 3. Bonificação

**Conceito:**
- Empresa distribui ações gratuitas aos acionistas
- Ex: Bonificação de 10% → quem tem 100 ações recebe mais 10
- Quantidade aumenta, preço médio diminui (diluição)

**Estrutura:**

```go
type BonusEvent struct {
    ID           string
    Ticker       string
    Date         time.Time
    Percentage   decimal.Decimal  // 10%, 20%, 50%...
    Description  string

    QuantityBefore int
    BonusShares    int              // Ações recebidas
    QuantityAfter  int

    PriceBefore  decimal.Decimal
    PriceAfter   decimal.Decimal    // Ajustado pela diluição

    CreatedAt    time.Time
}

// ApplyBonus aplica uma bonificação a um ativo
func (w *Wallet) ApplyBonus(ticker string, percentage decimal.Decimal, eventDate time.Time) (*BonusEvent, error) {
    asset, exists := w.Assets[ticker]
    if !exists {
        return nil, fmt.Errorf("ativo %s não encontrado", ticker)
    }

    if percentage.LessThanOrEqual(decimal.Zero) {
        return nil, fmt.Errorf("percentual deve ser maior que zero")
    }

    event := &BonusEvent{
        ID:             generateUUID(),
        Ticker:         ticker,
        Date:           eventDate,
        Percentage:     percentage,
        Description:    fmt.Sprintf("Bonificação de %s%%", percentage.String()),
        QuantityBefore: asset.Quantity,
        PriceBefore:    asset.AveragePrice,
        CreatedAt:      time.Now(),
    }

    // Calcular ações bonificadas
    bonusShares := percentage.Div(decimal.NewFromInt(100)).Mul(decimal.NewFromInt(int64(asset.Quantity)))
    event.BonusShares = int(bonusShares.IntPart())

    // Criar transação de bonificação (compra com preço zero)
    bonusTx := parser.Transaction{
        Date:        eventDate,
        Type:        "Bonificação",
        Institution: "Evento Corporativo",
        Ticker:      ticker,
        Quantity:    bonusShares,
        Price:       decimal.Zero,
        Amount:      decimal.Zero,
        Hash:        generateTransactionHash(),
    }

    // Adicionar transação
    asset.Negotiations = append(asset.Negotiations, bonusTx)

    // Recalcular (preço médio será ajustado automaticamente)
    w.RecalculateAsset(asset)

    event.QuantityAfter = asset.Quantity
    event.PriceAfter = asset.AveragePrice

    w.EventHistory = append(w.EventHistory, event)

    return event, nil
}
```

### 4. Fusões e Aquisições

**Conceito:**
- Empresa A é adquirida/fundida com empresa B
- Ações de A são convertidas em ações de B
- Ex: LAME3 → LAME4 (conversão 1:1)

**Implementação:**

```go
type AcquisitionEvent struct {
    ID              string
    FromTicker      string          // LAME3
    ToTicker        string          // LAME4
    Date            time.Time
    Ratio           ConversionRatio // 1:1, 2:1, etc
    Description     string

    TransactionsMoved int
    EarningsMoved     int

    CreatedAt       time.Time
}

type ConversionRatio struct {
    From int  // Ações antigas
    To   int  // Ações novas
}

// ApplyAcquisition converte ativo antigo em novo
func (w *Wallet) ApplyAcquisition(fromTicker, toTicker string, ratio ConversionRatio, eventDate time.Time) (*AcquisitionEvent, error) {
    // Validar que fromTicker existe
    fromAsset, exists := w.Assets[fromTicker]
    if !exists {
        return nil, fmt.Errorf("ativo origem %s não encontrado", fromTicker)
    }

    event := &AcquisitionEvent{
        ID:          generateUUID(),
        FromTicker:  fromTicker,
        ToTicker:    toTicker,
        Date:        eventDate,
        Ratio:       ratio,
        Description: fmt.Sprintf("Conversão %s → %s (%d:%d)", fromTicker, toTicker, ratio.From, ratio.To),
        CreatedAt:   time.Now(),
    }

    // Criar ou obter ativo de destino
    toAsset, exists := w.Assets[toTicker]
    if !exists {
        toAsset = &Asset{
            ID:           toTicker,
            Negotiations: make([]parser.Transaction, 0),
            Earnings:     make([]parser.Earning, 0),
            Type:         fromAsset.Type,
            SubType:      fromAsset.SubType,
            Segment:      fromAsset.Segment,
        }
        w.Assets[toTicker] = toAsset
    }

    // Converter transações
    conversionMultiplier := decimal.NewFromInt(int64(ratio.To)).Div(decimal.NewFromInt(int64(ratio.From)))

    for _, tx := range fromAsset.Negotiations {
        convertedTx := tx
        convertedTx.Ticker = toTicker
        convertedTx.Quantity = tx.Quantity.Mul(conversionMultiplier)
        convertedTx.Price = tx.Price.Div(conversionMultiplier)
        // Amount permanece igual

        toAsset.Negotiations = append(toAsset.Negotiations, convertedTx)
        event.TransactionsMoved++
    }

    // Converter proventos
    for _, earning := range fromAsset.Earnings {
        convertedEarning := earning
        convertedEarning.Ticker = toTicker
        convertedEarning.Quantity = earning.Quantity.Mul(conversionMultiplier)
        convertedEarning.UnitPrice = earning.UnitPrice.Div(conversionMultiplier)
        // TotalAmount permanece igual

        toAsset.Earnings = append(toAsset.Earnings, convertedEarning)
        event.EarningsMoved++
    }

    // Recalcular ativo de destino
    w.RecalculateAsset(toAsset)

    // Remover ativo antigo
    delete(w.Assets, fromTicker)

    w.EventHistory = append(w.EventHistory, event)

    return event, nil
}
```

### 5. Histórico de Eventos

```go
type EventHistory struct {
    Events []CorporateEvent
}

type CorporateEvent interface {
    GetID() string
    GetTicker() string
    GetDate() time.Time
    GetType() string
    GetDescription() string
}

// GetEventHistory retorna todos os eventos de um ativo
func (w *Wallet) GetEventHistory(ticker string) []CorporateEvent {
    var events []CorporateEvent

    for _, event := range w.EventHistory {
        if event.GetTicker() == ticker {
            events = append(events, event)
        }
    }

    // Ordenar por data
    sort.Slice(events, func(i, j int) bool {
        return events[i].GetDate().Before(events[j].GetDate())
    })

    return events
}
```

---

## 🎨 Interface do Usuário (CLI)

### Comandos propostos:

```bash
# Desdobramento
b3cli events split ITSA4 1:2 2024-05-01
# Saída:
# ✓ Desdobramento aplicado com sucesso
#
# ITSA4 - Desdobramento 1:2
# Data: 01/05/2024
#
# Antes:  1,000 ações × R$ 10.50 = R$ 10,500.00
# Depois: 2,000 ações × R$  5.25 = R$ 10,500.00
#
# ✓ 45 transações ajustadas
# ✓ Preço médio recalculado

# Grupamento
b3cli events merge COGN3 10:1 2024-03-15
# Similar ao split, mas dividindo quantidade

# Bonificação
b3cli events bonus PETR4 10% 2024-06-20
# ou
b3cli events bonus PETR4 10 2024-06-20  # 10%

# Fusão/Aquisição
b3cli events acquisition LAME3 LAME4 1:1 2024-04-10

# Histórico de eventos
b3cli events history ITSA4
# Saída:
# 📋 Histórico de Eventos - ITSA4
#
# 2022-05-15  Desdobramento 1:2
# 2023-08-20  Bonificação 5%
# 2024-05-01  Desdobramento 1:2

# Listar todos os eventos
b3cli events list --year=2024
```

### TUI - Aplicar evento interativo:

```
╔══════════════════════════════════════════════════════════════════════╗
║              📊 APLICAR EVENTO CORPORATIVO                           ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Tipo de evento:                                                     ║
║  ○ Desdobramento (Split)                                             ║
║  ● Grupamento (Reverse Split)                                        ║
║  ○ Bonificação                                                       ║
║  ○ Fusão/Aquisição                                                   ║
║                                                                      ║
║  Ticker:         [COGN3___]                                          ║
║  Data:           [15/03/2024]                                        ║
║  Proporção:      [10] : [1]                                          ║
║                                                                      ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  PRÉVIA:                                                             ║
║                                                                      ║
║  Antes:   5,000 ações × R$  2.80 = R$ 14,000.00                     ║
║  Depois:    500 ações × R$ 28.00 = R$ 14,000.00                     ║
║                                                                      ║
║  ✓ 28 transações serão ajustadas                                    ║
║                                                                      ║
║  [Enter] Aplicar  [Esc] Cancelar                                    ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 🔄 Fluxo de Trabalho do Usuário

### Quando um evento acontece:
1. Usuário vê comunicado da empresa
2. `b3cli events split ITSA4 1:2 2024-05-01`
3. Sistema ajusta automaticamente
4. Carteira atualizada instantaneamente

### Ao importar dados históricos:
1. Importar transações antigas
2. Aplicar eventos corporativos em ordem cronológica
3. Dados ficam corretos automaticamente

---

## 🧪 Casos de Teste

```go
func TestSplit(t *testing.T) {
    // Split 1:2 → quantidade dobra, preço cai pela metade
    // Split 1:3 → quantidade triplica, preço divide por 3
    // Valor total permanece igual
}

func TestMerge(t *testing.T) {
    // Merge 10:1 → quantidade divide por 10, preço multiplica por 10
}

func TestBonus(t *testing.T) {
    // Bonificação 10% → +10% ações, preço médio ajustado
}

func TestAcquisition(t *testing.T) {
    // Conversão 1:1 → ticker muda, valores permanecem
    // Conversão 2:1 → ajuste proporcional
}
```

---

## 📊 Métricas de Sucesso

- ✅ 100% dos cálculos matematicamente corretos
- ✅ Zero perda de dados históricos
- ✅ Auditável (todos os eventos salvos)
- ✅ Compatível com declaração de IR

---

## 🚀 Expansões Futuras

1. **Detecção automática**
   - Integração com APIs que notificam eventos
   - Sugestões: "ITSA4 teve split, aplicar?"

2. **Reversão de eventos**
   - Desfazer evento aplicado incorretamente
   - Rollback completo

3. **Simulação**
   - Prever impacto antes de aplicar

---

**Estimativa de implementação:** 1-2 semanas
**ROI para usuários:** Muito Alto (elimina trabalho manual complexo)
