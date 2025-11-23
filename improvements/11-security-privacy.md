# Melhoria 11: Segurança e Privacy

**Prioridade:** P3 (Baixa)
**Complexidade:** Média
**Impacto:** Baixo-Médio

---

## 📋 Visão Geral

Funcionalidades para proteger dados sensíveis através de criptografia, anonimização e auditoria.

---

## 🎯 Valor para o Usuário

### Problemas que resolve:

1. **Dados sensíveis desprotegidos**
   - wallet.yaml em plain text
   - Qualquer um com acesso ao PC vê tudo
   - Risco em ambientes compartilhados

2. **Impossível compartilhar screenshots**
   - Valores visíveis em capturas de tela
   - Não pode demonstrar a ferramenta
   - Privacy comprometida

3. **Sem auditoria**
   - Mudanças acidentais
   - Impossível desfazer
   - Sem histórico de modificações

---

## 🏗️ Implementação

### 1. Criptografia de Wallet

```go
import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "crypto/sha256"
    "golang.org/x/crypto/pbkdf2"
)

// EncryptWallet criptografa wallet com senha
func EncryptWallet(walletPath string, password string) error {
    // Ler wallet
    data, err := os.ReadFile(filepath.Join(walletPath, "wallet.yaml"))
    if err != nil {
        return err
    }

    // Derivar chave da senha usando PBKDF2
    salt := make([]byte, 32)
    rand.Read(salt)
    key := pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)

    // Criptografar com AES-256-GCM
    block, err := aes.NewCipher(key)
    if err != nil {
        return err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return err
    }

    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)

    ciphertext := gcm.Seal(nonce, nonce, data, nil)

    // Salvar arquivo criptografado
    encrypted := append(salt, ciphertext...)
    return os.WriteFile(filepath.Join(walletPath, "wallet.encrypted"), encrypted, 0600)
}

// DecryptWallet descriptografa wallet
func DecryptWallet(walletPath string, password string) error {
    // Ler arquivo criptografado
    encrypted, err := os.ReadFile(filepath.Join(walletPath, "wallet.encrypted"))
    if err != nil {
        return err
    }

    // Extrair salt
    salt := encrypted[:32]
    ciphertext := encrypted[32:]

    // Derivar chave
    key := pbkdf2.Key([]byte(password), salt, 100000, 32, sha256.New)

    // Descriptografar
    block, err := aes.NewCipher(key)
    if err != nil {
        return err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return err
    }

    nonceSize := gcm.NonceSize()
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return fmt.Errorf("senha incorreta")
    }

    // Salvar descriptografado
    return os.WriteFile(filepath.Join(walletPath, "wallet.yaml"), plaintext, 0600)
}
```

### 2. Modo Anonimizado

```go
// AnonymizeWallet gera versão anonimizada para demonstrações
func (w *Wallet) AnonymizeWallet() *Wallet {
    anon := &Wallet{
        Assets: make(map[string]*Asset),
    }

    // Substituir valores reais por fictícios
    for ticker, asset := range w.Assets {
        anonAsset := &Asset{
            ID:                 ticker,  // Mantém ticker
            Type:               asset.Type,
            SubType:            asset.SubType,
            Segment:            asset.Segment,
            Quantity:           asset.Quantity,
            // Valores multiplicados por fator aleatório
            AveragePrice:       randomizeDec(asset.AveragePrice),
            TotalInvestedValue: randomizeDec(asset.TotalInvestedValue),
            TotalEarnings:      randomizeDec(asset.TotalEarnings),
        }

        anon.Assets[ticker] = anonAsset
    }

    return anon
}

func randomizeDec(val decimal.Decimal) decimal.Decimal {
    // Multiplicar por fator aleatório entre 0.8 e 1.2
    factor := 0.8 + rand.Float64()*0.4
    return val.Mul(decimal.NewFromFloat(factor))
}
```

### 3. Auditoria de Operações

```go
type AuditLog struct {
    Timestamp time.Time
    Operation string  // "add_transaction", "edit_asset", etc
    User      string
    Details   map[string]interface{}
    BeforeSnapshot string  // Hash do estado anterior
    AfterSnapshot  string  // Hash do estado posterior
}

// LogOperation registra operação no audit log
func (w *Wallet) LogOperation(operation string, details map[string]interface{}) {
    beforeHash := w.CalculateStateHash()

    log := AuditLog{
        Timestamp:      time.Now(),
        Operation:      operation,
        Details:        details,
        BeforeSnapshot: beforeHash,
    }

    // Executar operação aqui...

    log.AfterSnapshot = w.CalculateStateHash()

    w.AuditLogs = append(w.AuditLogs, log)
}

// Rollback desfaz operação
func (w *Wallet) Rollback(logID int) error {
    // Restaurar do snapshot anterior
    // ...
}
```

---

## 🎨 Comandos CLI

```bash
# Criptografia
b3cli wallet encrypt --password
b3cli wallet decrypt --password

# Anonimização
b3cli privacy anonymize --output=demo-wallet/
b3cli privacy hide-values  # Oculta valores na TUI

# Auditoria
b3cli audit log
b3cli audit rollback --to=OPERATION_ID
```

---

## 📊 TUI - Modo Privacidade:

```
╔══════════════════════════════════════════════════════════════════════╗
║              📊 CARTEIRA (MODO PRIVACIDADE)                          ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Ticker    Qtd    Investido      Atual         Lucro/Prejuízo       ║
║  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━    ║
║  PETR4     200    R$ •••••       R$ •••••      R$ ••••• (••%)       ║
║  VALE3     100    R$ •••••       R$ •••••      R$ ••••• (••%)       ║
║  ITSA4     500    R$ •••••       R$ •••••      R$ ••••• (••%)       ║
║                                                                      ║
║  💡 Modo privacidade ativo                                           ║
║     Pressione [P] para revelar valores                              ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝
```

---

## 📊 Valor para o Usuário

- 🔒 **Segurança:** Dados protegidos com AES-256
- 🎭 **Privacy:** Pode compartilhar sem expor valores
- 📝 **Auditoria:** Histórico completo de alterações
- ↩️ **Rollback:** Desfazer mudanças acidentais

---

**Estimativa de implementação:** 1 semana
**ROI para usuários:** Baixo-Médio (para usuários preocupados com segurança)
