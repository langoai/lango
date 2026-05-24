## Why

The codebase has funded escrow release and settlement execution slices, but no
receipt-backed escrow refund service for the first refund path. We need a small
service that can execute refunds from canonical transaction state without
touching the broader receipt or app wiring yet.

## What Changes

- Add a new `internal/escrowrefund` package for the first refund slice.
- Accept only `transaction_receipt_id` as input.
- Deny execution when the receipt is missing, has no current submission, is not
  funded, is not in `review-needed`, or cannot resolve the refund amount.
- Resolve the refund amount from canonical transaction context.
- Return canonical refund execution results with runtime reference and the
  current `review-needed` progression state preserved in the result.

## Capabilities

### New Capabilities
- `escrow-refund`: Receipt-backed escrow refund execution for the first refund
  slice.

### Modified Capabilities
- None.

## Impact

- Affected code: `internal/escrowrefund`.
- No receipt store, app wiring, CLI, or public docs changes in this slice.
