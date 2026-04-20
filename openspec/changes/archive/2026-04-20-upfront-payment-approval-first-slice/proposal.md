## Why

Lango now has first-slice exportability, artifact release approval, and dispute-ready receipt foundations, but it still lacks an explicit control plane for deciding whether an upfront payment may open a transaction. Without that, prepayment decisions remain implicit and cannot be reused by later payment execution or receipt flows.

## What Changes

- Introduce a first-slice upfront payment approval domain model.
- Add structured `approve / reject / escalate` decisioning with suggested payment modes and amount/risk classes.
- Add an approval receipt subtype that updates transaction-level canonical payment approval state.
- Add a minimal operator-facing doc that explains what this slice does and does not yet do.

## Capabilities

### New Capabilities
- `upfront-payment-approval`: Structured policy decisioning for whether an upfront payment path may open a `knowledge exchange v1` transaction.

### Modified Capabilities
- `dispute-ready-receipts`: Transaction receipts gain canonical payment approval status updates and payment approval event linkage.
- `meta-tools`: Add an `approve_upfront_payment` tool that evaluates a prepayment request and records the result.
- `security-docs-sync`: Add truthful operator docs and surrounding links for the first upfront payment approval slice.

## Impact

- Affected code: `internal/paymentapproval/*`, `internal/receipts/*`, `internal/app/tools_meta.go`
- Affected storage: receipt canonical state and payment approval event trail
- Affected docs: security docs, architecture docs, README
