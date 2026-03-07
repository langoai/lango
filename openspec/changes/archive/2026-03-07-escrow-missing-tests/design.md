## Context

The on-chain escrow system (Hub, Vault, Factory contracts + Go clients + settlers + event monitor) was implemented without tests due to:
1. Foundry/Anvil not installed on the development machine
2. `contract.Caller` being a concrete struct, making Go unit test mocking impossible

The system is functionally complete but lacks any test coverage — a significant quality risk.

## Goals / Non-Goals

**Goals:**
- Extract `ContractCaller` interface from `contract.Caller` for dependency injection
- Achieve comprehensive Solidity forge test coverage for all 3 contracts
- Achieve comprehensive Go unit test coverage for all hub package types
- Provide Anvil-based integration tests for full E2E validation
- Zero regression on existing tests

**Non-Goals:**
- Changing any business logic or contract behavior
- Adding fuzz testing or formal verification
- CI/CD pipeline integration for Anvil tests
- Gas optimization or contract upgrades

## Decisions

### D1: Interface extraction over test doubles generation
**Decision**: Extract a `ContractCaller` interface with `Read`/`Write` methods from the existing `Caller` struct.
**Rationale**: The concrete struct has RPC client, wallet, nonce mutex — all unsuitable for unit tests. An interface allows simple mock implementations. The existing `*Caller` satisfies the interface automatically, so no caller-site changes needed.
**Alternative considered**: Using build tags to swap implementations — rejected as overly complex for this use case.

### D2: Package-internal mocks over generated mocks
**Decision**: Hand-written `mockCaller` in `mock_test.go` rather than using mockgen/gomock.
**Rationale**: The interface has only 2 methods. Hand-written mocks are simpler, more readable, and avoid a new dependency. The mock supports configurable results and call recording.

### D3: Build tag for integration tests
**Decision**: Use `//go:build integration` tag for Anvil-dependent tests.
**Rationale**: Integration tests require a running Anvil instance. Build tags ensure `go test ./...` never fails due to missing infrastructure. Developers opt-in with `-tags integration`.

### D4: Forge artifacts for contract deployment in integration tests
**Decision**: Read compiled bytecode from `contracts/out/` (forge build output) at test time.
**Rationale**: Avoids embedding large bytecode blobs in Go source. Requires `forge build` before integration tests, which is documented.

## Risks / Trade-offs

- [Risk] Integration tests depend on Anvil being available → Mitigated by build tag; CI can skip or provision Anvil
- [Risk] forge-std as git submodule adds repo size → Mitigated by `.gitignore` for `contracts/lib/`; developers install via `forge install`
- [Trade-off] Mock-based unit tests don't verify ABI encoding correctness → Integration tests cover this gap
