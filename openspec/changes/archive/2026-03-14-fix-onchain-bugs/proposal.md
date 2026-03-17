## Why

Smart account deployment (`smart_account_deploy`) and payment send (`payment_send`) tools report success but produce no on-chain state changes. `cast code` shows no deployed bytecode (0x), and USDC balances remain unchanged. Root cause: 7 bugs across the on-chain interaction stack where errors are silently swallowed or wrong parameters are passed, causing transactions to fail while being reported as successful.

## What Changes

- Add receipt status validation and timeout error propagation in `contract/caller.go`
- Fix bundler nonce retrieval to use EntryPoint.getNonce() instead of EOA nonce
- Add gas fee parameter fetching (maxFeePerGas/maxPriorityFeePerGas) so UserOps have non-zero gas prices
- Replace `isModuleInstalled()` with `eth_getCode` for deployment detection
- Add post-deploy on-chain verification in account manager
- Remove dead code path in factory Deploy()
- Fix UserOp signing to use SignTransaction (raw sign) instead of SignMessage (double-hashing)

## Capabilities

### New Capabilities

_(none — all changes are bug fixes to existing capabilities)_

### Modified Capabilities

- `contract-interaction`: Receipt status must be checked; timeout must propagate as error instead of silent success
- `smart-account`: EntryPoint nonce, gas fee parameters, deployment verification, correct signing method

## Impact

- `internal/contract/caller.go` — New sentinel errors `ErrTxReverted`, `ErrReceiptTimeout`; Write() now returns errors on receipt failure
- `internal/smartaccount/bundler/client.go` — GetNonce uses eth_call to EntryPoint; new GetGasFees() method
- `internal/smartaccount/bundler/types.go` — New GasFees struct
- `internal/smartaccount/manager.go` — Gas fee wiring, post-deploy verification, SignTransaction instead of SignMessage
- `internal/smartaccount/factory.go` — ethclient.Client field added to Factory; IsDeployed uses CodeAt; dead code removed
- `internal/app/wiring_smartaccount.go` — Pass rpcClient to NewFactory
- `internal/cli/smartaccount/deps.go` — Pass rpcClient to NewFactory
- All downstream callers of `caller.Write()` that relied on silent error swallowing will now receive proper errors
