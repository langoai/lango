## 1. DID-to-Address Resolver

- [x] 1.1 Create `internal/economy/escrow/address_resolver.go` with `ResolveAddress()` and `ErrInvalidDID`
- [x] 1.2 Create `internal/economy/escrow/address_resolver_test.go` with table-driven tests (valid, missing prefix, invalid hex, invalid pubkey)

## 2. USDC Settler

- [x] 2.1 Create `internal/economy/escrow/usdc_settler.go` implementing `SettlementExecutor` (Lock/Release/Refund)
- [x] 2.2 Implement `transferFromAgent` with nonce mutex, tx build, sign, retry, receipt polling
- [x] 2.3 Add functional options: `WithReceiptTimeout`, `WithMaxRetries`, `WithLogger`
- [x] 2.4 Create `internal/economy/escrow/usdc_settler_test.go` with interface check and option tests

## 3. Escrow Wiring Update

- [x] 3.1 Add `EscrowSettlementConfig` to `internal/config/types_economy.go`
- [x] 3.2 Update `initEconomy` in `wiring_economy.go` to accept `*paymentComponents` and create `USDCSettler` when available
- [x] 3.3 Update `app.go` to pass `pc` to `initEconomy`

## 4. ABI Cache & Contract Types

- [x] 4.1 Create `internal/contract/types.go` with `ContractCallRequest`, `ContractCallResult`, `ParseABI`
- [x] 4.2 Create `internal/contract/abi_cache.go` with thread-safe `ABICache` (Get/Set/GetOrParse)
- [x] 4.3 Create `internal/contract/abi_cache_test.go` with cache tests and concurrent access test

## 5. Generic Contract Caller

- [x] 5.1 Create `internal/contract/caller.go` with `Read()`, `Write()`, `LoadABI()` methods
- [x] 5.2 Implement EIP-1559 tx building, signing, retry, and receipt polling in `Write()`
- [x] 5.3 Create `internal/contract/caller_test.go` with constructor and LoadABI tests

## 6. Contract Agent Tools & Wiring

- [x] 6.1 Create `internal/app/tools_contract.go` with `contract_read`, `contract_call`, `contract_abi_load` tools
- [x] 6.2 Create `internal/app/wiring_contract.go` with `initContract()`
- [x] 6.3 Wire contract tools in `app.go` at step 5p (after economy)
- [x] 6.4 Add `"lango contract"` guard to `blockLangoExec` in `tools.go`

## 7. Contract CLI Commands

- [x] 7.1 Create `internal/cli/contract/group.go` with `NewContractCmd`
- [x] 7.2 Create `internal/cli/contract/read.go` with `lango contract read` command
- [x] 7.3 Create `internal/cli/contract/call.go` with `lango contract call` command
- [x] 7.4 Create `internal/cli/contract/abi.go` with `lango contract abi load` command
- [x] 7.5 Wire contract CLI in `cmd/lango/main.go` with GroupID "infra"

## 8. Verification

- [x] 8.1 `go build ./...` passes
- [x] 8.2 `go test ./internal/economy/escrow/...` passes
- [x] 8.3 `go test ./internal/contract/...` passes
