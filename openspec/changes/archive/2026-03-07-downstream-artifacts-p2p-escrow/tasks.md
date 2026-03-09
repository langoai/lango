## 1. System Prompts

- [x] 1.1 Update `prompts/TOOL_USAGE.md` — replace old `economy_escrow_*` tool names with 10 new `escrow_*` tools
- [x] 1.2 Add 4 `sentinel_*` tools (sentinel_status, sentinel_alerts, sentinel_config, sentinel_acknowledge) to TOOL_USAGE.md
- [x] 1.3 Add on-chain workflow guidance (create → fund → activate → submit_work → release/dispute → resolve)

## 2. Feature Documentation

- [x] 2.1 Expand `docs/features/economy.md` with On-Chain Escrow section (Hub vs Vault modes, deal lifecycle)
- [x] 2.2 Add Security Sentinel subsection to economy.md (5 detectors, alert severity, config)
- [x] 2.3 Add 6 new on-chain events to Events Summary table in economy.md
- [x] 2.4 Expand `docs/features/contracts.md` with LangoEscrowHub, LangoVault, LangoVaultFactory details and Foundry build/test instructions
- [x] 2.5 Expand `docs/features/p2p-network.md` with conflict resolution strategies, assignment strategies, payment coordination, team events

## 3. CLI Documentation

- [x] 3.1 Add `lango economy escrow list`, `escrow show`, `escrow sentinel status` commands to `docs/cli/economy.md`
- [x] 3.2 Add team coordination features (conflict resolution, assignment, payment) notes to `docs/cli/p2p.md`

## 4. Configuration Documentation

- [x] 4.1 Add 10 on-chain escrow config keys to `docs/configuration.md` (`economy.escrow.onChain.*`, `economy.escrow.settlement.*`)
- [x] 4.2 Update JSON/YAML example block in configuration.md with settlement and onChain sections

## 5. TUI Settings (Code)

- [x] 5.1 Add `NewEconomyEscrowOnChainForm()` with 10 fields in `internal/cli/settings/forms_economy.go`
- [x] 5.2 Add `economy_escrow_onchain` menu category to Economy section in `internal/cli/settings/menu.go`
- [x] 5.3 Wire `economy_escrow_onchain` case in `handleMenuSelection` in `internal/cli/settings/editor.go`

## 6. README

- [x] 6.1 Update `README.md` features section with on-chain escrow, Security Sentinel, Foundry contracts, P2P Teams
- [x] 6.2 Add escrow CLI commands to README CLI section
- [x] 6.3 Add `contracts/`, `escrow/hub/`, `escrow/sentinel/` to architecture tree in README

## 7. Verification

- [x] 7.1 Run `go build ./...` to verify TUI code compiles
- [x] 7.2 Run `go test ./internal/cli/settings/...` to verify settings tests pass
