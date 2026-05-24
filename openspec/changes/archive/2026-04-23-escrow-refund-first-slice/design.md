## Context

The refund slice is intentionally smaller than the full escrow lifecycle. It
only needs receipt-backed lookup, amount resolution, and a runtime call. The
first slice does not change receipt persistence or app wiring yet.

## Goals / Non-Goals

**Goals**

- implement a focused refund service in `internal/escrowrefund`
- keep canonical input limited to `transaction_receipt_id`
- preserve `review-needed` status in both success and failure results
- support deterministic denial reasons for the first slice

**Non-Goals**

- no receipt mutation or persistence changes
- no CLI, app, or docs wiring
- no dispute orchestration

## Decisions

### 1. Use receipt-backed validation only

The service reads the transaction and current submission from the receipt store
before any runtime call. This keeps the first slice aligned with the existing
receipt-backed services.

### 2. Keep progression state unchanged in the result

The service returns `review-needed` in the result for both success and runtime
failure. That keeps this slice read-only from the receipt state perspective
until the later wiring work lands.

### 3. Resolve amount from canonical transaction context

The refund amount is derived from the transaction's canonical context rather
than from request input. This keeps the slice deterministic and avoids any
second source of truth.
