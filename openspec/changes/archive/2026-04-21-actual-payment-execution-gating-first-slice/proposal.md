## Why

Lango now has exportability, artifact release approval, dispute-ready receipts, and upfront payment approval decisioning, but direct payment execution still lacks a final enforcement layer. Without an execution gate, the system can decide that a payment should or should not be allowed without actually enforcing that decision at `payment_send` and `p2p_pay`.

## What Changes

- Introduce a first direct-payment execution gate for `payment_send` and `p2p_pay`.
- Make direct payment execution receipt-backed through transaction canonical payment approval state.
- Record both `allow` and `deny` execution outcomes into audit and receipt trails.
- Add truthful operator docs describing the first gate slice and its limits.

## Capabilities

### New Capabilities
- `payment-execution-gating`: Receipt-backed allow/deny gate for direct payment execution.

### Modified Capabilities
- `dispute-ready-receipts`: Receipt trails gain direct payment execution authorization/denial events.
- `meta-tools`: No new meta tool is required, but tool-adjacent execution surfaces now depend on receipt-backed canonical payment approval state.
- `security-docs-sync`: Add truthful operator docs for the first actual payment execution gate slice.

## Impact

- Affected code: `internal/paymentgate/*`, `internal/tools/payment/*`, `internal/app/tools_p2p.go`, `internal/receipts/*`
- Affected behavior: direct payment execution now enforces receipt-backed canonical approval state
- Affected docs: security docs, architecture docs, README
