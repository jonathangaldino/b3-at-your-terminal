# Melhoria 06: Cotações e Dados de Mercado

**Prioridade:** P2 (Média)
**Complexidade:** Média
**Impacto:** Alto

---

## 📋 Visão Geral

Integração com APIs de cotações para obter preços atuais, calcular valor de mercado da carteira, e acompanhar lucro/prejuízo em tempo real.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Desconhecimento do valor atual**
   - Usuário só sabe quanto investiu
   - Não sabe quanto vale hoje
   - Impossível saber se está lucrando

2. **Análise limitada**
   - Decisões baseadas em custo, não em valor atual
   - Não sabe quando comprar/vender
   - Difícil avaliar performance real

3. **Falta de alertas**
   - Perde oportunidades de compra
   - Não é notificado de quedas
   - Impossível fazer stop loss

### Benefícios mensuráveis:

- 💰 **Visibilidade total:** Valor real da carteira a qualquer momento
- 📊 **Performance real:** ROI baseado em preço atual
- 🔔 **Alertas inteligentes:** Notificações de preços-alvo
- 🎯 **Decisões melhores:** Comprar/vender no momento certo

---

## 🏗️ Arquitetura Proposta

### Componentes principais:

```
internal/market/
├── providers/
│   ├── brapi.go         # Brapi (API brasileira gratuita)
│   ├── yahoo.go         # Yahoo Finance
│   ├── alphavantage.go  # Alpha Vantage
│   └── provider.go      # Interface comum
├── quotes.go            # Cotações
├── cache.go             # Cache de preços
├── alerts.go            # Sistema de alertas
└── fundamentals.go      # Dados fundamentalistas

cmd/b3cli/
├── market.go            # Comando principal
├── market_update.go     # Atualizar cotações
├── market_prices.go     # Ver preços
└── market_alerts.go     # Gerenciar alertas
```

---

## 💡 Implementação Técnica

### 1. Interface de Provider

**Design pattern: Strategy**

```go
type PriceProvider interface {
    GetPrice(ticker string) (*Quote, error)
    GetPrices(tickers []string) (map[string]*Quote, error)
    GetHistoricalPrices(ticker string, from, to time.Time) ([]HistoricalQuote, error)
    IsAvailable() bool
    GetName() string
}

type Quote struct {
    Ticker          string
    Price           decimal.Decimal
    Change          decimal.Decimal  // Variação em R$
    ChangePercent   decimal.Decimal  // Variação em %
    Volume          int64
    Timestamp       time.Time
    Source          string           // "brapi", "yahoo", etc
}

type HistoricalQuote struct {
    Date   time.Time
    Open   decimal.Decimal
    High   decimal.Decimal
    Low    decimal.Decimal
    Close  decimal.Decimal
    Volume int64
}
```

### 2. Provider: Brapi (API Brasileira)

**Website:** https://brapi.dev/

```go
type BrapiProvider struct {
    APIKey     string
    BaseURL    string
    HTTPClient *http.Client
}

func NewBrapiProvider(apiKey string) *BrapiProvider {
    return &BrapiProvider{
        APIKey:     apiKey,
        BaseURL:    "https://brapi.dev/api",
        HTTPClient: &http.Client{Timeout: 10 * time.Second},
    }
}

// GetPrice obtém cotação de um único ativo
func (p *BrapiProvider) GetPrice(ticker string) (*Quote, error) {
    url := fmt.Sprintf("%s/quote/%s?token=%s", p.BaseURL, ticker, p.APIKey)

    resp, err := p.HTTPClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response struct {
        Results []struct {
            Symbol            string  `json:"symbol"`
            RegularMarketPrice float64 `json:"regularMarketPrice"`
            RegularMarketChange float64 `json:"regularMarketChange"`
            RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
            RegularMarketTime  int64   `json:"regularMarketTime"`
        } `json:"results"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }

    if len(response.Results) == 0 {
        return nil, fmt.Errorf("ticker %s não encontrado", ticker)
    }

    result := response.Results[0]

    return &Quote{
        Ticker:        result.Symbol,
        Price:         decimal.NewFromFloat(result.RegularMarketPrice),
        Change:        decimal.NewFromFloat(result.RegularMarketChange),
        ChangePercent: decimal.NewFromFloat(result.RegularMarketChangePercent),
        Timestamp:     time.Unix(result.RegularMarketTime, 0),
        Source:        "brapi",
    }, nil
}

// GetPrices obtém múltiplas cotações de uma vez
func (p *BrapiProvider) GetPrices(tickers []string) (map[string]*Quote, error) {
    // Brapi suporta múltiplos tickers: /quote/PETR4,VALE3,ITSA4
    tickersList := strings.Join(tickers, ",")
    url := fmt.Sprintf("%s/quote/%s?token=%s", p.BaseURL, tickersList, p.APIKey)

    resp, err := p.HTTPClient.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var response struct {
        Results []struct {
            Symbol            string  `json:"symbol"`
            RegularMarketPrice float64 `json:"regularMarketPrice"`
            RegularMarketChange float64 `json:"regularMarketChange"`
            RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
            RegularMarketTime  int64   `json:"regularMarketTime"`
        } `json:"results"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        return nil, err
    }

    quotes := make(map[string]*Quote)

    for _, result := range response.Results {
        quotes[result.Symbol] = &Quote{
            Ticker:        result.Symbol,
            Price:         decimal.NewFromFloat(result.RegularMarketPrice),
            Change:        decimal.NewFromFloat(result.RegularMarketChange),
            ChangePercent: decimal.NewFromFloat(result.RegularMarketChangePercent),
            Timestamp:     time.Unix(result.RegularMarketTime, 0),
            Source:        "brapi",
        }
    }

    return quotes, nil
}
```

### 3. Cache de Cotações

**Evitar chamadas excessivas à API:**

```go
type QuoteCache struct {
    data      map[string]*CachedQuote
    mutex     sync.RWMutex
    ttl       time.Duration  // Tempo de vida do cache
}

type CachedQuote struct {
    Quote     *Quote
    CachedAt  time.Time
}

func NewQuoteCache(ttl time.Duration) *QuoteCache {
    return &QuoteCache{
        data: make(map[string]*CachedQuote),
        ttl:  ttl,
    }
}

// Get obtém cotação do cache (se válida)
func (c *QuoteCache) Get(ticker string) (*Quote, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    cached, exists := c.data[ticker]
    if !exists {
        return nil, false
    }

    // Verificar se expirou
    if time.Since(cached.CachedAt) > c.ttl {
        return nil, false
    }

    return cached.Quote, true
}

// Set salva cotação no cache
func (c *QuoteCache) Set(ticker string, quote *Quote) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.data[ticker] = &CachedQuote{
        Quote:    quote,
        CachedAt: time.Now(),
    }
}

// Clear limpa cache
func (c *QuoteCache) Clear() {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    c.data = make(map[string]*CachedQuote)
}
```

### 4. Integração com Wallet

**Adicionar preços à wallet:**

```go
type MarketData struct {
    Quotes        map[string]*Quote
    LastUpdate    time.Time
    Provider      PriceProvider
    Cache         *QuoteCache
}

// UpdatePrices atualiza cotações de todos os ativos
func (w *Wallet) UpdatePrices() error {
    if w.Market == nil {
        return fmt.Errorf("market data não configurado")
    }

    // Obter lista de tickers
    tickers := make([]string, 0)
    for ticker := range w.Assets {
        tickers = append(tickers, ticker)
    }

    // Buscar cotações
    quotes, err := w.Market.Provider.GetPrices(tickers)
    if err != nil {
        return err
    }

    // Atualizar cache
    for ticker, quote := range quotes {
        w.Market.Cache.Set(ticker, quote)
    }

    w.Market.Quotes = quotes
    w.Market.LastUpdate = time.Now()

    return nil
}

// GetCurrentPrice retorna preço atual de um ativo
func (w *Wallet) GetCurrentPrice(ticker string) (decimal.Decimal, bool) {
    if w.Market == nil {
        return decimal.Zero, false
    }

    // Tentar cache primeiro
    if quote, ok := w.Market.Cache.Get(ticker); ok {
        return quote.Price, true
    }

    // Buscar da API
    quote, err := w.Market.Provider.GetPrice(ticker)
    if err != nil {
        return decimal.Zero, false
    }

    // Salvar no cache
    w.Market.Cache.Set(ticker, quote)

    return quote.Price, true
}

// CalculateMarketValue calcula valor de mercado da carteira
func (w *Wallet) CalculateMarketValue() decimal.Decimal {
    total := decimal.Zero

    for _, asset := range w.GetActiveAssets() {
        if price, ok := w.GetCurrentPrice(asset.ID); ok {
            assetValue := price.Mul(decimal.NewFromInt(int64(asset.Quantity)))
            total = total.Add(assetValue)
        }
    }

    return total
}

// CalculateUnrealizedPL calcula lucro/prejuízo não realizado
func (w *Wallet) CalculateUnrealizedPL() decimal.Decimal {
    marketValue := w.CalculateMarketValue()
    invested := decimal.Zero

    for _, asset := range w.GetActiveAssets() {
        invested = invested.Add(asset.TotalInvestedValue)
    }

    return marketValue.Sub(invested)
}
```

### 5. Sistema de Alertas

**Notificar quando preço atingir meta:**

```go
type PriceAlert struct {
    ID         string
    Ticker     string
    TargetPrice decimal.Decimal
    Condition   AlertCondition  // "above", "below"
    IsActive    bool
    CreatedAt   time.Time
    TriggeredAt *time.Time
}

type AlertCondition string

const (
    AlertAbove AlertCondition = "above"
    AlertBelow AlertCondition = "below"
)

// CreateAlert cria um alerta de preço
func (w *Wallet) CreateAlert(ticker string, targetPrice decimal.Decimal, condition AlertCondition) (*PriceAlert, error) {
    alert := &PriceAlert{
        ID:          generateUUID(),
        Ticker:      ticker,
        TargetPrice: targetPrice,
        Condition:   condition,
        IsActive:    true,
        CreatedAt:   time.Now(),
    }

    w.PriceAlerts = append(w.PriceAlerts, alert)

    return alert, nil
}

// CheckAlerts verifica alertas ativos
func (w *Wallet) CheckAlerts() []TriggeredAlert {
    triggered := make([]TriggeredAlert, 0)

    for _, alert := range w.PriceAlerts {
        if !alert.IsActive {
            continue
        }

        currentPrice, ok := w.GetCurrentPrice(alert.Ticker)
        if !ok {
            continue
        }

        shouldTrigger := false

        switch alert.Condition {
        case AlertAbove:
            shouldTrigger = currentPrice.GreaterThanOrEqual(alert.TargetPrice)
        case AlertBelow:
            shouldTrigger = currentPrice.LessThanOrEqual(alert.TargetPrice)
        }

        if shouldTrigger {
            now := time.Now()
            alert.TriggeredAt = &now
            alert.IsActive = false

            triggered = append(triggered, TriggeredAlert{
                Alert:        alert,
                CurrentPrice: currentPrice,
            })
        }
    }

    return triggered
}

type TriggeredAlert struct {
    Alert        *PriceAlert
    CurrentPrice decimal.Decimal
}
```

---

## 🎨 Interface do Usuário (CLI)

### Comandos propostos:

```bash
# Atualizar cotações
b3cli market update
b3cli market update --cache=5m  # Cache de 5 minutos

# Ver preços atuais
b3cli market prices
b3cli market prices PETR4 VALE3

# Valor de mercado
b3cli market value
# Saída:
# Valor investido:    R$ 125.430,00
# Valor de mercado:   R$ 142.850,00
# Lucro não realizado: R$  17.420,00 (+13.89%)

# Criar alerta
b3cli market alert PETR4 --above=40.00
b3cli market alert ITSA4 --below=10.00

# Listar alertas
b3cli market alerts list

# Verificar alertas
b3cli market alerts check
```

### TUI - Preços atuais:

```
╔══════════════════════════════════════════════════════════════════════╗
║              💹 COTAÇÕES ATUAIS                                      ║
╠══════════════════════════════════════════════════════════════════════╣
║  Atualizado em: 23/11/2024 15:30:45                                 ║
║                                                                      ║
║  Ticker    Preço      Variação      Qtd    Investido    Atual       ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  PETR4    R$ 38.50   +2.1% ↑      200    R$ 7,200    R$ 7,700  ✓    ║
║  VALE3    R$ 65.20   -1.5% ↓      100    R$ 6,800    R$ 6,520  ✗    ║
║  ITSA4    R$ 10.90   +0.3% ↑      500    R$ 5,450    R$ 5,450  =    ║
║  MXRF11   R$ 10.50   +0.5% ↑      300    R$ 3,150    R$ 3,150  =    ║
║                                                                      ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  TOTAL                              R$ 125.430   R$ 142.850          ║
║  Lucro não realizado: R$ 17.420,00 (+13.89%)                        ║
║                                                                      ║
║  [U] Atualizar  [A] Alertas  [q] Sair                               ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 🔄 Fluxo de Trabalho do Usuário

### Diário:
1. `b3cli market update`
2. Ver valor atual da carteira
3. Verificar alertas

### Ao decidir comprar/vender:
1. `b3cli market prices PETR4`
2. Ver preço atual e variação
3. Tomar decisão informada

---

## 📊 Métricas de Sucesso

- ✅ Cotações atualizadas em < 2 segundos
- ✅ Cache reduz chamadas à API em 80%+
- ✅ Alertas detectados em tempo real
- ✅ Suporte a 95%+ dos ativos B3

---

## 🚀 Expansões Futuras

1. **Dados fundamentalistas**
   - P/L, P/VP, ROE
   - Dividend Yield histórico
   - Lucro por ação

2. **Gráficos de preços**
   - Histórico de cotações
   - Candlestick charts
   - Comparação com IBOV

3. **Notificações push**
   - Email quando alerta disparar
   - Telegram bot
   - Desktop notifications

---

**Estimativa de implementação:** 1 semana
**ROI para usuários:** Alto (visibilidade de valor real)
