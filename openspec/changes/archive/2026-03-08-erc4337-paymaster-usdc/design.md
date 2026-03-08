## Context

Lango's Smart Account system (ERC-7579 + ERC-4337) submits all UserOperations with empty `paymasterAndData`, requiring users to hold ETH for gas. On Base chain, Circle operates a USDC paymaster that sponsors gas in exchange for USDC. The design must integrate paymaster support without breaking existing non-paymaster flows.

Current flow: `buildUserOp → estimateGas → sign → submit` (PaymasterAndData = empty)

## Goals / Non-Goals

**Goals:**
- Circle Paymaster as 1st-class provider; Pimlico/Alchemy behind same interface
- 2-phase paymaster interaction: stub data for gas estimation, final data after gas confirmation
- Graceful degradation: `paymasterFn == nil` → existing flow unchanged
- On-chain paymaster allowlist in SessionValidator for session key security
- Callback injection pattern to prevent import cycles

**Non-Goals:**
- Gas token selection UI (USDC-only for now)
- Multi-token paymaster support
- Custom paymaster contract deployment
- Gas price oracle integration

## Decisions

### 1. Two-Phase Paymaster Flow
**Decision**: Phase 1 (stub=true) gets temporary paymasterAndData for gas estimation, Phase 2 (stub=false) gets final signed data after gas values are confirmed.
**Rationale**: Paymasters need accurate gas values to sign sponsorship data. Single-phase would either under-estimate (no paymaster verification gas) or require re-estimation.
**Alternative**: Single-phase with fixed gas buffer — rejected because paymaster-specific verification gas varies significantly.

### 2. Callback Injection (`PaymasterDataFunc`)
**Decision**: `Manager.SetPaymasterFunc(fn)` callback instead of direct provider import.
**Rationale**: Follows existing `session.RegisterOnChainFunc` pattern. Prevents import cycle: `manager.go` → `paymaster/` → `bundler/` would create a cycle through shared types. The callback uses `smartaccount.UserOperation` directly.
**Alternative**: Interface injection — would work but callback is simpler for a single method and matches codebase convention.

### 3. Paymaster-Local Mirror Types
**Decision**: `paymaster.UserOpData` mirrors `smartaccount.UserOperation` fields.
**Rationale**: Same pattern as `bundler.UserOperation`. Prevents import cycle between `paymaster/` and `smartaccount/`.

### 4. On-Chain Allowlist (Solidity)
**Decision**: `allowedPaymasters` array in `SessionPolicy` struct, empty = all allowed.
**Rationale**: Session keys should restrict which paymasters can be used to prevent unauthorized gas sponsorship. Empty-array-means-all pattern matches existing `allowedTargets` behavior for backward compatibility.

### 5. Shared JSON-RPC Client Pattern
**Decision**: Each provider has its own `call()` helper following `bundler/client.go` pattern (`http.Client` + `atomic.Int64` reqID).
**Rationale**: Code duplication is minimal (each provider has different RPC methods/params). Shared base class would add abstraction without benefit for 3 simple providers.

## Risks / Trade-offs

- **[Paymaster downtime]** → Graceful degradation: if paymaster fails, error propagates clearly; user can disable paymaster and pay in ETH
- **[USDC approval frontrunning]** → Standard ERC-20 risk; recommend `approve(0)` before `approve(amount)` for security-sensitive users
- **[Gas override manipulation]** → Trust paymaster provider; overrides are optional and only apply to gas limits
- **[Struct storage growth]** → `allowedPaymasters` adds dynamic array to SessionPolicy; gas cost for registration increases with array size
