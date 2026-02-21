## Why

The `x402_fetch` tool is not routed to any sub-agent because `PartitionTools()` uses `strings.HasPrefix` matching and `x402_fetch` does not match any vault prefix (`crypto_`, `secrets_`, `payment_`). It falls into the Unmatched bucket and is never assigned to a sub-agent. This is the only tool out of 71 with this routing bug.

## What Changes

- Rename `x402_fetch` → `payment_x402_fetch` so it matches the `payment_` prefix and routes to the vault agent
- Update the approval description switch case in `internal/app/tools.go`
- Update all documentation references (README, docs, openspec specs)

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `x402-v2`: Tool name changes from `x402_fetch` to `payment_x402_fetch` to fix prefix-based routing
- `payment-tools`: The x402 fetch tool now correctly uses the `payment_` prefix convention

## Impact

- `internal/tools/payment/payment.go` — tool name definition
- `internal/app/tools.go` — approval description switch case
- `README.md` — tool table and protocol flow
- `docs/payments/x402.md` — X402 protocol documentation
- `docs/payments/usdc.md` — USDC payment tool table
- `openspec/specs/x402-v2/spec.md` — X402 V2 specification
