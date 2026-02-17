## 1. Configuration

- [x] 1.1 Add PaymentConfig, PaymentNetworkConfig, SpendingLimitsConfig, X402Config structs to `internal/config/types.go`
- [x] 1.2 Add Payment field to Config struct after A2A
- [x] 1.3 Add default values in DefaultConfig() in `internal/config/loader.go`
- [x] 1.4 Add viper defaults for payment settings
- [x] 1.5 Add payment validation in Validate() (rpcUrl required, walletProvider validation)
- [x] 1.6 Add payment.network.rpcUrl to substituteEnvVars

## 2. Dependencies and Schema

- [x] 2.1 Add `github.com/ethereum/go-ethereum` dependency
- [x] 2.2 Create `internal/ent/schema/payment_tx.go` with PaymentTx entity
- [x] 2.3 Run `go generate ./internal/ent/...` to generate Ent code

## 3. Wallet Package

- [x] 3.1 Create `internal/wallet/wallet.go` with WalletProvider interface, WalletInfo, NetworkName, ConnectionChecker
- [x] 3.2 Create `internal/wallet/local_wallet.go` with LocalWallet (SecretsStore key loading, secp256k1 signing, key zeroing)
- [x] 3.3 Create `internal/wallet/spending.go` with SpendingLimiter interface and EntSpendingLimiter (ParseUSDC, FormatUSDC)
- [x] 3.4 Create `internal/wallet/rpc_wallet.go` with RPCWallet (companion delegation, correlation IDs, 30s timeout)
- [x] 3.5 Create `internal/wallet/composite_wallet.go` with CompositeWallet (primary/fallback)
- [x] 3.6 Create `internal/wallet/spending_test.go` with tests for ParseUSDC, FormatUSDC, NetworkName

## 4. Payment Package

- [x] 4.1 Create `internal/payment/types.go` with PaymentRequest, PaymentReceipt, TransactionInfo, X402Challenge
- [x] 4.2 Create `internal/payment/tx_builder.go` with TxBuilder (ERC-20 ABI encoding, EIP-1559, gas estimation)
- [x] 4.3 Create `internal/payment/service.go` with Service (Send flow, Balance, History, HandleX402)
- [x] 4.4 Create `internal/payment/tx_builder_test.go` with tests for ValidateAddress and ERC20TransferMethodID

## 5. Payment Tools

- [x] 5.1 Create `internal/tools/payment/payment.go` with 5 agent tools (send, balance, history, limits, wallet_info)

## 6. X402 Protocol

- [x] 6.1 Create `internal/x402/types.go` with Challenge, PaymentPayload, header constants
- [x] 6.2 Create `internal/x402/handler.go` with ParseChallenge and BuildPaymentHeader
- [x] 6.3 Create `internal/x402/interceptor.go` with Interceptor (auto-pay with limit check)
- [x] 6.4 Create `internal/x402/handler_test.go` with tests for ParseChallenge and BuildPaymentHeader

## 7. App Wiring

- [x] 7.1 Add WalletProvider and PaymentService fields to `internal/app/types.go`
- [x] 7.2 Add paymentComponents struct and initPayment() to `internal/app/wiring.go`
- [x] 7.3 Add buildPaymentTools() to `internal/app/tools.go`
- [x] 7.4 Add payment_send case to buildApprovalSummary() in `internal/app/tools.go`
- [x] 7.5 Wire initPayment() and payment tools in `internal/app/app.go` (step 5h)

## 8. Orchestration

- [x] 8.1 Add `"payment_"` to executorPrefixes in `internal/orchestration/tools.go`
- [x] 8.2 Add `"payment_": "blockchain payments (USDC on Base)"` to capabilityMap

## 9. Verification

- [x] 9.1 Run `go build ./...` — compilation succeeds
- [x] 9.2 Run `go test ./internal/wallet/...` — all tests pass
- [x] 9.3 Run `go test ./internal/payment/...` — all tests pass
- [x] 9.4 Run `go test ./internal/x402/...` — all tests pass
- [x] 9.5 Run `go test ./internal/config/...` — all tests pass
- [x] 9.6 Run `go test ./internal/orchestration/...` — all tests pass
