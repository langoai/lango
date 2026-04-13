## 1. Settings Surface

- [x] 1.1 Add `runledger` category to the settings menu under `Automation`
- [x] 1.2 Add `NewRunLedgerForm(cfg)` in `internal/cli/settings/forms_automation.go`
- [x] 1.3 Wire `runledger` → `NewRunLedgerForm` in `setup_flow.go`
- [x] 1.4 Add `ConfigState.UpdateConfigFromForm` mappings for all RunLedger form keys
- [x] 1.5 Add form tests covering field presence and save/update behavior

## 2. Doctor Surface

- [x] 2.1 Add `RunLedgerCheck` in `internal/cli/doctor/checks/`
- [x] 2.2 Register `RunLedgerCheck` in `checks.AllChecks()`
- [x] 2.3 Update `lango doctor` long description to include RunLedger diagnostics
- [x] 2.4 Add doctor tests covering skip/fail/pass cases

## 3. Downstream Sync

- [x] 3.1 Update settings command help text to mention RunLedger in Automation coverage
- [x] 3.2 Update README and relevant CLI docs for settings/doctor RunLedger coverage

## 4. Verification

- [x] 4.1 Run `go build ./...`
- [x] 4.2 Run `go test ./internal/cli/settings/... ./internal/cli/doctor/...`
- [x] 4.3 Run `go test ./...`
