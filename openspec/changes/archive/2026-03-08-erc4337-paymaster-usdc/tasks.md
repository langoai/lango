## 1. Solidity — SessionValidator Paymaster Allowlist

- [x] 1.1 Add `allowedPaymasters` field to `ISessionValidator.SessionPolicy` struct
- [x] 1.2 Add paymaster allowlist validation in `LangoSessionValidator.validateUserOp()`
- [x] 1.3 Update `_setSession()` to persist `allowedPaymasters`

## 2. Solidity — Foundry Tests

- [x] 2.1 Create `LangoSessionValidator_Paymaster.t.sol` with 6 paymaster allowlist tests
- [x] 2.2 Create `PaymasterIntegration.t.sol` with MockPaymaster and MockUSDC integration tests
- [x] 2.3 Update `LangoSessionValidator.t.sol` `_defaultPolicy()` for new struct field

## 3. Go — Paymaster Package

- [x] 3.1 Create `paymaster/types.go` — PaymasterProvider interface, SponsorRequest/Result, UserOpData
- [x] 3.2 Create `paymaster/errors.go` — sentinel errors
- [x] 3.3 Create `paymaster/circle.go` — CircleProvider with JSON-RPC client
- [x] 3.4 Create `paymaster/approve.go` — USDC approve calldata builder
- [x] 3.5 Create `paymaster/circle_test.go` — table-driven tests

## 4. Go — Pimlico + Alchemy Providers

- [x] 4.1 Create `paymaster/pimlico.go` — PimlicoProvider with policy ID support
- [x] 4.2 Create `paymaster/pimlico_test.go` — tests with policy ID verification
- [x] 4.3 Create `paymaster/alchemy.go` — AlchemyProvider with combined endpoint
- [x] 4.4 Create `paymaster/alchemy_test.go` — tests with gas override verification

## 5. Go — Config + Manager Integration

- [x] 5.1 Add `SmartAccountPaymasterConfig` to `config/types_smartaccount.go`
- [x] 5.2 Add `Paymaster` field to `SmartAccountConfig`
- [x] 5.3 Add `PaymasterGasOverrides` and `PaymasterDataFunc` to `smartaccount/types.go`
- [x] 5.4 Add `paymasterFn` field and `SetPaymasterFunc()` setter to Manager
- [x] 5.5 Update `submitUserOp()` with 2-phase paymaster flow

## 6. Go — App Wiring

- [x] 6.1 Add `paymasterProvider` field to `smartAccountComponents`
- [x] 6.2 Create `initPaymasterProvider()` factory function
- [x] 6.3 Wire paymaster callback in `initSmartAccount()` after manager creation

## 7. Go — Agent Tools

- [x] 7.1 Add `paymaster_status` tool (Safe)
- [x] 7.2 Add `paymaster_approve` tool (Dangerous)
- [x] 7.3 Register tools in `buildSmartAccountTools()`

## 8. Go — CLI Commands

- [x] 8.1 Create `cli/smartaccount/paymaster.go` with status and approve subcommands
- [x] 8.2 Register `paymasterCmd` in `NewAccountCmd()`

## 9. Go — Manager Integration Tests

- [x] 9.1 Add `TestSubmitUserOp_NoPaymaster` — verify existing flow unchanged
- [x] 9.2 Add `TestSubmitUserOp_PaymasterTwoPhase` — verify stub + final called
- [x] 9.3 Add `TestSubmitUserOp_PaymasterStubFails` — verify error propagation
- [x] 9.4 Add `TestSubmitUserOp_PaymasterFinalFails` — verify error propagation
- [x] 9.5 Add `TestSubmitUserOp_PaymasterGasOverrides` — verify override application
