## Why

The escrow execution implementation treats an empty `transaction_receipt_id` as an actionable validation error before any receipt-backed business rejection logic runs. The public escrow execution doc and capability spec describe the happy path and unsupported receipt states, but they do not make that request-validation boundary explicit.

## What Changes

- Update the escrow execution public security doc to distinguish request validation from later execution rejection states.
- Add capability-spec coverage for the missing transaction receipt id validation error.
- Add security-docs-sync coverage so the public escrow execution doc stays aligned with that validation contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `escrow-execution`: the public contract now explicitly covers missing request-id validation before business rejection logic.
- `security-docs-sync`: escrow execution docs now stay aligned with the validation-vs-rejection split.

## Impact

- Affected docs: `docs/security/escrow-execution.md`
- Affected specs: `openspec/specs/escrow-execution/spec.md`, `openspec/specs/security-docs-sync/spec.md`
