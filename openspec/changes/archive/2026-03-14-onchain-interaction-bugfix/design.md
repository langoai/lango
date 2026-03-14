## Context

The project has multiple on-chain interaction paths — payment service, smart account factory/manager, EIP-3009 signing, session key signing, and contract caller. A full audit revealed that the reference implementations (`economy/escrow/usdc_settler.go`, `p2p/settlement/service.go`) correctly implement nonce locking, exponential-backoff retry, receipt polling, and status checking, but several other paths lack one or more of these safeguards.

## Goals / Non-Goals

**Goals:**
- Every on-chain transaction path confirms receipt before reporting success
- Nonce collisions prevented via mutex serialization in each service
- Transaction submission retries with exponential backoff
- Correct CREATE2 address computation matching SafeProxyFactory behavior
- EIP-3009 and session key signatures valid for on-chain verification
- Sensitive key material properly zeroed after use

**Non-Goals:**
- Global nonce manager across all services (Bug 12 — architectural, deferred)
- Refactoring all services to share a common transaction submission layer
- Adding new on-chain features or transaction types

## Decisions

### D1: Per-service nonce mutex (not global nonce manager)
Each service that submits transactions holds its own `sync.Mutex` around the nonce-fetch-through-submit critical section. This matches the reference implementations and avoids cross-service coupling. A global nonce manager would require a shared dependency that doesn't exist today.

### D2: Receipt polling with exponential backoff
Adopted the `waitForConfirmation` pattern from `usdc_settler.go`: poll `TransactionReceipt`/`GetUserOperationReceipt` with 1s initial backoff, doubling up to 16s, with a 2-minute overall deadline. This balances responsiveness with RPC rate limiting.

### D3: SignTransaction for pre-hashed digests
`SignTransaction` signs raw bytes; `SignMessage` applies keccak256 first. EIP-712 typed data and ERC-4337 UserOp hashes are already keccak256 digests, so `SignTransaction` is the correct method. The `WalletSigner` interface is extended to include both methods.

### D4: Shared ComputeUserOpHash as exported package function
Extracted from `Manager.computeUserOpHash()` to `ComputeUserOpHash()` (exported, package-level). Session manager imports and uses it with entryPoint/chainID passed via options. This avoids duplicating the ERC-4337 v0.7 hash algorithm.

### D5: Factory proxyCreationCode via view call + caching
`ComputeAddress` now calls the factory's `proxyCreationCode()` view function and caches the result. This is more robust than hardcoding the creation code, which varies by Safe version. The cache uses a simple mutex-guarded field since the code is immutable per factory contract.

## Risks / Trade-offs

- **[Risk] Receipt polling adds latency** → Transactions now block until confirmed (up to 2 minutes). This is intentional — reporting unconfirmed transactions as successful is the bug we're fixing.
- **[Risk] proxyCreationCode() RPC call** → `ComputeAddress` now requires a context and can fail. Mitigated by caching after first successful call.
- **[Risk] WalletSigner interface change** → Adding `SignTransaction` is additive (not breaking) since all existing wallet implementations already have this method from the `WalletProvider` interface.
- **[Trade-off] Per-service mutex limits throughput** → Concurrent transactions from the same service are serialized. Acceptable for current usage patterns where parallel submission is rare.
