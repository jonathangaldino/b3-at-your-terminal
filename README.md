# B3 At Your Terminal

> **English speakers**: This is a CLI tool for parsing and analyzing financial transactions from B3 (Brasil, Bolsa, Balcão - the Brazilian stock exchange, similar to NASDAQ). The tool processes Excel files exported from B3 brokerage accounts, calculates weighted average prices, and helps manage investment portfolios. Documentation is in Portuguese as the tool is specific to the Brazilian market.

---

## 📊 Sobre o Projeto

**B3 At Your Terminal** é uma ferramenta de linha de comando desenvolvida em Go para processar e analisar transações financeiras da B3 (Brasil, Bolsa, Balcão).

A ferramenta permite que você:

- 📁 Processe arquivos Excel (.xlsx) exportados da sua conta na B3
- 💰 Calcule automaticamente o preço médio ponderado dos seus ativos
- 📈 Visualize sua carteira de investimentos de forma organizada
- 🔍 Identifique e elimine transações duplicadas
- 🏢 Agrupe suas negociações por ativo (ticker)

## ✨ Funcionalidades

- ✅ **Parser de arquivos .xlsx** - Lê arquivos Excel exportados diretamente da B3
- ✅ **Cálculo automático de preço médio** - Calcula o custo médio ponderado de cada ativo
- ✅ **Deduplicação inteligente** - Usa hash SHA256 para identificar e eliminar transações duplicadas
- ✅ **Normalização de códigos** - Unifica ativos do mercado fracionário (remove "F" quando aplicável)
- ✅ **Carteira consolidada** - Visualize todos os seus ativos em um único lugar
- ✅ **Suporte a múltiplos arquivos** - Processe vários períodos de uma só vez

## 🚀 Como Usar

### Pré-requisitos

- Go 1.24 ou superior instalado
- Arquivos .xlsx exportados da sua conta B3

### Instalação

```bash
# Clone o repositório
git clone https://github.com/john/b3-project.git
cd b3-project

# Compile o projeto
go build -o b3cli ./cmd/b3cli

# (Opcional) Mova para um diretório no PATH
sudo mv b3cli /usr/local/bin/
```

### Exportando arquivos da B3

⚠️ **IMPORTANTE**: Esta CLI aceita **apenas arquivos .xlsx exportados diretamente da sua conta na B3** ou da sua corretora.

Os arquivos devem conter as seguintes colunas:

- Data do Negócio
- Tipo de Movimentação (Compra/Venda)
- Mercado
- Prazo/Vencimento
- Instituição
- Código de Negociação
- Quantidade
- Preço
- Valor

### Exemplos de Uso

**Processar um único arquivo:**

```bash
./b3cli parse arquivos/compras-2024.xlsx
```

**Processar múltiplos arquivos:**

```bash
./b3cli parse arquivos/compras-2023.xlsx arquivos/compras-2024.xlsx
```

**Processar todos os arquivos de uma pasta:**

```bash
./b3cli parse files/*.xlsx
```

**Processar apenas arquivos numerados:**

```bash
./b3cli parse files/[0-9]*.xlsx
```

### Exemplo de Saída

```
Processando 7 arquivo(s)...

=== RESUMO ===
Total de transações únicas: 191
Total de ativos diferentes: 22

=== ATIVOS ===

[ITSA4] - renda variável
  Negociações: 13
  Preço Médio: R$ 10.18
  Quantidade em carteira: 150

[BCFF11] - renda variável
  Negociações: 4
  Preço Médio: R$ 86.96
  Quantidade em carteira: 25

[BBAS3] - renda variável
  Negociações: 10
  Preço Médio: R$ 27.64
  Quantidade em carteira: 103

=== TRANSAÇÕES ===
Hash                 | Data       | Tipo   | Ticker | Qtd    | Preço   | Valor
--------------------------------------------------------------------------------
2d512e793528d6f5...  | 08/09/2020 | Compra | ITSA4  |     10 |    9.61 |   96.10
24253c8131a40951...  | 02/09/2020 | Compra | BCFF11 |      1 |   90.30 |   90.30
...
```

## 📁 Estrutura do Projeto

```
b3-project/
├── cmd/
│   └── b3cli/
│       └── main.go              # Entry point da aplicação
├── internal/
│   ├── cli/                     # Comandos CLI (Cobra)
│   │   ├── root.go
│   │   └── parse.go
│   ├── parser/                  # Lógica de parsing
│   │   ├── transaction.go
│   │   ├── hash.go
│   │   └── parser.go
│   └── wallet/                  # Gestão de carteira
│       ├── asset.go
│       ├── wallet.go
│       └── calculator.go
├── specs/                       # Documentação técnica
├── files/                       # Seus arquivos .xlsx
├── go.mod
├── go.sum
└── README.md
```

## 🏗️ Arquitetura

O projeto foi desenvolvido seguindo princípios de **separação de responsabilidades** e **desacoplamento**:

- **Parser**: Responsável apenas por ler e parsear arquivos .xlsx
- **Wallet**: Gerencia a carteira, ativos e cálculos financeiros
- **CLI**: Interface de linha de comando, independente da lógica de negócio

### Tecnologias Utilizadas

- **Go 1.24** - Linguagem principal
- **Cobra** - Framework para CLI
- **Excelize** - Biblioteca para leitura de arquivos Excel (.xlsx)
- **SHA256** - Algoritmo de hash para deduplicação

## 🔒 Privacidade e Segurança

- ✅ Todos os dados são processados **localmente** no seu computador
- ✅ Nenhuma informação é enviada para servidores externos
- ✅ Seus arquivos e transações permanecem **100% privados**
- ✅ Código aberto e auditável

## 📝 Como Funciona

1. **Parsing**: A CLI lê seus arquivos .xlsx e extrai as transações
2. **Normalização**: Códigos do mercado fracionário são normalizados (ex: ITSA4F → ITSA4)
3. **Deduplicação**: Cada transação recebe um hash SHA256 único para evitar duplicatas
4. **Agregação**: Transações são agrupadas por ticker (código do ativo)
5. **Cálculo**: O preço médio ponderado é calculado automaticamente
6. **Visualização**: Resultados são exibidos de forma organizada no terminal

### Cálculo do Preço Médio Ponderado

```
Preço Médio = Σ(preço × quantidade) / Σ(quantidade)
```

Apenas transações de **compra** são consideradas no cálculo.

## 🤝 Contribuindo

Contribuições são bem-vindas! Sinta-se à vontade para:

1. Fazer um fork do projeto
2. Criar uma branch para sua feature (`git checkout -b feature/MinhaFeature`)
3. Commit suas mudanças (`git commit -m 'Adiciona MinhaFeature'`)
4. Push para a branch (`git push origin feature/MinhaFeature`)
5. Abrir um Pull Request

### Regras do Projeto

- Manter segregação entre pacotes (CLI, Parser, Wallet)
- Não acoplar pacotes diretamente
- Perguntar antes de adicionar novas dependências
- Escrever código claro e bem documentado

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo `LICENSE` para mais detalhes.

## ⚠️ Disclaimer

Esta ferramenta é fornecida "como está", sem garantias de qualquer tipo. Os cálculos de preço médio e análises são baseados nos dados fornecidos nos arquivos Excel. Sempre consulte um profissional de investimentos certificado para decisões financeiras importantes.

## 📧 Contato

Para dúvidas, sugestões ou reportar problemas, abra uma [issue](https://github.com/john/b3-project/issues) no GitHub.

---

**Desenvolvido com ❤️ para investidores brasileiros**
