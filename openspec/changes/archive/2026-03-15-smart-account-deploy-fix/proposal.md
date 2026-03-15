## Why

Safe smart account deployment via `Factory.Deploy()` fails with "execution reverted" because the factory passes the Safe7579 adapter address as the singleton to `createProxyWithNonce()`. Safe7579 is an ERC-7579 adapter contract that does not have a `setup()` function -- the Safe L2 implementation contract does. This causes every deployment attempt to revert.

Additionally, when deployment does revert, the error messages are opaque: bundler JSON-RPC errors lack the revert data field, the contract caller does not extract revert reasons from go-ethereum errors, and tool execution failures are silently swallowed (returned as agent text but never logged server-side).

## What Changes

- **Bug 1 (Root cause)**: Separate `singletonAddr` (Safe L2 implementation, has `setup()`) from `safe7579Addr` (ERC-7579 adapter, delegate-called during setup). Add `SafeSingletonAddress` config field with default `0x29fcB43b46531BcA003ddC8FCB67FFE91900C762` (Safe L2 v1.4.1).
- **Bug 2 (Revert diagnostics)**: Add `Data` field to bundler `jsonrpcError`, implement `DecodeRevertReason()` for `Error(string)` and `Panic(uint256)`, add `extractRevertReason()` for go-ethereum `DataError`, add eth_call replay fallback for receipt failures.
- **Bug 3 (Tool logging)**: Add WARN-level log in `adk/tools.go` for all tool call failures. Subscribe to `ToolExecutedEvent` in observability wiring for failed tool event logging.

## Capabilities

### Modified Capabilities
- `smart-account`: Correct singleton/adapter separation in Factory, new `SafeSingletonAddress` config field, TUI field
- `contract-interaction`: Revert reason extraction from go-ethereum errors and eth_call replay fallback
- `tool-observability`: WARN log on tool call failures in ADK adapter, event bus logging for failed tool events

## Impact

- `internal/config/types_smartaccount.go` -- New `SafeSingletonAddress` field
- `internal/smartaccount/factory.go` -- Separated `singletonAddr` from `safe7579Addr`, updated `NewFactory` signature
- `internal/app/wiring_smartaccount.go` -- Singleton resolution with default Safe L2 v1.4.1
- `internal/cli/smartaccount/deps.go` -- Same singleton resolution for CLI
- `internal/cli/settings/forms_smartaccount.go` -- TUI field for Safe Singleton
- `internal/cli/tuicore/state_update.go` -- State update for `sa_singleton_address`
- `internal/smartaccount/bundler/types.go` -- `Data` field + `RevertReason()` method
- `internal/smartaccount/bundler/revert.go` -- New file: `DecodeRevertReason()` decoder
- `internal/smartaccount/bundler/client.go` -- Include revert reason in bundler error messages
- `internal/contract/caller.go` -- `extractRevertReason()` + `replayForRevertReason()` + eth_call fallback
- `internal/adk/tools.go` -- WARN log on tool call failure
- `internal/app/wiring_observability.go` -- Log failed tool events via event bus
- Tests: `revert_test.go` (new), `factory_test.go`, `manager_test.go` (updated `NewFactory` calls)
