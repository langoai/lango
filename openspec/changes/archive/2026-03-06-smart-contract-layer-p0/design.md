## Context

The P2P economy layer's escrow engine has a complete state machine (pending→funded→active→completed→released/refunded) but uses `noopSettler{}` for all fund operations. The payment system (`internal/payment/`) already provides `TxBuilder` for ERC-20 transfers and `wallet.WalletProvider` for signing. The settlement service (`internal/p2p/settlement/`) demonstrates the tx lifecycle pattern (nonce management, retry, receipt polling). Smart contract interaction is limited to hardcoded USDC `transfer()`/`balanceOf()`.

## Goals / Non-Goals

**Goals:**
- Replace `noopSettler` with `USDCSettler` that performs real on-chain USDC transfers via agent wallet as custodian
- Provide generic smart contract read/write capability for arbitrary contracts via ABI-based encoding
- Maintain backward compatibility (graceful fallback to `noopSettler` when payment is disabled)
- Expose contract interaction as agent tools and CLI commands

**Non-Goals:**
- On-chain escrow smart contract (uses agent wallet custodian model instead)
- Multi-chain support in a single session (one RPC client per instance)
- ABI auto-discovery (Etherscan/Sourcify integration deferred)
- Gas estimation optimization or gas sponsorship

## Decisions

### D1: Agent wallet as custodian vs. on-chain escrow contract
**Decision**: Use agent wallet as temporary custodian. Lock = balance check, Release/Refund = USDC transfer from agent wallet.
**Rationale**: No custom smart contract deployment needed. `SettlementExecutor` interface allows future swap to on-chain escrow. Matches the existing `settlement.Service` tx lifecycle patterns.
**Alternative**: Deploy escrow smart contract on Base — rejected for P0 due to deployment complexity and audit requirements.

### D2: DID-to-Address resolution via crypto.DecompressPubkey
**Decision**: Parse `did:lango:<hex>` suffix as compressed secp256k1 pubkey, decompress, derive Ethereum address.
**Rationale**: Deterministic, no external lookup. Reuses the identity package's DID format exactly.
**Alternative**: Maintain a DID↔address registry — rejected as it adds state and trust assumptions.

### D3: Generic contract caller with ABI cache
**Decision**: Thread-safe `ABICache` keyed by `chainID:address`, `Caller` struct with `Read()` and `Write()` methods using `go-ethereum/accounts/abi` for pack/unpack.
**Rationale**: Reuses existing gas fee constants from `payment.TxBuilder`. ABI caching avoids repeated JSON parsing. Same nonce-mutex + retry pattern as settlement service.
**Alternative**: Use `abigen` for type-safe bindings — rejected as it requires compile-time code generation per contract.

### D4: Functional options for USDCSettler
**Decision**: `WithReceiptTimeout`, `WithMaxRetries`, `WithLogger` options.
**Rationale**: Matches project conventions (Go functional options pattern). Allows config-driven customization without breaking constructor signature.

## Risks / Trade-offs

- [Custodian model trust] Agent wallet holds funds between Lock and Release → Mitigated by: `SettlementExecutor` interface allows future upgrade to on-chain escrow
- [Nonce collision under concurrent escrow ops] Multiple escrow releases at same time → Mitigated by: `nonceMu sync.Mutex` serializes all tx building
- [ABI cache unbounded growth] No eviction policy → Mitigated by: Minimal memory per ABI entry; acceptable for agent workloads. Add LRU if needed later
- [CLI commands are validation-only] `lango contract read/call` validate ABI but require `lango serve` for live execution → Acceptable for P0; full CLI execution requires bootstrap RPC setup
