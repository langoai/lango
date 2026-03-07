## Why

The on-chain escrow system was implemented without tests because Foundry/Anvil were not installed and `contract.Caller` was a concrete struct preventing mocking. This change adds all missing tests and extracts a `ContractCaller` interface to enable unit testing.

## What Changes

- Extract `ContractCaller` interface from `contract.Caller` struct for mockability
- Update all hub package clients/settlers to accept the interface instead of concrete `*Caller`
- Add 3 Solidity forge test files (73 test cases) covering Hub, Vault, and Factory contracts
- Add 9 Go unit test files (80 test cases) covering all hub package types
- Add 1 Go integration test file (7 test cases) for Anvil E2E testing
- Install Foundry toolchain and forge-std dependency
- Add `remappings` to `foundry.toml`

## Capabilities

### New Capabilities
- `escrow-test-coverage`: Comprehensive test coverage for on-chain escrow contracts and Go clients including Solidity forge tests, Go unit tests with mock caller, and Anvil integration tests

### Modified Capabilities
- `contract-interaction`: Extract `ContractCaller` interface from concrete `Caller` struct to enable dependency injection and mocking
- `onchain-escrow`: Update client/settler constructors to accept `ContractCaller` interface instead of `*Caller`

## Impact

- `internal/contract/caller.go` — new `ContractCaller` interface
- `internal/economy/escrow/hub/*.go` — field types and constructor params changed to interface
- `contracts/foundry.toml` — remappings added
- `contracts/test/` — 3 new Solidity test files
- `internal/economy/escrow/hub/*_test.go` — 10 new Go test files
- `.gitignore` — forge build artifacts excluded
- No breaking changes for external callers (concrete `*Caller` satisfies the interface)
