## 1. Prerequisites

- [x] 1.1 Install Foundry toolchain (forge, anvil, cast, chisel)
- [x] 1.2 Install forge-std library in contracts/lib/
- [x] 1.3 Add remappings to contracts/foundry.toml
- [x] 1.4 Add contracts/out/, contracts/cache/, contracts/lib/ to .gitignore

## 2. ContractCaller Interface Extraction

- [x] 2.1 Add ContractCaller interface (Read/Write) to internal/contract/caller.go
- [x] 2.2 Add compile-time interface check: var _ ContractCaller = (*Caller)(nil)
- [x] 2.3 Update HubClient to accept contract.ContractCaller
- [x] 2.4 Update VaultClient to accept contract.ContractCaller
- [x] 2.5 Update FactoryClient to accept contract.ContractCaller
- [x] 2.6 Update HubSettler constructor to accept contract.ContractCaller
- [x] 2.7 Update VaultSettler constructor and field to accept contract.ContractCaller
- [x] 2.8 Verify go build ./... passes with no errors

## 3. Solidity Forge Tests

- [x] 3.1 Create contracts/test/LangoEscrowHub.t.sol with ~38 test cases
- [x] 3.2 Create contracts/test/LangoVault.t.sol with ~26 test cases
- [x] 3.3 Create contracts/test/LangoVaultFactory.t.sol with ~9 test cases
- [x] 3.4 Verify forge test -vvv passes all Solidity tests

## 4. Go Unit Tests — Shared Mock

- [x] 4.1 Create hub/mock_test.go with mockCaller and mockOnChainStore

## 5. Go Unit Tests — Types and ABI

- [x] 5.1 Create hub/abi_test.go testing ParseHubABI, ParseVaultABI, ParseFactoryABI
- [x] 5.2 Create hub/types_test.go testing OnChainDealStatus.String() for all 7 statuses + unknown

## 6. Go Unit Tests — Clients

- [x] 6.1 Create hub/client_test.go testing all 9 HubClient methods (success + error)
- [x] 6.2 Create hub/vault_client_test.go testing all 8 VaultClient methods (success + error)
- [x] 6.3 Create hub/factory_client_test.go testing all 3 FactoryClient methods (success + error + edge cases)

## 7. Go Unit Tests — Settlers

- [x] 7.1 Create hub/hub_settler_test.go testing interface compliance, mapping, no-ops, accessors, concurrency
- [x] 7.2 Create hub/vault_settler_test.go testing interface compliance, mapping, CreateVault, VaultClientFor, concurrency

## 8. Go Unit Tests — Monitor

- [x] 8.1 Create hub/monitor_test.go testing helper functions, resolveEscrowID, handleEvent (6 types), processLog edge cases

## 9. Go Integration Tests

- [x] 9.1 Create hub/integration_test.go with //go:build integration tag and 7 E2E test cases

## 10. Verification

- [x] 10.1 Verify go build ./... passes
- [x] 10.2 Verify go test ./... passes with zero failures (no regressions)
- [x] 10.3 Verify forge test -vvv passes all Solidity tests
