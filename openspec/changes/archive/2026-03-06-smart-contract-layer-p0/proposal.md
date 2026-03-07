## Why

The P2P economy layer has two critical gaps: escrow settlement is a no-op (`noopSettler{}`) so no USDC actually moves on-chain, and there is no way to call arbitrary smart contracts beyond hardcoded `transfer()`/`balanceOf()`. Solving both enables agents to manage funds on Base and interact with any dApp.

## What Changes

- Implement `USDCSettler` that performs real on-chain USDC transfers for escrow Lock/Release/Refund using the agent wallet as custodian
- Add `DID-to-Address` resolver to convert `did:lango:<compressed-pubkey>` to Ethereum addresses
- Wire `USDCSettler` into the escrow engine when payment is enabled (graceful fallback to `noopSettler`)
- Create a generic smart contract interaction layer (`internal/contract/`) with ABI caching, read (view/pure), and write (state-changing tx) capabilities
- Register 3 agent tools: `contract_read` (Safe), `contract_call` (Dangerous), `contract_abi_load` (Safe)
- Add `lango contract read|call|abi load` CLI commands

## Capabilities

### New Capabilities
- `contract-interaction`: Generic smart contract caller with ABI cache, read/write methods, and agent tools
- `escrow-settlement`: On-chain USDC settlement executor for the escrow engine using agent wallet as custodian

### Modified Capabilities
- `economy-escrow`: Escrow engine now accepts real settlement executor when payment is enabled
- `economy-wiring`: `initEconomy` accepts `paymentComponents` parameter for settler wiring

## Impact

- New package: `internal/contract/` (types, abi_cache, caller)
- New files: `internal/economy/escrow/address_resolver.go`, `usdc_settler.go`
- Modified: `internal/app/wiring_economy.go` (new parameter), `internal/app/app.go` (pass pc)
- New wiring: `internal/app/wiring_contract.go`, `tools_contract.go`
- New CLI: `internal/cli/contract/` (group, read, call, abi)
- Modified: `cmd/lango/main.go` (add contract CLI), `internal/app/tools.go` (blockLangoExec guard)
- Config: `EscrowSettlementConfig` added to `EscrowConfig`
- Dependencies: uses existing `go-ethereum v1.16.8`, `payment.TxBuilder`, `wallet.WalletProvider`
