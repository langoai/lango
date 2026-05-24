## Why

Lango now has first-slice exportability and approval-flow outputs, but it still lacks a dedicated receipt model that can hold canonical submission/transaction state while preserving the event trail needed for later disputes. Without that, evidence remains scattered across audit, provenance, and settlement-adjacent surfaces.

## What Changes

- Introduce a first `dispute-ready receipt lite` domain with separate submission and transaction receipts.
- Add canonical current-state tracking plus append-only event trail.
- Add a minimal receipt-creation surface for artifact submissions.
- Add truthful operator docs describing the first receipt slice and its limits.

## Capabilities

### New Capabilities
- `dispute-ready-receipts`: Dedicated submission and transaction receipts with canonical state, event trail, and external evidence references.

### Modified Capabilities
- `meta-tools`: Add a narrow receipt-creation tool for artifact submissions.
- `security-docs-sync`: Add truthful operator docs and surrounding documentation links for dispute-ready receipt lite.

## Impact

- Affected code: `internal/receipts/*`, `internal/app/tools_meta.go`
- Affected storage: new receipt models and their event trail
- Affected docs: security docs, architecture docs, README
