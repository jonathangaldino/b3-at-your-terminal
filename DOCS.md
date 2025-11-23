# B3CLI - Guia do Usuário

B3CLI é uma ferramenta de linha de comando para gerenciar sua carteira de investimentos da B3 (Bolsa de Valores Brasileira). Com ela, você pode importar transações de arquivos Excel, visualizar seus ativos, calcular preços médios ponderados e muito mais.

## Índice

- [Instalação](#instalação)
- [Comandos de Carteira](#comandos-de-carteira)
- [Comandos de Assets](#comandos-de-assets)
- [Comandos de Transação](#comandos-de-transação)
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

### `parse` - Importar transações de arquivos Excel

Parseia um ou mais arquivos `.xlsx` contendo transações financeiras da B3 e adiciona à carteira atual.

**Sintaxe:**
```bash
b3cli parse <arquivo1.xlsx> [arquivo2.xlsx] [...]
```

**Formato esperado do arquivo Excel:**
- Data do Negócio (DD/MM/YYYY)
- Tipo de Movimentação (Compra/Venda)
- Prazo/Vencimento (ignorado)
- Instituição
- Código da Negociação (ticker)
- Quantidade
- Preço
- Valor

**Exemplo:**
```bash
$ b3cli parse transactions-2023.xlsx transactions-2024.xlsx

Carregando wallet de: /Users/john/my-wallet
Processando 2 arquivo(s)...

✓ Wallet atualizada com sucesso!
  Transações antes: 0
  Transações novas: 245
  Transações duplicadas (ignoradas): 0
  Total de transações: 245

=== RESUMO ===
Total de transações únicas: 245
Total de ativos diferentes: 18

=== ATIVOS ===

[BBAS3] - renda variável
  Negociações: 12
  Preço Médio: R$ 27.6351
  Valor Total Investido: R$ 2846.4200
  Quantidade em carteira: 103
...
```

---

## Comandos de Assets

### `assets overview` - Visualizar ativos ativos

Exibe um resumo dos ativos que você possui atualmente (quantity != 0), organizados por tipo e segmento.

**Sintaxe:**
```bash
b3cli assets overview
```

**Exemplo:**
```bash
$ b3cli assets overview

=== RESUMO DE ATIVOS ===
Ativos em carteira: 16

[ações / bancos]
  BBAS3 - 103 ativos - R$ 2846.42 investido - PM: R$ 27.6351
  ITSA4 - 313 ativos - R$ 3293.14 investido - PM: R$ 10.5212
  SANB11 - 76 ativos - R$ 2003.25 investido - PM: R$ 26.3586

[ações / energia elétrica]
  ENBR3 - 299 ativos - R$ 3832.89 investido - PM: R$ 12.8190
  TAEE4 - 63 ativos - R$ 1356.77 investido - PM: R$ 21.5360

[fundos imobiliários / CRA]
  VGIA11 - 303 ativos - R$ 2581.61 investido - PM: R$ 8.5202

ℹ  Você possui 2 ativo(s) vendido(s) completamente.
   Use 'b3cli assets sold' para visualizá-los.
```

**Legenda:**
- **PM** = Preço Médio Ponderado
- **ativos** = Quantidade de ações/cotas em carteira
- **investido** = Valor total que você investiu (soma das compras)

---

### `assets sold` - Visualizar ativos vendidos

Exibe uma lista de ativos que foram vendidos completamente (quantity == 0).

**Sintaxe:**
```bash
b3cli assets sold
```

**Exemplo:**
```bash
$ b3cli assets sold

=== ATIVOS VENDIDOS COMPLETAMENTE ===
Total: 2

AESB3 - Vendido - R$ 142.02 investido (PM: R$ 14.2020)
PETR4 - Vendido - R$ 1245.80 investido (PM: R$ 28.9952)

ℹ  Estes ativos foram vendidos completamente mas seu histórico
   de transações ainda está disponível em transactions.yaml
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

Sua carteira é composta por três arquivos YAML:

```
minha-carteira/
├── assets.yaml          # Ativos ativos (quantity != 0)
├── sold-assets.yaml     # Ativos vendidos completamente
└── transactions.yaml    # Todas as transações
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
