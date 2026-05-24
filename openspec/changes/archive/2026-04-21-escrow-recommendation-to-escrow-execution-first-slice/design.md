## Overview

This change is a closeout slice for already-landed implementation. The goal is not to alter runtime behavior. The goal is to make the operator docs and OpenSpec state match the code that now exists.

The landed implementation has three important boundaries:

1. `approve_upfront_payment` can recommend `escrow` and bind escrow execution input onto the transaction receipt.
2. `execute_escrow_recommendation` executes the approved escrow path using only `transaction_receipt_id`.
3. The runtime currently performs only `create + fund`, while preserving canonical receipt evidence.

## Source Of Truth

The truth source for this closeout is the currently landed implementation in:

- `internal/app/tools_meta.go`
- `internal/escrowexecution/service.go`
- `internal/receipts/types.go`
- `internal/receipts/store.go`

The docs and specs in this change must reflect those boundaries exactly:

- create and fund are landed
- activate, release, refund, and dispute are not landed
- there is no human approval UI
- transaction receipts are the canonical escrow execution state surface

## Documentation Shape

The new operator document should mirror the style of the existing security slice docs:

- what ships in this slice
- operator entry points
- execution model
- receipt evidence
- current limits
- related docs

Surrounding docs should be updated only where the previous text is no longer true.

## Spec Sync Approach

This change adds one new main capability spec and updates three existing specs:

- new `escrow-execution`
- updated `dispute-ready-receipts`
- updated `upfront-payment-approval`
- updated `security-docs-sync`

The synced main specs should describe currently implemented behavior only. They should not speculate about future escrow lifecycle work beyond clearly marked limits in docs.
