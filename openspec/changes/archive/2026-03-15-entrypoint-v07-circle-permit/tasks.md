## 1. EIP-2612 Permit Builder

- [x] 1.1 Create `internal/smartaccount/paymaster/permit/builder.go` with PermitSigner/EthCaller interfaces, DomainSeparator, PermitStructHash, TypedDataHash, Sign, GetPermitNonce
- [x] 1.2 Create `internal/smartaccount/paymaster/permit/builder_test.go` with tests for domain separator, struct hash, signing, nonce queries

## 2. Bundler v0.7 Field Format

- [x] 2.1 Add PaymasterVerificationGasLimit and PaymasterPostOpGasLimit to GasEstimate in `bundler/types.go`
- [x] 2.2 Update `bundler/client.go` userOpToMap() to split initCode → factory/factoryData and paymasterAndData → paymaster/gas limits/paymasterData
- [x] 2.3 Update EstimateGas() to parse optional paymaster gas fields from v0.7 bundler response

## 3. Circle Permit Provider

- [x] 3.1 Add Mode field to SmartAccountPaymasterConfig in `config/types_smartaccount.go`
- [x] 3.2 Add CirclePermitProvider to `paymaster/circle.go` with stub and real SponsorUserOp modes
- [x] 3.3 Add CirclePermitProvider tests to `paymaster/circle_test.go` (type, stub, real sponsor)

## 4. Wiring

- [x] 4.1 Extend initPaymasterProvider() in `app/wiring_smartaccount.go` with walletProvider/ethCallerClient interfaces and permit mode branching
- [x] 4.2 Update initPaymasterProvider call site to pass wallet and rpcClient

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./internal/smartaccount/...` all pass
- [x] 5.3 `go test ./internal/payment/...` no regression
