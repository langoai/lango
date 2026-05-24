## Why

The direct-payment slice now treats `transaction_receipt_id` as a first-class required input for both `payment_send` and `p2p_pay`, and the docs already describe that requirement. But the `p2p_pay` wrapper still deferred the missing-input case to deeper payment-gate validation, and the `p2p-payment` spec still described only `peer_did` and `amount` as required parameters.

## What Changes

- Make `p2p_pay` enforce `transaction_receipt_id` at the wrapper layer.
- Add a regression for the missing `transaction_receipt_id` wrapper error.
- Sync the `p2p-payment` and production-readiness specs, plus the affected docs, with the implemented required-input contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `p2p-payment`: `p2p_pay` now rejects missing `transaction_receipt_id` at the wrapper layer.
- `production-readiness`: wrapper-level request-guard coverage now includes `p2p_pay`.
- `downstream-docs-sync`: public payment and P2P docs describe the immediate missing-parameter failure.

## Impact

- Affected code: `internal/app/tools_p2p.go`
- Affected tests: `internal/app/tools_p2p_payment_test.go`
- Affected docs: `README.md`, `docs/features/p2p-network.md`
- Affected specs: `openspec/specs/p2p-payment/spec.md`, `openspec/specs/production-readiness/spec.md`
