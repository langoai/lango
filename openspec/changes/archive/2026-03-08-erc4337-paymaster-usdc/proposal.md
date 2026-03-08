## Why

Users must hold ETH to pay gas for every UserOperation on the smart account, creating friction for USDC-native workflows. By integrating ERC-4337 paymaster support (Circle Paymaster as primary, Pimlico/Alchemy as alternatives), users can pay gas in USDC — enabling fully gasless transactions on Base chain.

## What Changes

- Add `PaymasterProvider` interface with Circle, Pimlico, and Alchemy implementations
- Modify `Manager.submitUserOp()` to support 2-phase paymaster flow (stub → final)
- Add `PaymasterDataFunc` callback injection to avoid import cycles
- Extend `SmartAccountConfig` with `SmartAccountPaymasterConfig`
- Add `allowedPaymasters` field to Solidity `SessionPolicy` struct for on-chain paymaster restriction
- Add `paymaster_status` and `paymaster_approve` agent tools
- Add `lango account paymaster status|approve` CLI commands
- Wire paymaster provider into app initialization via callback pattern

## Capabilities

### New Capabilities
- `paymaster`: ERC-4337 paymaster integration for gasless USDC transactions — provider interface, Circle/Pimlico/Alchemy implementations, 2-phase sponsorship flow, USDC approval helper, config, CLI, and agent tools

### Modified Capabilities
- `smart-account`: SessionPolicy gains `allowedPaymasters` field for on-chain paymaster allowlist enforcement

## Impact

- **Solidity**: `ISessionValidator.sol`, `LangoSessionValidator.sol` — new struct field and validation logic
- **Go packages**: New `internal/smartaccount/paymaster/` package (7 files), modified `manager.go`, `types.go`, `config/types_smartaccount.go`
- **App wiring**: `wiring_smartaccount.go` — paymaster provider initialization and callback injection
- **CLI/Tools**: `tools_smartaccount.go` — 2 new tools; `cli/smartaccount/paymaster.go` — 2 new commands
- **Dependencies**: No new external dependencies (uses existing `go-ethereum` and `net/http`)
