## 1. Payment Service — Receipt, Retry, Nonce (Bugs 1, 2, 3)

- [x] 1.1 Add nonceMu sync.Mutex, receiptTimeout, maxRetries to Service struct
- [x] 1.2 Add submitWithRetry() with exponential backoff (1s, 2s, 4s)
- [x] 1.3 Add waitForConfirmation() with receipt polling and 2min timeout
- [x] 1.4 Update Send() to lock nonceMu around build/sign/submit, confirm receipt, update status to confirmed/failed
- [x] 1.5 Add GasUsed and BlockNumber fields to PaymentReceipt type

## 2. Smart Account Factory — CREATE2 Fix (Bugs 4, 6)

- [x] 2.1 Fix safeFactoryABI — proxyCreationCode() has no inputs
- [x] 2.2 Add proxyCode cache fields and getProxyCreationCode() method
- [x] 2.3 Fix ComputeAddress() — initCodeHash = keccak256(proxyCode ++ singletonPadded), add ctx/error return
- [x] 2.4 Fix Deploy() — parse actual address from result, fallback to ComputeAddress
- [x] 2.5 Update callers in manager.go (GetOrDeploy, Info)
- [x] 2.6 Update factory_test.go and manager_test.go for new ComputeAddress signature

## 3. Smart Account Manager — UserOp Receipt (Bug 5)

- [x] 3.1 Add waitForUserOpReceipt() polling GetUserOperationReceipt with exponential backoff
- [x] 3.2 Update submitUserOp() to call waitForUserOpReceipt and return on-chain tx hash
- [x] 3.3 Update mock bundler in manager_test.go to handle eth_getUserOperationReceipt

## 4. EIP-3009 Double-Hashing Fix (Bug 8)

- [x] 4.1 Add SignTransaction to WalletSigner interface in eip3009/builder.go
- [x] 4.2 Change Sign() to use wallet.SignTransaction instead of wallet.SignMessage
- [x] 4.3 Update testWallet in builder_test.go to implement SignTransaction

## 5. Session Key Hash Fix (Bug 9)

- [x] 5.1 Export ComputeUserOpHash as package-level function in manager.go
- [x] 5.2 Add WithEntryPoint and WithChainID options to session Manager
- [x] 5.3 Replace hashUserOp() with sa.ComputeUserOpHash() in session/manager.go
- [x] 5.4 Wire entryPoint/chainID in app/wiring_smartaccount.go and cli/smartaccount/deps.go
- [x] 5.5 Update integration_test.go — use ComputeUserOpHash for verification, add options to test managers

## 6. X402 Key Zeroing + Contract Caller Backoff (Bugs 10, 11)

- [x] 6.1 Fix x402/signer.go — use hex.Encode into mutable []byte buffer
- [x] 6.2 Fix contract/caller.go — change linear backoff to exponential (1<<attempt seconds)

## 7. Payment Tool Status (Bug 7)

- [x] 7.1 Update payment_send tool to return receipt.Status instead of hardcoded "submitted"
- [x] 7.2 Include gasUsed and blockNumber in tool response when available

## 8. Gas Fee Fallback Warning (Bug 13)

- [x] 8.1 Add WARNING log to tx_builder.go when baseFee is nil
- [x] 8.2 Add WARNING log to contract/caller.go when baseFee is nil

## 9. Verification

- [x] 9.1 go build ./... passes
- [x] 9.2 go test ./internal/payment/... passes
- [x] 9.3 go test ./internal/smartaccount/... passes
- [x] 9.4 go test ./internal/x402/... ./internal/contract/... passes
