## Why

Full audit of on-chain interactions revealed 13 bugs where transactions appear to succeed locally but fail or produce invalid results on-chain. Payment sends record "submitted" without confirming receipts, smart account deploys return wrong addresses due to incorrect CREATE2 computation, EIP-3009 signatures are invalid due to double-hashing, and session key signatures use a non-standard hash that doesn't match the EntryPoint contract.

## What Changes

- **Payment service**: Add nonce mutex, exponential-backoff retry, and receipt polling before reporting status. Return `confirmed`/`failed` instead of `submitted`.
- **Smart account factory**: Fix CREATE2 address computation by fetching actual `proxyCreationCode()` and using correct `initCodeHash = keccak256(proxyCode ++ singletonPadded)`.
- **Smart account manager**: Add UserOp receipt polling via `GetUserOperationReceipt()` (already implemented in bundler client but unused). Return actual on-chain tx hash.
- **EIP-3009 builder**: Fix double-hashing by using `SignTransaction` (raw sign) instead of `SignMessage` (which applies an extra keccak256).
- **Session key manager**: Replace broken `hashUserOp()` (raw byte concatenation) with `ComputeUserOpHash()` (proper ABI encoding matching EntryPoint).
- **X402 signer**: Fix key zeroing to use mutable `[]byte` buffer instead of immutable string conversion.
- **Contract caller**: Fix retry backoff from linear to exponential.
- **Payment tool**: Return receipt-based status, gasUsed, blockNumber instead of hardcoded "submitted".
- **Gas fee fallback**: Add warning log when baseFee is nil and fallback value is used.

## Capabilities

### New Capabilities

### Modified Capabilities
- `payment-service`: Add nonce locking, retry with exponential backoff, receipt-based confirmation
- `smart-account`: Fix CREATE2 address computation, UserOp receipt verification, session key hash algorithm
- `contract-interaction`: Fix retry backoff to exponential, add gas fee fallback warning
- `payment-tools`: Return on-chain confirmed status instead of hardcoded "submitted"

## Impact

- `internal/payment/service.go` — nonce mutex, retry, receipt polling
- `internal/payment/tx_builder.go` — gas fee fallback warning
- `internal/payment/types.go` — GasUsed, BlockNumber fields added to PaymentReceipt
- `internal/payment/eip3009/builder.go` — SignTransaction instead of SignMessage, WalletSigner interface extended
- `internal/smartaccount/factory.go` — ComputeAddress signature change (ctx, error), proxyCreationCode caching
- `internal/smartaccount/manager.go` — ComputeUserOpHash exported, waitForUserOpReceipt added
- `internal/smartaccount/session/manager.go` — Uses ComputeUserOpHash, added entryPoint/chainID fields
- `internal/x402/signer.go` — Key zeroing fix
- `internal/contract/caller.go` — Exponential backoff, gas fee warning
- `internal/tools/payment/payment.go` — Receipt-based status output
- `internal/app/wiring_smartaccount.go` — Session manager entryPoint/chainID wiring
- `internal/cli/smartaccount/deps.go` — Same wiring for CLI path
