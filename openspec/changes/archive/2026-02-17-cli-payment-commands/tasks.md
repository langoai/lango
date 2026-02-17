## 1. Payment CLI Package Setup

- [x] 1.1 Create `internal/cli/payment/payment.go` with `NewPaymentCmd()`, `paymentDeps` struct, and `initPaymentDeps()` using bootLoader pattern
- [x] 1.2 Register payment command in `cmd/lango/main.go` with bootLoader closure

## 2. Read-Only Subcommands

- [x] 2.1 Create `internal/cli/payment/balance.go` — `newBalanceCmd()` with `--json` flag
- [x] 2.2 Create `internal/cli/payment/history.go` — `newHistoryCmd()` with `--json` and `--limit` flags
- [x] 2.3 Create `internal/cli/payment/limits.go` — `newLimitsCmd()` with `--json` flag
- [x] 2.4 Create `internal/cli/payment/info.go` — `newInfoCmd()` with `--json` flag

## 3. Send Subcommand

- [x] 3.1 Create `internal/cli/payment/send.go` — `newSendCmd()` with `--to`, `--amount`, `--purpose`, `--force`, `--json` flags and interactive confirmation prompt

## 4. Build and Test Verification

- [x] 4.1 Run `go build ./...` and verify no compilation errors
- [x] 4.2 Run `go test ./...` and verify all existing tests pass

## 5. README Documentation

- [x] 5.1 Add blockchain payments to Features list
- [x] 5.2 Add `lango payment` commands to CLI Commands section
- [x] 5.3 Add `payment/`, `wallet/`, `x402/` to architecture directory tree and `cli/payment/` to CLI tree
- [x] 5.4 Add `payment.*` configuration rows to Configuration Reference table
- [x] 5.5 Update Multi-Agent Orchestration table with `payment_*` tools for executor
- [x] 5.6 Add new "Blockchain Payments" section after A2A Protocol section (payment tools, wallet providers, X402 protocol, CLI usage, configuration)
- [x] 5.7 Add "Wallet Key Security" subsection to Security section
- [x] 5.8 Add payment to onboard wizard guide list

## 6. Onboard TUI Integration

- [x] 6.1 Add "payment" category to onboard menu (`menu.go`)
- [x] 6.2 Create `NewPaymentForm()` with all payment config fields (`forms_impl.go`)
- [x] 6.3 Add "payment" case to `handleMenuSelection()` (`wizard.go`)
- [x] 6.4 Add payment field mappings to `UpdateConfigFromForm()` (`state_update.go`)
