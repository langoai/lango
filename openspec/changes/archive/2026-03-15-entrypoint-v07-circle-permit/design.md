## Context

Lango's smart account system currently sends bundler RPC requests using ERC-4337 v0.6 field format. The v0.7 EntryPoint (`0x0000000071727De22E5E9d8BAf0edAc6f37da032`) expects split fields: `initCode` → `factory`+`factoryData`, `paymasterAndData` → `paymaster`+gas limits+`paymasterData`. Circle's on-chain paymaster (`0x31BE08D380A21fc740883c0BC434FcFc88740b58`) on Base Sepolia accepts EIP-2612 permit signatures for USDC gas payment without API keys.

Key discovery: `ComputeUserOpHash()` already uses v0.7 PackedUserOperation hash format, so the migration scope is limited to bundler RPC serialization and paymaster data assembly.

## Goals / Non-Goals

**Goals:**
- Support EntryPoint v0.7 bundler RPC field format
- Add Circle on-chain permit paymaster (no API key required)
- Maintain backward compatibility — existing RPC-mode paymaster providers continue working
- Add `mode` config field for paymaster mode selection

**Non-Goals:**
- Dual v0.6/v0.7 support (v0.7 only going forward)
- Circle Paymaster RPC mode changes (existing `CircleProvider` stays as-is)
- Changes to `ComputeUserOpHash` or `UserOperation` struct (already v0.7 compatible)

## Decisions

### 1. v0.7 field splitting in `userOpToMap()` only

**Decision**: Split `initCode` and `paymasterAndData` at the RPC serialization layer, not in the `UserOperation` struct.

**Rationale**: The struct uses `[]byte` fields that naturally carry packed v0.7 data. Splitting only at RPC boundary keeps the internal model clean and avoids breaking session signing, hash computation, and all existing consumers.

**Alternative**: Add separate `Factory`/`FactoryData` fields to `UserOperation` — rejected because it would require updating every constructor and consumer.

### 2. EIP-2612 permit builder as a separate package

**Decision**: `internal/smartaccount/paymaster/permit/` package with `PermitSigner` and `EthCaller` interfaces.

**Rationale**: Follows the existing `internal/payment/eip3009/` pattern. Interface-based design avoids import cycles with `wallet` package. The permit builder is reusable for future permit-based interactions.

### 3. CirclePermitProvider assembles PaymasterAndData locally

**Decision**: No RPC call to Circle — the provider builds `PaymasterAndData` entirely client-side using permit signing.

**Rationale**: Circle's on-chain paymaster is permissionless. The contract verifies the permit signature on-chain, so no off-chain coordination is needed. This eliminates a network dependency and simplifies the flow.

### 4. Stub mode returns correct-length zero data

**Decision**: For gas estimation (stub=true), return `PaymasterAndData` of the correct final length (170 bytes) with zero-filled signature.

**Rationale**: Bundlers use PaymasterAndData length to estimate verification gas. Wrong length causes inaccurate estimates or rejection.

## Risks / Trade-offs

- **[v0.6 bundler incompatibility]** → v0.7 field format is now the default. Users with v0.6 bundlers will get RPC errors. Mitigation: Document that v0.7 EntryPoint is required.
- **[Fixed permit amount (10 USDC)]** → The permit amount is hardcoded at 10 USDC per UserOp. Mitigation: This is generous for gas costs on Base Sepolia; can be made configurable later if needed.
- **[Permit nonce race]** → If multiple UserOps are submitted concurrently, permit nonces could collide. Mitigation: Sequential UserOp submission is the current pattern; concurrent support would need nonce management.
