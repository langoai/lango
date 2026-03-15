## Context

Safe smart account deployment has been failing with "execution reverted" on every attempt. Investigation revealed three layered bugs: (1) the factory was using the wrong address as the proxy singleton, (2) revert reasons from the EVM were not surfaced in error messages, and (3) tool execution errors were silently swallowed.

## Goals / Non-Goals

**Goals:**
- Fix the root cause: use Safe L2 singleton for proxy creation, Safe7579 for delegate-call setup
- Surface EVM revert reasons in bundler and contract caller error messages
- Log tool execution failures server-side for operational visibility

**Non-Goals:**
- Changing the Safe account deployment flow (owner, threshold, module setup)
- Adding support for non-Safe L2 singletons
- Structured error types for revert reasons (string-based for now)

## Decisions

### 1. Separate singleton and Safe7579 addresses in Factory

**Decision**: Add a dedicated `singletonAddr` field to `Factory` struct, distinct from `safe7579Addr`. Update `NewFactory` to accept both as separate parameters.

**Rationale**: The Safe proxy factory's `createProxyWithNonce(_singleton, initializer, saltNonce)` creates a proxy that delegates to `_singleton`. The singleton must be the Safe L2 implementation (which has `setup()`). The Safe7579 adapter is only called via delegate call during setup (the `to` parameter in `setup()`). Passing Safe7579 as the singleton caused the proxy to delegate to a contract without `setup()`, reverting immediately.

**Alternative**: Hardcode the singleton address -- rejected because different chains may deploy Safe L2 at different addresses.

### 2. SafeSingletonAddress config field with default

**Decision**: Add `SafeSingletonAddress` to `SmartAccountConfig`. If empty, default to `0x29fcB43b46531BcA003ddC8FCB67FFE91900C762` (Safe L2 v1.4.1 on major chains).

**Rationale**: Safe L2 v1.4.1 is the canonical implementation on Ethereum, Polygon, Base, Arbitrum, Optimism. Defaulting reduces config burden while allowing override for custom deployments.

### 3. Revert reason decoding via DecodeRevertReason()

**Decision**: New `bundler.DecodeRevertReason(hexData)` function handles `Error(string)` (selector `08c379a0`), `Panic(uint256)` (selector `4e487b71`), and unknown selectors (returned as truncated hex).

**Rationale**: ERC-4337 bundlers return revert data in the JSON-RPC error `data` field. The existing `jsonrpcError` struct lacked this field entirely. The decoder is reused by `contract.Caller` for go-ethereum `DataError` extraction.

### 4. eth_call replay for receipt-level reverts

**Decision**: When a confirmed transaction has `receipt.Status != 1`, replay the same call as `eth_call` at the revert block to extract the revert reason.

**Rationale**: Transaction receipts do not include revert reasons. Replaying as `eth_call` triggers the same EVM execution and returns the revert data in the error. This is the standard technique used by Etherscan and Foundry.

### 5. WARN log in ADK tool adapter

**Decision**: Add `logging.Agent().Warnw("tool call failed", ...)` in `adaptToolWithOptions` when `err != nil`.

**Rationale**: Tool errors were only returned as text to the agent's response stream. Server operators had no visibility into tool failures without this log line. Also subscribe to `ToolExecutedEvent` on the event bus in observability wiring for centralized failure logging.

## File Changes

| File | Change |
|------|--------|
| `internal/config/types_smartaccount.go` | Add `SafeSingletonAddress` field |
| `internal/smartaccount/factory.go` | Separate `singletonAddr`/`safe7579Addr` in struct + `NewFactory` |
| `internal/app/wiring_smartaccount.go` | Resolve singleton with default |
| `internal/cli/smartaccount/deps.go` | Same singleton resolution |
| `internal/cli/settings/forms_smartaccount.go` | TUI field for singleton |
| `internal/cli/tuicore/state_update.go` | `sa_singleton_address` case |
| `internal/smartaccount/bundler/types.go` | `Data` field + `RevertReason()` |
| `internal/smartaccount/bundler/revert.go` | `DecodeRevertReason()` (new) |
| `internal/smartaccount/bundler/client.go` | Revert reason in error messages |
| `internal/contract/caller.go` | `extractRevertReason()` + `replayForRevertReason()` |
| `internal/adk/tools.go` | WARN log on failure |
| `internal/app/wiring_observability.go` | Event bus tool failure logging |

## Risks / Trade-offs

- [Risk] eth_call replay adds one extra RPC call on reverted transactions. Mitigation: Only triggered on confirmed reverts, not on the happy path.
- [Risk] `DecodeRevertReason` may fail on custom error selectors. Mitigation: Falls back to truncated hex string, never panics.
- [Risk] Default singleton address may differ on exotic L2s. Mitigation: Config field allows override; default covers all major chains.
