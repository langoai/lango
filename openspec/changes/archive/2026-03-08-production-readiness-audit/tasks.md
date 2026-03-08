## 1. Stub Fixes & Dead Code Removal

- [x] 1.1 Replace enclave provider crash with actionable error listing valid providers in `internal/app/wiring.go`
- [x] 1.2 Add table-driven test for unsupported provider names in `internal/app/wiring_test.go`
- [x] 1.3 Implement Telegram `DownloadFile` with HTTP GET + 30s timeout in `internal/channels/telegram/telegram.go`
- [x] 1.4 Create `telegram_download_test.go` with httptest mock (success, HTTP error, empty body)
- [x] 1.5 Remove dead `NewX402Client` function and its `context.TODO()` from `internal/x402/handler.go`
- [x] 1.6 Improve GVisor stub doc comments in `internal/sandbox/gvisor_runtime.go`
- [x] 1.7 Create `gvisor_runtime_test.go` verifying IsAvailable=false, Run=ErrRuntimeUnavailable, Name="gvisor"

## 2. Wallet Package Tests

- [x] 2.1 Create `internal/wallet/wallet_test.go` — NetworkName, ChainID constants, zeroBytes
- [x] 2.2 Create `internal/wallet/local_wallet_test.go` — Address derivation, SignTransaction, SignMessage
- [x] 2.3 Create `internal/wallet/composite_wallet_test.go` — Fallback logic, UsedLocal sticky
- [x] 2.4 Create `internal/wallet/create_test.go` — CreateWallet, ErrWalletExists
- [x] 2.5 Create `internal/wallet/rpc_wallet_test.go` — RPC dispatching, timeout, context cancellation

## 3. Security Package Tests

- [x] 3.1 Create `internal/security/key_registry_test.go` — Full CRUD, GetDefaultKey, KeyType.Valid
- [x] 3.2 Create `internal/security/secrets_store_test.go` — Store/Get/List/Delete, encryption failure, access count

## 4. Payment Package Tests

- [x] 4.1 Create `internal/payment/service_test.go` — Send error branches, History, RecordX402Payment, failTx

## 5. Smart Account Package Tests

- [x] 5.1 Create `internal/smartaccount/factory_test.go` — CREATE2 determinism, buildSafeInitializer, Deploy
- [x] 5.2 Create `internal/smartaccount/session/crypto_test.go` — Key generate/serialize/deserialize roundtrip
- [x] 5.3 Create `internal/smartaccount/errors_test.go` — PolicyViolationError unwrap, sentinel errors
- [x] 5.4 Create `internal/smartaccount/module/abi_encoder_test.go` — ABI encoding byte-level verification
- [x] 5.5 Create `internal/smartaccount/paymaster/approve_test.go` — Approve calldata selector
- [x] 5.6 Create `internal/smartaccount/paymaster/errors_test.go` — IsTransient/IsPermanent classification
- [x] 5.7 Create `internal/smartaccount/policy/syncer_test.go` — PushToChain, PullFromChain, DetectDrift
- [x] 5.8 Create `internal/smartaccount/types_test.go` — ModuleType.String, SessionKey.IsMaster/IsExpired/IsActive
- [x] 5.9 Create `internal/smartaccount/policy/types_test.go` — SpendTracker.ResetIfNeeded

## 6. Economy & P2P Package Tests

- [x] 6.1 Create `internal/economy/risk/factors_test.go` — trustFactor, amountFactor, verifiabilityFactor, classifyRisk
- [x] 6.2 Create `internal/economy/risk/strategy_test.go` — 9-combination matrix, boundary values
- [x] 6.3 Create `internal/p2p/team/conflict_test.go` — All 4 strategies, empty results, unknown fallback
- [x] 6.4 Create `internal/p2p/protocol/messages_test.go` — ResponseStatus.Valid, RequestType constants, JSON roundtrip
- [x] 6.5 Create `internal/p2p/protocol/remote_agent_test.go` — NewRemoteAgent, accessor methods

## 7. Verification

- [x] 7.1 `go build ./...` passes with no errors
- [x] 7.2 `go test ./...` passes with no failures
- [x] 7.3 `go vet ./...` passes with no issues
