## 1. Solidity Contracts

- [x] 1.1 Create Foundry project structure (foundry.toml, .gitignore)
- [x] 1.2 Implement IERC20 interface and MockUSDC test token
- [x] 1.3 Implement LangoEscrowHub contract (deal lifecycle, events, modifiers)
- [x] 1.4 Implement LangoVault contract (single-deal vault, initializable for EIP-1167)
- [x] 1.5 Implement LangoVaultFactory contract (EIP-1167 minimal proxy cloning)
- [x] 1.6 Create Deploy.s.sol deployment script

## 2. Go ABI Package + Typed Clients

- [x] 2.1 Create ABI JSON files from contract interfaces (Hub, Vault, Factory)
- [x] 2.2 Implement abi.go with go:embed directives and Parse*ABI() helpers
- [x] 2.3 Implement types.go (OnChainDealStatus, OnChainDeal, VaultInfo)
- [x] 2.4 Implement HubClient wrapping contract.Caller
- [x] 2.5 Implement VaultClient wrapping contract.Caller
- [x] 2.6 Implement FactoryClient wrapping contract.Caller

## 3. Settlement Executors + Config

- [x] 3.1 Add EscrowOnChainConfig to internal/config/types_economy.go
- [x] 3.2 Implement HubSettler (SettlementExecutor) with deal mapping
- [x] 3.3 Implement VaultSettler (SettlementExecutor) with vault creation
- [x] 3.4 Add selectSettler() function to wiring_economy.go

## 4. Persistent Escrow Store

- [x] 4.1 Create Ent schema escrow_deal.go with on-chain tracking fields
- [x] 4.2 Run go generate for Ent code generation
- [x] 4.3 Implement EntStore (Store interface + on-chain methods)
- [x] 4.4 Write EntStore tests with in-memory SQLite

## 5. Event Monitor

- [x] 5.1 Add 6 on-chain event types to eventbus/economy_events.go
- [x] 5.2 Implement EventMonitor with eth_getLogs polling and event decoding
- [x] 5.3 Implement Start/Stop lifecycle for EventMonitor

## 6. Security Sentinel Engine

- [x] 6.1 Implement sentinel types (Alert, SentinelConfig, Detector interface)
- [x] 6.2 Implement 5 anomaly detectors (rapid creation, large withdrawal, repeated dispute, unusual timing, balance drop)
- [x] 6.3 Implement Sentinel engine (Start/Stop, event subscriptions, alert management)
- [x] 6.4 Write detector tests (table-driven, all 5 detectors)
- [x] 6.5 Write engine tests (lifecycle, detection, acknowledge, status)

## 7. Agent Tools

- [x] 7.1 Implement 10 escrow tools in tools_escrow.go (buildOnChainEscrowTools)
- [x] 7.2 Implement 4 sentinel tools in tools_sentinel.go (buildSentinelTools)
- [x] 7.3 Create security-sentinel.yaml skill definition
- [x] 7.4 Write tool tests for escrow and sentinel tools

## 8. CLI Commands

- [x] 8.1 Add `lango economy escrow list` subcommand
- [x] 8.2 Add `lango economy escrow show` subcommand with --id flag
- [x] 8.3 Add `lango economy escrow sentinel status` subcommand

## 9. App Wiring + Integration

- [x] 9.1 Wire sentinel engine init in initEconomy() after escrow engine
- [x] 9.2 Register escrow and sentinel tool categories in app.go catalog
- [x] 9.3 Create OpenSpec spec.md and delta.md documentation

## 10. OpenSpec Documentation

- [x] 10.1 Create openspec/specs/onchain-escrow/spec.md
- [x] 10.2 Create openspec/specs/onchain-escrow/delta.md
