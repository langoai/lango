## Why

Lango already landed the first escrow recommendation execution slice in code, but the operator docs and OpenSpec state still lag behind the implementation. Without truthful docs and synced specs, the repository overstates some gaps and understates the parts that now exist.

## What Changes

- Add a dedicated operator document for escrow execution.
- Truth-align surrounding security, architecture, and README surfaces with the landed first slice.
- Capture the first `create + fund` escrow execution capability in OpenSpec delta specs.
- Sync main specs and archive the completed change.

## Capabilities

### New Capabilities
- `escrow-execution`: Receipt-backed execution of approved escrow recommendations through the first `create + fund` path.

### Modified Capabilities
- `dispute-ready-receipts`: Receipt trails and transaction receipts now carry escrow execution evidence.
- `upfront-payment-approval`: Escrow-approved transactions bind escrow execution input onto transaction receipts.
- `security-docs-sync`: Security docs now include truthful operator docs for escrow execution.

## Impact

- Affected code: none in this task slice; implementation is already landed in `internal/app`, `internal/escrowexecution`, and `internal/receipts`
- Affected behavior: documented operator-visible behavior now matches the shipped first slice
- Affected docs: security docs, architecture docs, README, MkDocs nav, OpenSpec change/spec archive
