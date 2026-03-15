## 1. EntryPoint v0.6 → v0.7 Address Migration

- [x] 1.1 Replace v0.6 address in `docs/features/smart-accounts.md` (1 occurrence)
- [x] 1.2 Replace v0.6 address in `docs/cli/smartaccount.md` (3 occurrences)
- [x] 1.3 Replace v0.6 address in `internal/wallet/userop_test.go` (2 occurrences)
- [x] 1.4 Replace v0.6 address in `internal/smartaccount/manager_test.go` (6 occurrences)
- [x] 1.5 Replace v0.6 address in `internal/smartaccount/bundler/client_test.go` (5 occurrences)
- [x] 1.6 Replace v0.6 address in `internal/smartaccount/integration_test.go` (5 occurrences)
- [x] 1.7 Replace v0.6 address in paymaster test files (circle, pimlico, alchemy, approve — 5 occurrences)
- [x] 1.8 Verify zero matches for v0.6 address in entire codebase

## 2. CLI Mode Field

- [x] 2.1 Add `Mode` to `statusInfo` struct and table/JSON output in `paymaster.go`
- [x] 2.2 Add permit mode support to `initPaymasterProvider()` in `deps.go`

## 3. TUI Mode Field

- [x] 3.1 Add `sa_paymaster_mode` select field to `forms_smartaccount.go`
- [x] 3.2 Add `sa_paymaster_mode` case to `state_update.go`

## 4. Documentation

- [x] 4.1 Update `docs/features/smart-accounts.md` — provider table, config table, config example
- [x] 4.2 Update `docs/cli/smartaccount.md` — paymaster status output example
- [x] 4.3 Update `docs/configuration.md` — add `smartAccount.paymaster.mode` row

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 All smartaccount, wallet, cli tests pass
