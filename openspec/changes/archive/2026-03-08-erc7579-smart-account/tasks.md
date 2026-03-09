# Tasks: ERC-7579 Smart Account

## Solidity Contracts

- [x] 1.1 Create ISessionValidator interface (`contracts/src/modules/ISessionValidator.sol`)
- [x] 1.2 Implement LangoSessionValidator module (`contracts/src/modules/LangoSessionValidator.sol`)
- [x] 1.3 Implement LangoSpendingHook module (`contracts/src/modules/LangoSpendingHook.sol`)
- [x] 1.4 Implement LangoEscrowExecutor module (`contracts/src/modules/LangoEscrowExecutor.sol`)
- [x] 1.5 Write Foundry tests for SessionValidator (`contracts/test/LangoSessionValidator.t.sol`) — 20 tests
- [x] 1.6 Write Foundry tests for SpendingHook (`contracts/test/LangoSpendingHook.t.sol`) — 17 tests
- [x] 1.7 Write Foundry tests for EscrowExecutor (`contracts/test/LangoEscrowExecutor.t.sol`) — 11 tests

## Go Core Types & Config

- [x] 2.1 Create smartaccount package with doc.go, types.go, errors.go
- [x] 2.2 Create SmartAccountConfig in `internal/config/types_smartaccount.go`
- [x] 2.3 Add SmartAccount field to Config struct in `internal/config/types.go`
- [x] 2.4 Create UserOpSigner interface and LocalUserOpSigner in `internal/wallet/userop.go`
- [x] 2.5 Write tests for LocalUserOpSigner (`internal/wallet/userop_test.go`)

## Session Key Management

- [x] 3.1 Create session.Store interface and MemoryStore (`internal/smartaccount/session/store.go`)
- [x] 3.2 Create session key crypto helpers (`internal/smartaccount/session/crypto.go`)
- [x] 3.3 Implement session.Manager with Create/Revoke/SignUserOp (`internal/smartaccount/session/manager.go`)
- [x] 3.4 Write MemoryStore tests (`internal/smartaccount/session/store_test.go`) — 11 tests
- [x] 3.5 Write Manager tests (`internal/smartaccount/session/manager_test.go`) — 14 tests

## Policy Engine

- [x] 4.1 Define HarnessPolicy and SpendTracker types (`internal/smartaccount/policy/types.go`)
- [x] 4.2 Implement Validator.Check (`internal/smartaccount/policy/validator.go`)
- [x] 4.3 Implement policy.Engine (`internal/smartaccount/policy/engine.go`)
- [x] 4.4 Write Validator tests (`internal/smartaccount/policy/validator_test.go`) — 10 tests
- [x] 4.5 Write Engine tests (`internal/smartaccount/policy/engine_test.go`) — 12 tests

## Account Manager & Bundler

- [x] 5.1 Implement bundler.Client JSON-RPC (`internal/smartaccount/bundler/client.go`)
- [x] 5.2 Define bundler types (`internal/smartaccount/bundler/types.go`)
- [x] 5.3 Implement Factory for Safe deployment (`internal/smartaccount/factory.go`)
- [x] 5.4 Implement Manager (AccountManager interface) (`internal/smartaccount/manager.go`)
- [x] 5.5 Write bundler client tests (`internal/smartaccount/bundler/client_test.go`) — 6 tests
- [x] 5.6 Write manager tests (`internal/smartaccount/manager_test.go`) — 8 tests

## Module Registry

- [x] 6.1 Define ModuleDescriptor type (`internal/smartaccount/module/types.go`)
- [x] 6.2 Implement Registry (`internal/smartaccount/module/registry.go`)
- [x] 6.3 Implement ABI encoder (`internal/smartaccount/module/abi_encoder.go`)
- [x] 6.4 Write Registry tests (`internal/smartaccount/module/registry_test.go`) — 9 tests

## ABI Bindings

- [x] 7.1 Create ParseABI helper (`internal/smartaccount/bindings/abi.go`)
- [x] 7.2 Create SessionValidatorClient (`internal/smartaccount/bindings/session_validator.go`)
- [x] 7.3 Create SpendingHookClient (`internal/smartaccount/bindings/spending_hook.go`)
- [x] 7.4 Create EscrowExecutorClient (`internal/smartaccount/bindings/escrow_executor.go`)
- [x] 7.5 Create Safe7579Client (`internal/smartaccount/bindings/safe7579.go`)

## App Wiring & Tools

- [x] 8.1 Create wiring_smartaccount.go with initSmartAccount()
- [x] 8.2 Create tools_smartaccount.go with 10 agent tools
- [x] 8.3 Add SmartAccountManager field to App struct in types.go
- [x] 8.4 Add step 5p' in app.go for smart account initialization

## CLI Commands

- [x] 9.1 Create smartaccount.go root command (`internal/cli/smartaccount/`)
- [x] 9.2 Create deploy.go (`lango account deploy`)
- [x] 9.3 Create info.go (`lango account info`)
- [x] 9.4 Create session.go (`lango account session create/list/revoke`)
- [x] 9.5 Create module.go (`lango account module list/install`)
- [x] 9.6 Create policy.go (`lango account policy show/set`)
- [x] 9.7 Register account command in cmd/lango/main.go

## Economy Integration

- [x] 10.1 Create OnChainTracker (`internal/economy/budget/onchain.go`)
- [x] 10.2 Create PolicyAdapter (`internal/economy/risk/policy_adapter.go`)
- [x] 10.3 Create SessionGuard (`internal/economy/escrow/sentinel/session_guard.go`)
- [x] 10.4 Write OnChainTracker tests (`internal/economy/budget/onchain_test.go`)
- [x] 10.5 Write PolicyAdapter tests (`internal/economy/risk/policy_adapter_test.go`)
- [x] 10.6 Write SessionGuard tests (`internal/economy/escrow/sentinel/session_guard_test.go`)

## Verification

- [x] 11.1 `go build ./...` passes
- [x] 11.2 `go test ./...` all pass (42 new smartaccount tests + existing tests)
- [x] 11.3 `forge build` compiles all Solidity contracts
- [x] 11.4 `forge test` all 121 Foundry tests pass (48 new module tests)
