## 1. Feature Documentation

- [x] 1.1 Create `docs/features/economy.md` — P2P Economy feature page with experimental warning, architecture diagram, 5 subsystem sections, config block
- [x] 1.2 Create `docs/features/contracts.md` — Smart Contracts feature page with ABI cache, read/write ops, agent tools, config
- [x] 1.3 Create `docs/features/observability.md` — Observability feature page with metrics, token tracking, health, audit, API endpoints

## 2. CLI Documentation

- [x] 2.1 Create `docs/cli/economy.md` — Economy CLI reference with 5 subcommand sections, flags tables, example output
- [x] 2.2 Create `docs/cli/contract.md` — Contract CLI reference with read, call, abi load sections
- [x] 2.3 Create `docs/cli/metrics.md` — Metrics CLI reference with 5 subcommands, persistent flags

## 3. Documentation Index Updates

- [x] 3.1 Update `docs/features/index.md` — Add 3 feature cards and 3 status table rows for Economy, Contracts, Observability
- [x] 3.2 Update `docs/cli/index.md` — Add Economy (5 cmds), Contract (3 cmds), Metrics (5 cmds) tables
- [x] 3.3 Update `docs/configuration.md` — Add Economy and Observability config sections with JSON blocks and key tables
- [x] 3.4 Update `mkdocs.yml` — Add 6 nav entries (3 Features, 3 CLI Reference)

## 4. Prompts & README

- [x] 4.1 Update `prompts/TOOL_USAGE.md` — Add Economy Tool (13 tools) and Contract Tool (3 tools) sections, update exec blocklist
- [x] 4.2 Update `prompts/AGENTS.md` — Change tool count to thirteen, add Economy/Contract/Observability bullets
- [x] 4.3 Update `README.md` — Add features, CLI commands, architecture tree entries

## 5. TUI Settings Forms

- [x] 5.1 Create `internal/cli/settings/forms_economy.go` — 5 economy form constructors (base, risk, negotiation, escrow, pricing)
- [x] 5.2 Create `internal/cli/settings/forms_observability.go` — Observability form constructor
- [x] 5.3 Update `internal/cli/settings/menu.go` — Add Economy section (5 categories) and Observability to Infrastructure
- [x] 5.4 Update `internal/cli/settings/editor.go` — Add 6 new cases in handleMenuSelection()
- [x] 5.5 Update `internal/cli/tuicore/state_update.go` — Add ~30 economy/observability case statements + parseFloatSlice helper

## 6. Doctor Health Checks

- [x] 6.1 Create `internal/cli/doctor/checks/economy.go` — EconomyCheck with budget/risk/escrow/negotiate/pricing validation
- [x] 6.2 Create `internal/cli/doctor/checks/contract.go` — ContractCheck with rpcURL/chainID validation
- [x] 6.3 Create `internal/cli/doctor/checks/observability.go` — ObservabilityCheck with retention/interval validation
- [x] 6.4 Update `internal/cli/doctor/checks/checks.go` — Register 3 new checks in AllChecks()

## 7. Verification

- [x] 7.1 Run `go build ./...` — Verify all Go code compiles
- [x] 7.2 Run `go test ./...` — Verify all tests pass
