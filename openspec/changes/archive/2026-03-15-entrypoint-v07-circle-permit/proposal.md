## Why

Lango's smart account system needs to support EntryPoint v0.7 bundler RPC format and Circle's on-chain permissionless paymaster. The current bundler client sends v0.6 field format (`initCode`, `paymasterAndData` as single fields), but v0.7 bundlers expect these split into separate fields (`factory`/`factoryData`, `paymaster`/`paymasterVerificationGasLimit`/`paymasterPostOpGasLimit`/`paymasterData`). Circle Paymaster uses EIP-2612 permit signatures to pay gas in USDC without requiring an API key — making onboarding frictionless on Base Sepolia.

## What Changes

- Add EIP-2612 permit signing builder (`internal/smartaccount/paymaster/permit/`) for USDC permit interactions
- Update bundler client `userOpToMap()` to emit v0.7 split fields (`factory`/`factoryData`, `paymaster`/gas limits/`paymasterData`)
- Add `PaymasterVerificationGasLimit` and `PaymasterPostOpGasLimit` to `GasEstimate` struct
- Add `CirclePermitProvider` to the paymaster package — builds `PaymasterAndData` with EIP-2612 permit signature
- Add `Mode` field to `SmartAccountPaymasterConfig` (`"rpc"` default | `"permit"`)
- Extend `initPaymasterProvider()` wiring to support `mode="permit"` + `provider="circle"` combination

## Capabilities

### New Capabilities
- `paymaster-permit`: EIP-2612 permit-based paymaster mode for Circle's on-chain paymaster contract, enabling USDC gas payment without API keys

### Modified Capabilities
- `paymaster`: Add `mode` config field and `CirclePermitProvider` alongside existing RPC-based providers
- `smart-account`: Bundler client now emits v0.7 split field format for `initCode` and `paymasterAndData`

## Impact

- **Code**: `internal/smartaccount/paymaster/permit/` (new), `internal/smartaccount/bundler/` (modified), `internal/smartaccount/paymaster/circle.go` (modified), `internal/config/types_smartaccount.go` (modified), `internal/app/wiring_smartaccount.go` (modified)
- **Config**: New `smartAccount.paymaster.mode` field (backward compatible, defaults to `"rpc"`)
- **Dependencies**: No new external dependencies — uses existing `go-ethereum` crypto primitives
- **Bundler compatibility**: v0.7 field format is now the default — existing bundler URLs must support v0.7
