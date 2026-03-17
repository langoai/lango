## Context

The on-chain transaction stack spans 3 packages: `contract` (low-level tx submission), `smartaccount/bundler` (ERC-4337 bundler client), and `smartaccount` (manager/factory). Each package was implemented correctly in isolation, but integration gaps cause silent failures:

1. `caller.Write()` treats receipt timeout as success
2. UserOp gas prices are always zero (bundler rejects or ignores)
3. Bundler nonce uses EOA nonce instead of EntryPoint nonce
4. Deployment check uses a fragile `isModuleInstalled()` call instead of `eth_getCode`
5. No post-deploy verification
6. Double-hashing in UserOp signing

## Goals / Non-Goals

**Goals:**
- Every on-chain transaction failure surfaces as an error to the caller
- UserOps include correct gas fee parameters from the network
- EntryPoint nonce is used for UserOp ordering
- Deployment detection is reliable (eth_getCode)
- Post-deploy verification catches factory failures

**Non-Goals:**
- Changing the overall ERC-4337 architecture
- Adding gas estimation retry/fallback logic
- Modifying payment service's own Send() path (uses separate tx builder)

## Decisions

### 1. Receipt status check in caller.Write()

Return `ErrTxReverted` with receipt status on revert; return `ErrReceiptTimeout` on timeout instead of swallowing to `nil`. Reference pattern: `p2p/settlement/service.go:297-300`.

**Alternative**: Return the result with a status field — rejected because callers expect `error != nil` to mean failure.

### 2. EntryPoint nonce via eth_call

Use `eth_call` to `EntryPoint.getNonce(address, uint192 key=0)` (selector `0x35567e1a`) instead of `eth_getTransactionCount`. The ABI-encoded result is decoded via `hexutil.Decode` + `big.Int.SetBytes` (not `hexutil.DecodeBig` which rejects leading zeros in ABI encoding).

**Alternative**: Use bundler-specific `eth_getUserOperationCount` — rejected because not all bundlers support it.

### 3. Gas fee retrieval

New `GetGasFees()` method on bundler client: calls `eth_maxPriorityFeePerGas` (with fallback to 1.5 gwei default) and `eth_getBlockByNumber("latest")` for baseFee. Formula: `maxFeePerGas = 2 * baseFee + priorityFee`.

Gas fees are fetched before gas estimation so the bundler has realistic fee values during estimation.

### 4. Factory receives ethclient.Client

`NewFactory()` signature adds `*ethclient.Client` parameter. `IsDeployed()` uses `rpc.CodeAt()` — returns `len(code) > 0`. Errors are propagated (not swallowed as "not deployed").

### 5. SignTransaction instead of SignMessage

`computeUserOpHash()` returns a keccak256 digest. `SignMessage()` applies an additional `crypto.Keccak256()` internally, causing double-hashing. `SignTransaction()` signs the raw bytes directly.

## Risks / Trade-offs

- [Risk] `eth_maxPriorityFeePerGas` unsupported on some RPCs → Mitigation: fallback to default 1.5 gwei
- [Risk] `eth_getBlockByNumber` via bundler URL may not be supported → Mitigation: bundler URLs typically proxy to full nodes
- [Risk] Factory constructor signature change → Mitigation: only 2 call sites (app wiring, CLI deps), both updated
- [Risk] `caller.Write()` now errors on timeout instead of returning partial result → Mitigation: this was always a bug; callers should handle the error
