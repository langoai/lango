## Why

`payment.Service.Send` records a pending transaction before building, signing, submitting, and confirming the on-chain payment. When a later step fails, `failTx` attempts to mark the record failed, but it currently discards `TxStore.UpdateStatus` errors. If that status update fails, callers see only the original payment failure while the ledger can remain `pending`, hiding the persistence failure from operators.

## What Changes

- Make failed-status persistence errors observable to `Send` callers.
- Preserve the original payment failure while also surfacing the failed-status update error.
- Add focused tests for `failTx` when the store cannot persist the failed state.

## Impact

- Payment failures fail more loudly when the ledger cannot be updated.
- Existing successful payment behavior is unchanged.
- No schema or configuration changes.
