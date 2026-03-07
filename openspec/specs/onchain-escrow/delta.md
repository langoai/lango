# On-Chain Escrow — Change Delta

## New Files

### Solidity Contracts
- `contracts/foundry.toml` — Foundry project config
- `contracts/src/LangoEscrowHub.sol` — Master escrow hub
- `contracts/src/LangoVault.sol` — Individual vault (EIP-1167 target)
- `contracts/src/LangoVaultFactory.sol` — Vault factory
- `contracts/src/interfaces/IERC20.sol` — ERC-20 interface
- `contracts/test/mocks/MockUSDC.sol` — Test mock
- `contracts/script/Deploy.s.sol` — Deployment script

### Go ABI + Clients
- `internal/economy/escrow/hub/abi.go` — Embedded ABI + parsing helpers
- `internal/economy/escrow/hub/types.go` — OnChainDeal, VaultInfo types
- `internal/economy/escrow/hub/client.go` — HubClient (typed hub operations)
- `internal/economy/escrow/hub/vault_client.go` — VaultClient (typed vault operations)
- `internal/economy/escrow/hub/factory_client.go` — FactoryClient (vault creation)
- `internal/economy/escrow/hub/abi/*.abi.json` — 3 ABI JSON files

### Settlers
- `internal/economy/escrow/hub/hub_settler.go` — HubSettler (SettlementExecutor)
- `internal/economy/escrow/hub/vault_settler.go` — VaultSettler (SettlementExecutor)

### Event Monitor
- `internal/economy/escrow/hub/monitor.go` — Polling-based event monitor

### Sentinel Engine
- `internal/economy/escrow/sentinel/types.go` — Alert, SentinelConfig, Detector interface
- `internal/economy/escrow/sentinel/detector.go` — 5 anomaly detectors
- `internal/economy/escrow/sentinel/engine.go` — Sentinel engine
- `internal/economy/escrow/sentinel/engine_test.go` — Tests
- `internal/economy/escrow/sentinel/detector_test.go` — Tests

### Agent Tools
- `internal/app/tools_escrow.go` — 10 escrow tools + 4 sentinel tools
- `internal/app/tools_sentinel.go` — Sentinel tools

### Skill
- `skills/security-sentinel.yaml` — Sentinel skill definition

## Modified Files

### Config
- `internal/config/types_economy.go` — Added `EscrowOnChainConfig` sub-struct

### Event Bus
- `internal/eventbus/economy_events.go` — Added 6 on-chain event types

### Wiring
- `internal/app/wiring_economy.go` — Added `selectSettler()`, sentinel engine init
- `internal/app/app.go` — Registered escrow + sentinel tool categories

### CLI
- `internal/cli/economy/escrow.go` — Expanded with list, show, sentinel subcommands

## Backward Compatibility

- Fully backward compatible: existing `custodian` mode unchanged
- New config under `economy.escrow.onChain` (additive)
- Existing escrow tools in `tools_economy.go` unchanged
- New tools registered in separate catalog categories ("escrow", "sentinel")
