# Melhoria 12: Web e Mobile (Futuro)

**Prioridade:** P3 (Baixa)
**Complexidade:** Muito Alta
**Impacto:** Alto (longo prazo)

---

## 📋 Visão Geral

Expansão da ferramenta para além da CLI através de API REST, interface web e futuramente aplicativo mobile.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Limitação ao terminal**
   - Não acessível para usuários não-técnicos
   - Impossível usar no celular
   - Curva de aprendizado alta

2. **Falta de acessibilidade**
   - Pessoas sem conhecimento de CLI excluídas
   - Impossível mostrar para família/amigos
   - Adoção limitada

3. **Impossível usar em movimento**
   - Precisa estar no computador
   - Não pode checar carteira no celular
   - Sem notificações push

---

## 🏗️ Roadmap de Implementação

### Fase 1: API REST Local (3-4 semanas)

**Backend em Go:**

```go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/john/b3-project/internal/wallet"
)

type Server struct {
    router *gin.Engine
    wallet *wallet.Wallet
}

func NewServer(w *wallet.Wallet) *Server {
    s := &Server{
        router: gin.Default(),
        wallet: w,
    }

    s.setupRoutes()
    return s
}

func (s *Server) setupRoutes() {
    api := s.router.Group("/api/v1")
    {
        // Wallet
        api.GET("/wallet", s.getWallet)
        api.GET("/wallet/summary", s.getSummary)

        // Assets
        api.GET("/assets", s.listAssets)
        api.GET("/assets/:ticker", s.getAsset)
        api.POST("/assets/:ticker/metadata", s.updateMetadata)

        // Transactions
        api.GET("/transactions", s.listTransactions)
        api.POST("/transactions", s.addTransaction)

        // Earnings
        api.GET("/earnings", s.listEarnings)
        api.GET("/earnings/calendar", s.getCalendar)

        // Analytics
        api.GET("/analytics/allocation", s.getAllocation)
        api.GET("/analytics/performance", s.getPerformance)

        // Tax
        api.GET("/tax/calculate/:year/:month", s.calculateTax)
        api.GET("/tax/irpf/:year", s.getIRPFReport)

        // Market Data
        api.POST("/market/update", s.updatePrices)
        api.GET("/market/prices", s.getPrices)
    }
}

// Example handler
func (s *Server) getWallet(c *gin.Context) {
    c.JSON(200, gin.H{
        "path": s.wallet.Path,
        "assets_count": len(s.wallet.Assets),
        "total_invested": s.wallet.GetTotalInvested(),
    })
}

func (s *Server) Start(port int) error {
    return s.router.Run(fmt.Sprintf(":%d", port))
}
```

**Comandos:**

```bash
# Iniciar servidor
b3cli server start --port=8080

# Acessar API
curl http://localhost:8080/api/v1/wallet
curl http://localhost:8080/api/v1/assets
curl -X POST http://localhost:8080/api/v1/market/update
```

### Fase 2: Interface Web (4-6 semanas)

**Frontend em React/Vue:**

```jsx
// Dashboard.jsx
import React, { useState, useEffect } from 'react';
import { api } from './api';

function Dashboard() {
    const [wallet, setWallet] = useState(null);
    const [assets, setAssets] = useState([]);

    useEffect(() => {
        loadData();
    }, []);

    async function loadData() {
        const walletData = await api.get('/wallet');
        const assetsData = await api.get('/assets');

        setWallet(walletData);
        setAssets(assetsData);
    }

    return (
        <div className="dashboard">
            <header>
                <h1>📊 Minha Carteira</h1>
                <div className="summary">
                    <div className="metric">
                        <span>Total Investido</span>
                        <strong>R$ {wallet?.total_invested}</strong>
                    </div>
                    {/* ... */}
                </div>
            </header>

            <main>
                <section className="assets">
                    <h2>Ativos</h2>
                    <table>
                        <thead>
                            <tr>
                                <th>Ticker</th>
                                <th>Quantidade</th>
                                <th>Preço Médio</th>
                                <th>Total</th>
                            </tr>
                        </thead>
                        <tbody>
                            {assets.map(asset => (
                                <tr key={asset.id}>
                                    <td>{asset.ticker}</td>
                                    <td>{asset.quantity}</td>
                                    <td>R$ {asset.average_price}</td>
                                    <td>R$ {asset.total_invested}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </section>

                <section className="charts">
                    <AllocationChart data={wallet?.allocation} />
                    <PerformanceChart data={wallet?.performance} />
                </section>
            </main>
        </div>
    );
}
```

**Recursos:**
- ✅ Dashboard interativo
- ✅ Gráficos dinâmicos (Chart.js/Recharts)
- ✅ Filtros e buscas
- ✅ Edição inline de metadados
- ✅ Dark mode
- ✅ Responsivo (mobile-friendly)

### Fase 3: App Mobile (6-12 meses)

**Plataforma:** Flutter ou React Native

**Funcionalidades principais:**
- 📱 Dashboard mobile nativo
- 🔔 Notificações push (alertas de preço, proventos)
- 📊 Visualização de gráficos
- 💰 Registro rápido de transações
- 🔄 Sincronização com desktop
- 📴 Modo offline (cache local)

**Telas principais:**
1. Home/Dashboard
2. Lista de ativos
3. Detalhes do ativo
4. Adicionar transação
5. Proventos
6. Análises
7. Configurações

---

## 🎨 Mockups Conceituais

### API Endpoints:

```
GET    /api/v1/wallet
GET    /api/v1/wallet/summary
GET    /api/v1/assets
GET    /api/v1/assets/:ticker
POST   /api/v1/assets/:ticker/metadata
GET    /api/v1/transactions
POST   /api/v1/transactions
DELETE /api/v1/transactions/:hash
GET    /api/v1/earnings
GET    /api/v1/earnings/calendar
GET    /api/v1/analytics/allocation
GET    /api/v1/analytics/performance
GET    /api/v1/analytics/diversification
GET    /api/v1/tax/calculate/:year/:month
GET    /api/v1/tax/irpf/:year
POST   /api/v1/market/update
GET    /api/v1/market/prices
POST   /api/v1/market/alerts
GET    /api/v1/import/cei
POST   /api/v1/export/pdf
```

---

## 📊 Vantagens

**API REST:**
- ✅ Integração com outras ferramentas
- ✅ Automação via scripts
- ✅ Base para web/mobile

**Interface Web:**
- ✅ Acessível para não-programadores
- ✅ Visual mais amigável
- ✅ Gráficos interativos
- ✅ Fácil de compartilhar (localhost)

**App Mobile:**
- ✅ Acesso em qualquer lugar
- ✅ Notificações em tempo real
- ✅ Experiência mobile-first
- ✅ Maior adoção

---

## 🚧 Desafios

1. **Segurança**
   - API precisa de autenticação (JWT)
   - HTTPS obrigatório (certificado local)
   - Rate limiting

2. **Sincronização**
   - Conflitos entre CLI e Web
   - Lock de arquivo wallet.yaml
   - Concorrência

3. **Manutenção**
   - Mais código para manter
   - Bugs em múltiplas plataformas
   - Atualizações frequentes

---

## 🎯 Estratégia de Lançamento

**MVP (3 meses):**
1. API REST básica
2. Frontend web simples
3. Funcionalidades core (assets, transactions)

**v1.0 (6 meses):**
1. API completa
2. Web app com todas as features
3. Autenticação e segurança

**v2.0 (12+ meses):**
1. App mobile beta
2. Sincronização cloud (opcional)
3. Notificações push

---

## 📊 Valor para o Usuário

- 📈 **Adoção massiva:** Interface acessível
- 📱 **Mobilidade:** Usar em qualquer lugar
- 👥 **Compartilhamento:** Mostrar para família
- 🔔 **Proatividade:** Alertas em tempo real

---

**Estimativa de implementação:**
- API: 3-4 semanas
- Web: 4-6 semanas
- Mobile: 6-12 meses

**ROI para usuários:** Muito Alto (democratiza o acesso)
