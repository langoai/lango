## Why

The `feature/p2p-escrow` branch added on-chain escrow (Hub/Vault dual-mode), Security Sentinel (anomaly detection), P2P Settlement, and Team Coordination enhancements (+14,048 lines, 88 files). However, all downstream artifacts — docs, README, CLI docs, TUI settings, and system prompts — were NOT updated. Users cannot discover or configure these features without accurate documentation and UI surfaces.

## What Changes

- **System Prompts**: Replace old `economy_escrow_*` tool names with new `escrow_*` (10 tools) and add `sentinel_*` (4 tools) in TOOL_USAGE.md
- **Feature Docs**: Expand economy.md with on-chain escrow (Hub/Vault), Security Sentinel, and 6 new events; expand contracts.md with Foundry contract details; expand p2p-network.md with team coordination enhancements
- **CLI Docs**: Add `escrow list`, `escrow show`, `escrow sentinel status` commands to economy.md; enhance p2p.md with team features
- **Configuration Docs**: Add 10 on-chain escrow config keys (`economy.escrow.onChain.*`, `economy.escrow.settlement.*`)
- **TUI Settings**: New `NewEconomyEscrowOnChainForm()` with 10 fields, new menu category, editor wiring
- **README**: Update features, CLI commands, architecture tree with contracts/, hub/, sentinel/ directories

## Capabilities

### New Capabilities

(None — all capabilities already exist as specs; this change only updates documentation/UI artifacts)

### Modified Capabilities

- `onchain-escrow`: Documentation added for on-chain escrow (economy.md, contracts.md, configuration.md, TOOL_USAGE.md, README)
- `p2p-team-coordination`: Documentation expanded with conflict resolution, assignment strategies, payment coordination
- `p2p-settlement`: Documentation added for P2P settlement workflow

## Impact

- **Files changed**: 11 files (+510 lines)
- **Code**: `internal/cli/settings/forms_economy.go`, `menu.go`, `editor.go` (TUI form + wiring)
- **Docs**: `prompts/TOOL_USAGE.md`, `docs/features/economy.md`, `docs/features/contracts.md`, `docs/features/p2p-network.md`, `docs/cli/economy.md`, `docs/cli/p2p.md`, `docs/configuration.md`
- **README**: `README.md` (features, CLI, architecture)
- **No breaking changes** — purely additive documentation and UI updates
