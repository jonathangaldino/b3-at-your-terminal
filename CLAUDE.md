# Memory

If you read any file and texts, titles or descriptions are in portuguese-brazilian,
you can translate them to english when coding the implementation that will use the data extract from the file(s) you read.

# Architecture

## Code Organization Rules

### CLI Layer (`cmd/b3cli/`)
- All CLI commands and user interface code
- Cobra command definitions
- Bubble Tea TUI implementations
- User input/output handling
- **NO business logic** - only orchestration and presentation
- Must use `package main`

### Business Logic Layer (`internal/wallet/`)
- All wallet operations and business rules
- Transaction management (add, validate, deduplicate)
- Asset calculations and queries
- Data persistence (YAML save/load)
- **NO CLI/UI code** - pure Go functions and methods

### Separation Principles
1. CLI commands call wallet methods, never implement business logic directly
2. Wallet package has no knowledge of CLI or user interface
3. All validation happens in wallet layer, not CLI
4. CLI only handles: user input → wallet method → format output
5. Business logic is fully testable without any CLI dependencies

## Examples

### Good Practice ✓
```go
// cmd/b3cli/parse.go
added, duplicates, err := w.AddTransactions(newTransactions)
if err != nil {
    return fmt.Errorf("erro ao adicionar transações: %w", err)
}
```

### Bad Practice ✗
```go
// cmd/b3cli/parse.go
for _, t := range newTransactions {
    if _, exists := w.TransactionsByHash[t.Hash]; !exists {
        w.Transactions = append(w.Transactions, t)
        // ... more business logic in CLI
    }
}
```
