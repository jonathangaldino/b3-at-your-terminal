# B3CLI - Guia do Usuário

B3CLI é uma ferramenta de linha de comando para gerenciar sua carteira de investimentos da B3 (Bolsa de Valores Brasileira). Com ela, você pode importar transações de arquivos Excel, visualizar seus ativos, calcular preços médios ponderados e muito mais.

## Índice

- [Instalação](#instalação)
- [Comandos de Carteira](#comandos-de-carteira)
- [Comando de Importação](#comando-de-importação)
- [Comandos de Assets](#comandos-de-assets)
- [Comandos de Transação](#comandos-de-transação)
- [Comandos de Proventos](#comandos-de-proventos)
- [Fluxo de Trabalho Típico](#fluxo-de-trabalho-típico)

---

## Instalação

```bash
# Clone o repositório
git clone https://github.com/john/b3-project
cd b3-project

# Compile o projeto
go build ./cmd/b3cli

# Execute
./b3cli --help
```

---

## Comandos de Carteira

### `wallet create` - Criar uma nova carteira

Cria uma nova carteira vazia em um diretório especificado.

**Sintaxe:**
```bash
b3cli wallet create <diretório>
```

**Exemplo:**
```bash
$ b3cli wallet create ./my-wallet

✓ Carteira criada com sucesso em: /Users/john/my-wallet
✓ Arquivos criados:
  - /Users/john/my-wallet/assets.yaml
  - /Users/john/my-wallet/transactions.yaml

Próximos passos:
  1. Abra a carteira: b3cli wallet open /Users/john/my-wallet
  2. Importe transações: b3cli parse arquivos/*.xlsx
  3. Visualize seus ativos: b3cli assets overview
```

---

### `wallet open` - Abrir uma carteira existente

Define a carteira atual que será usada pelos outros comandos.

**Sintaxe:**
```bash
b3cli wallet open <diretório>
```

**Exemplo:**
```bash
$ b3cli wallet open ./my-wallet

✓ Wallet aberta: /Users/john/my-wallet
✓ Arquivos:
  - /Users/john/my-wallet/assets.yaml
  - /Users/john/my-wallet/transactions.yaml

Agora você pode usar os comandos sem especificar wallet:
  b3cli parse arquivos/*.xlsx
  b3cli assets overview
  b3cli assets subscription TICKER subscription@PARENT
```

---

### `wallet current` - Ver carteira atual

Mostra qual carteira está atualmente aberta.

**Sintaxe:**
```bash
b3cli wallet current
```

**Exemplo:**
```bash
$ b3cli wallet current

Wallet atual: /Users/john/my-wallet
```

---

### `wallet close` - Fechar carteira atual

Fecha a carteira atual. Após executar este comando, será necessário abrir uma carteira novamente.

**Sintaxe:**
```bash
b3cli wallet close
```

**Exemplo:**
```bash
$ b3cli wallet close

✓ Wallet fechada: /Users/john/my-wallet

Para trabalhar com uma wallet novamente:
  b3cli wallet open <diretório>
```

---

## Comando de Importação

### `parse` - Importar transações e proventos de arquivos Excel

Parseia automaticamente um ou mais arquivos `.xlsx` da B3, detectando se são transações ou proventos.

**Sintaxe:**
```bash
b3cli parse <arquivo1.xlsx> [arquivo2.xlsx] [...]
```

**Detecção Automática:**
O comando identifica automaticamente o tipo de arquivo baseado no número de colunas:
- **9 colunas**: Arquivo de TRANSAÇÕES (compra/venda)
- **8 colunas**: Arquivo de PROVENTOS (rendimentos/dividendos/JCP/resgates)

**Formato de Transações (9 colunas):**
- Data do Negócio (DD/MM/YYYY)
- Tipo de Movimentação (Compra/Venda)
- Mercado
- Prazo/Vencimento
- Instituição
- Código da Negociação (ticker)
- Quantidade
- Preço
- Valor

**Formato de Proventos (8 colunas):**
- Entrada/Saída
- Data (DD/MM/YYYY)
- Movimentação (Rendimento/Dividendo/Juros Sobre Capital Próprio/Resgate)
- Produto (formato: TICKER - Nome da empresa)
- Instituição
- Quantidade
- Preço unitário
- Valor da Operação

**Exemplo:**
```bash
$ b3cli parse transactions-2023.xlsx proventos-2024.xlsx

Carregando wallet de: /Users/john/my-wallet
Processando 2 arquivo(s)...

  - transactions-2023.xlsx: detectado como arquivo de TRANSAÇÕES
  - proventos-2024.xlsx: detectado como arquivo de PROVENTOS

Processando 1 arquivo(s) de transações...
  ✓ Transações: 245 adicionadas, 0 duplicadas

Processando 1 arquivo(s) de proventos...
  ✓ Proventos: 128 adicionados, 0 duplicados

✓ Wallet atualizada com sucesso!
  Total adicionado: 373
  Total duplicados (ignorados): 0

# Interface interativa colorida (Bubble Tea) é exibida
# Mostrando resumo de ativos, proventos e transações
```

**Resultado:** Uma interface terminal interativa (TUI) colorida é exibida mostrando:
- Resumo geral (transações, proventos, ativos)
- Lista detalhada de cada ativo
- Últimas 10 transações processadas

---

## Comandos de Assets

### `assets overview` - Visualizar ativos ativos

Exibe um resumo **interativo e colorido** dos ativos que você possui atualmente (quantity != 0), organizados por tipo e segmento.

**Sintaxe:**
```bash
b3cli assets overview
```

**Interface:**
Uma interface terminal interativa (Bubble Tea) colorida é exibida com:
- 📊 Título em destaque
- 📁 Grupos por tipo e segmento
- 🎨 Cores para facilitar leitura:
  - **Azul ciano**: Tickers
  - **Amarelo**: Quantidades
  - **Verde**: Valores monetários
  - **Azul claro**: Preço médio
- ℹ️ Dicas sobre ativos vendidos

**Navegação:**
- `q` ou `ESC`: Sair

**Exemplo visual:**
```
📊 Resumo de Ativos
Ativos em carteira: 16

📁 ações / bancos

  BBAS3    103 ativos • Investido: R$   2846.42 • PM: R$ 27.6351
  ITSA4    313 ativos • Investido: R$   3293.14 • PM: R$ 10.5212
  SANB11    76 ativos • Investido: R$   2003.25 • PM: R$ 26.3586

📁 ações / energia elétrica

  ENBR3    299 ativos • Investido: R$   3832.89 • PM: R$ 12.8190
  TAEE4     63 ativos • Investido: R$   1356.77 • PM: R$ 21.5360

ℹ  Você possui 2 ativo(s) vendido(s) completamente.
   Use 'b3cli assets sold' para visualizá-los.

q/esc: sair
```

**Legenda:**
- **PM** = Preço Médio Ponderado
- **ativos** = Quantidade de ações/cotas em carteira
- **investido** = Valor total que você investiu (soma das compras)

---

### `assets sold` - Visualizar ativos vendidos

Exibe uma lista **interativa e colorida** de ativos que foram vendidos completamente (quantity == 0).

**Sintaxe:**
```bash
b3cli assets sold
```

**Interface:**
Uma interface terminal interativa (Bubble Tea) colorida é exibida com:
- 🔴 Título em destaque
- 🎨 Status "Vendido" em vermelho itálico
- 💰 Valores e preços médios destacados

**Navegação:**
- `q` ou `ESC`: Sair

**Exemplo visual:**
```
🔴 Ativos Vendidos Completamente
Total: 2

AESB3      Vendido
  Investido: R$  142.02 • PM: R$ 14.2020

PETR4      Vendido
  Investido: R$ 1245.80 • PM: R$ 28.9952

ℹ  Estes ativos foram vendidos completamente mas seu histórico
   de transações ainda está disponível em transactions.yaml

q/esc: sair
```

---

### `assets manage` - Gerenciar metadados de ativos (TUI)

Interface interativa (Terminal UI) para gerenciar metadados dos ativos: tipo, subtipo e segmento.

**Sintaxe:**
```bash
b3cli assets manage
```

**Navegação:**
- `↑/↓` ou `j/k`: navegar pela lista
- `Enter`: selecionar ativo para editar
- `Tab/↑/↓`: navegar entre campos de edição
- `Enter`: salvar alterações
- `Esc`: voltar para lista
- `q` ou `Ctrl+C`: sair

**Exemplo de uso:**

1. Execute o comando:
```bash
$ b3cli assets manage
```

2. Você verá uma lista de ativos:
```
Gerenciar Ativos

┌─────────────────────────────────┐
│ Selecione um ativo para gerenciar
│ > BBAS3 (103 ativos)
│   PM: R$ 27.6351
│   ITSA4 (313 ativos)
│   PM: R$ 10.5212
│   ...
└─────────────────────────────────┘

enter: selecionar • q: sair
```

3. Pressione `Enter` para editar um ativo:
```
Editando: BBAS3

► Type:
  renda variável

  SubType:
  ações

  Segment:
  bancos

tab/↑/↓: navegar • enter: salvar • esc: voltar • ctrl+c: sair
```

4. Edite os campos e pressione `Enter` para salvar.

---

### `assets subscription` - Vincular direito de subscrição

Marca um ativo como sendo um direito de subscrição de outro ativo e transfere as transações.

**Sintaxe:**
```bash
b3cli assets subscription <ticker-subscrição> subscription@<ticker-pai>
```

**Exemplo:**
```bash
$ b3cli assets subscription MXRF12 subscription@MXRF11

Processando subscrição MXRF12 → MXRF11...

✓ Processamento concluído:
  - Compras encontradas: 5
  - Vendas encontradas: 0 (ignoradas)
  - Transações transferidas: 5

✓ Ativo MXRF12 removido da carteira
✓ Ativo MXRF11 atualizado:
  - Quantidade antes: 400
  - Quantidade depois: 450
  - Preço médio: R$ 9.5420

✓ Wallet atualizada em: /Users/john/my-wallet/wallet.yaml
```

---

## Comandos de Transação

### `assets buy` - Comprar ativos (TUI)

Interface interativa para registrar manualmente a compra de ativos.

**Sintaxe:**
```bash
b3cli assets buy
```

**Fluxo de uso:**

1. Execute o comando:
```bash
$ b3cli assets buy
```

2. Preencha os campos:
```
Buy Assets

Ticker:
BBAS3

Date:
2024-11-23

Quantity:
100

Unit Price:
27.50

Press Enter to continue, Esc to cancel
```

3. Revise o resumo:
```
Transaction Summary

Ticker:       BBAS3
Date:         2024-11-23
Quantity:     100.0000
Unit Price:   R$ 27.50
Total Amount: R$ 2750.00

Current Average Price: R$ 27.64
✓ Buying BELOW average price (-0.51%, R$ -0.14)

Proceed with this transaction?
Press Enter to confirm, N to edit, Esc to cancel
```

4. Pressione `Enter` para confirmar:
```
✓ Transaction saved successfully!
```

**Recursos:**
- **Data**: Deixe em branco para usar a data de hoje
- **Comparação de preço**: Mostra se está comprando acima ou abaixo do preço médio atual
- **Validação**: Todos os campos são validados automaticamente

---

### `assets sell` - Vender ativos (TUI)

Interface interativa para registrar manualmente a venda de ativos.

**Sintaxe:**
```bash
b3cli assets sell
```

**Fluxo de uso:**

1. Execute o comando:
```bash
$ b3cli assets sell
```

2. Preencha os campos:
```
Sell Assets

Ticker:
BBAS3

Date:
2024-11-23

Quantity:
50

Unit Price:
30.00

Press Enter to continue, Esc to cancel
```

3. Revise o resumo:
```
Transaction Summary

Ticker:       BBAS3
Date:         2024-11-23
Quantity:     50.0000
Unit Price:   R$ 30.00
Total Amount: R$ 1500.00

Remaining after sale: 53 shares

Proceed with this transaction?
Press Enter to confirm, N to edit, Esc to cancel
```

4. Pressione `Enter` para confirmar:
```
✓ Transaction saved successfully!
```

**Validações automáticas:**
- Verifica se o ativo existe na carteira
- Verifica se você tem quantidade suficiente para vender
- Exemplo de erro:
```
Error

insufficient quantity. You have 103 shares, trying to sell 200

Press Enter to go back, Esc to cancel
```

---

## Comandos de Proventos

### `earnings parse` - Importar proventos de arquivos Excel

Parseia um ou mais arquivos `.xlsx` contendo proventos recebidos da B3 e adiciona à carteira atual.

**Sintaxe:**
```bash
b3cli earnings parse <arquivo1.xlsx> [arquivo2.xlsx] [...]
```

**Tipos de proventos suportados:**
- **Rendimento**: Pagamentos periódicos (comum em FIIs)
- **Dividendo**: Distribuição de lucros
- **JCP (Juros Sobre Capital Próprio)**: Distribuição com benefício fiscal
- **Resgate**: Fechamento de capital ou retirada de circulação

**Formato esperado do arquivo Excel (8 colunas):**
- Entrada/Saída (ignorado)
- Data (DD/MM/YYYY)
- Movimentação (tipo: Rendimento/Dividendo/Juros Sobre Capital Próprio/Resgate)
- Produto (formato: TICKER - Nome da empresa)
- Instituição (ignorado)
- Quantidade
- Preço unitário
- Valor da operação (total a receber)

**Exemplo:**
```bash
$ b3cli earnings parse proventos-2024.xlsx

Carregando wallet de: /Users/john/my-wallet
Processando 1 arquivo(s) de proventos...

✓ Wallet atualizada com sucesso!
  Proventos antes: 45
  Proventos novos: 83
  Proventos duplicados (ignorados): 2
  Total de proventos: 128

=== RESUMO DE PROVENTOS POR ATIVO ===

[MXRF11]
  Total de proventos recebidos: 24
  Valor total recebido: R$ 245.80
    - Rendimentos: 24

[BBAS3]
  Total de proventos recebidos: 12
  Valor total recebido: R$ 89.40
    - Dividendos: 10
    - JCP: 2
```

**Recursos:**
- Deduplicação automática por hash
- Atualização do total de proventos por ativo
- Validação de tipo de provento
- Extração automática do ticker do campo "Produto"

---

### `earnings overview` - Resumo de proventos por tipo

Exibe um resumo **interativo e colorido** de todos os proventos recebidos, agrupados por tipo.

**Sintaxe:**
```bash
b3cli earnings overview
```

**Interface:**
Uma interface terminal interativa (Bubble Tea) colorida é exibida com:
- 💰 Título destacado
- 📊 Resumo geral (total de pagamentos e valor total)
- 🎨 Seções por categoria com cores:
  - **📊 Rendimentos** (verde)
  - **💵 Dividendos** (amarelo)
  - **🏦 JCP** (azul)
  - **🔄 Resgates** (roxo)
- 💡 Percentual de cada tipo
- 📈 Lista de ativos pagadores ordenada por valor

**Navegação:**
- `q` ou `ESC`: Sair

**Exemplo visual:**
```
💰 Resumo Geral de Proventos

Total de pagamentos recebidos: 245
Valor total recebido: R$ 4,832.50

────────────────────────────────────────────────────────────

📊 RENDIMENTOS

  Quantidade de pagamentos: 156
  Valor total: R$ 3,245.80
  Percentual do total: 67.15%

  Ativos que pagaram:
    MXRF11    R$    1,245.80
    HGLG11    R$      892.50
    VGIA11    R$      658.30
    ...

────────────────────────────────────────────────────────────

💵 DIVIDENDOS

  Quantidade de pagamentos: 78
  Valor total: R$ 1,245.70
  Percentual do total: 25.77%

  Ativos que pagaram:
    BBAS3     R$      456.20
    ITSA4     R$      389.50
    ...

q/esc: sair
```

**Informações exibidas:**
- Total de pagamentos por tipo
- Valor total recebido por tipo
- Percentual em relação ao total geral
- Lista de ativos ordenada por valor (maior primeiro)

---

### `earnings reports` - Relatórios por período

Exibe relatórios **interativos** de proventos recebidos, com visualização anual ou mensal.

**Sintaxe:**
```bash
b3cli earnings reports
```

**Interface:**
Uma interface terminal interativa (Bubble Tea) com múltiplas telas:

1. **Seleção de tipo de relatório:**
   - Anual (resumo por ano)
   - Mensal (resumo por mês)

2. **Seleção de ano** (se necessário para relatório mensal)

3. **Visualização do relatório** com cores e formatação

**Navegação:**
- `↑/↓` ou `j/k`: Navegar pelas opções
- `Enter`: Selecionar
- `ESC`: Voltar para tela anterior
- `q`: Sair

**Exemplo - Relatório Anual:**
```
📈 Relatório Anual de Proventos

2020:  R$  1,245.80
2021:  R$  2,389.50
2022:  R$  3,456.20
2023:  R$  4,832.10
2024:  R$  5,245.90

────────────────────────────────────
Total geral: R$ 17,169.50
Média anual: R$ 3,433.90

esc: voltar • q: sair
```

**Exemplo - Relatório Mensal:**
```
📅 Relatório Mensal de Proventos - 2024

Janeiro:      R$    456.80
Fevereiro:    R$    389.20
Março:        R$    512.30
Abril:        R$    445.60
Maio:         R$    498.70
Junho:        R$    432.10
...

────────────────────────────────────
Total do ano: R$ 5,245.90
Média mensal (meses com pagamento): R$ 437.16

esc: voltar • q: sair
```

**Recursos:**
- Seleção interativa de tipo de relatório
- Múltiplos anos suportados automaticamente
- Cálculo automático de médias
- Navegação fluida entre telas

---

## Fluxo de Trabalho Típico

### Cenário 1: Primeira vez usando o B3CLI

```bash
# 1. Crie uma nova carteira
b3cli wallet create ./minha-carteira

# 2. Abra a carteira
b3cli wallet open ./minha-carteira

# 3. Importe suas transações do arquivo Excel da B3
b3cli parse ~/Downloads/notas-corretagem-2024.xlsx

# 4. Visualize seus ativos
b3cli assets overview

# 5. Organize seus ativos (opcional)
b3cli assets manage
```

### Cenário 2: Adicionando novas transações

```bash
# 1. Abra sua carteira (se ainda não estiver aberta)
b3cli wallet open ./minha-carteira

# 2. Importe o novo arquivo
b3cli parse ~/Downloads/notas-novembro-2024.xlsx

# 3. Visualize o resumo atualizado
b3cli assets overview
```

### Cenário 3: Registrando uma compra manual

```bash
# 1. Abra sua carteira
b3cli wallet open ./minha-carteira

# 2. Registre a compra
b3cli assets buy
# (siga o fluxo interativo)

# 3. Visualize o ativo atualizado
b3cli assets overview
```

### Cenário 4: Lidando com direitos de subscrição

```bash
# 1. Abra sua carteira
b3cli wallet open ./minha-carteira

# 2. Vincule o direito de subscrição ao ativo pai
b3cli assets subscription PETR12 subscription@PETR4

# 3. Verifique o ativo pai atualizado
b3cli assets overview
```

### Cenário 5: Acompanhando proventos

```bash
# 1. Abra sua carteira
b3cli wallet open ./minha-carteira

# 2. Importe arquivo de proventos da B3
b3cli earnings parse ~/Downloads/proventos-2024.xlsx

# 3. Visualize resumo por tipo (Rendimentos, Dividendos, JCP)
b3cli earnings overview

# 4. Veja evolução anual ou mensal
b3cli earnings reports
# (navegue interativamente entre relatórios anuais e mensais)
```

### Cenário 6: Processamento completo (Transações + Proventos)

```bash
# 1. Abra sua carteira
b3cli wallet open ./minha-carteira

# 2. Importe tudo de uma vez (detecção automática)
b3cli parse ~/Downloads/transacoes-2024.xlsx ~/Downloads/proventos-2024.xlsx

# 3. Visualize seus ativos
b3cli assets overview

# 4. Acompanhe seus ganhos passivos
b3cli earnings overview
```

---

## Dicas e Boas Práticas

### 1. Backup Regular
Sempre faça backup dos arquivos YAML da sua carteira:
```bash
cp -r ./minha-carteira ./minha-carteira-backup-$(date +%Y%m%d)
```

### 2. Organização de Ativos
Use o comando `assets manage` para classificar seus ativos por tipo, subtipo e segmento. Isso facilita a visualização organizada no `assets overview`.

### 3. Importação Incremental
O B3CLI detecta automaticamente transações duplicadas (por hash). Você pode importar o mesmo arquivo várias vezes sem medo de duplicação.

### 4. Verificação de Dados
Sempre revise o resumo após importar transações:
```bash
b3cli parse arquivo.xlsx
b3cli assets overview  # Verifique se os valores fazem sentido
```

### 5. Formato de Data
Ao usar `assets buy` ou `assets sell`, use sempre o formato `YYYY-MM-DD` (ex: 2024-11-23).

---

## Estrutura de Arquivos

Sua carteira é composta por quatro arquivos YAML:

```
minha-carteira/
├── assets.yaml          # Ativos ativos (quantity != 0)
├── sold-assets.yaml     # Ativos vendidos completamente
├── transactions.yaml    # Todas as transações de compra/venda
└── earnings.yaml        # Todos os proventos recebidos
```

### `assets.yaml`
```yaml
- ticker: BBAS3
  type: renda variável
  subtype: ações
  segment: bancos
  average_price: "27.6351"
  total_invested_value: "2846.4200"
  quantity: 103
```

### `transactions.yaml`
```yaml
- date: "2020-08-10"
  type: Compra
  institution: RICO INVESTIMENTOS - GRUPO XP
  ticker: BBAS3
  quantity: "100.0000"
  price: "27.5000"
  amount: "2750.0000"
  hash: 5c03b2001f5ca1bbc796d292e40d3e95fb777c55631665f12c9db10f1b43f9e5
```

### `earnings.yaml`
```yaml
- date: "2024-03-15"
  type: Dividendo
  ticker: BBAS3
  quantity: "100.0000"
  unit_price: "0.4500"
  total_amount: "45.0000"
  hash: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6
```

---

## Troubleshooting

### Erro: "no wallet is currently open"
**Solução:** Abra uma carteira primeiro:
```bash
b3cli wallet open ./minha-carteira
```

### Erro: "wallet não encontrada"
**Solução:** Crie uma carteira primeiro:
```bash
b3cli wallet create ./minha-carteira
```

### Erro: "arquivo não encontrado"
**Solução:** Verifique o caminho do arquivo Excel:
```bash
ls ~/Downloads/*.xlsx
b3cli parse ~/Downloads/arquivo-correto.xlsx
```

### Erro: "duplicate transaction detected"
**Explicação:** Essa transação já foi importada anteriormente. Isso é normal e esperado ao reimportar arquivos.

### Erro: "insufficient quantity"
**Explicação:** Você está tentando vender mais ações do que possui. Verifique a quantidade disponível:
```bash
b3cli assets overview
```

---

## Suporte

Para reportar problemas ou sugerir melhorias, abra uma issue no GitHub:
https://github.com/john/b3-project/issues

---

## Licença

Este projeto é open source. Veja o arquivo LICENSE para detalhes.
