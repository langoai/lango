## Why

The on-chain escrow wrapper hardening previously locked only `escrow_create` and `escrow_resolve`, leaving the rest of the `escrow_*` surface without exact missing-parameter regressions. That leaves room for drift back toward generic downstream failures before escrow lookup, dispute mutation, or settlement execution begins.

## What Changes

- Add exact missing-parameter regressions for `escrow_fund`, `escrow_activate`, `escrow_submit_work`, `escrow_release`, `escrow_refund`, `escrow_dispute`, and `escrow_status`.
- Update prompt/public docs to describe the broader on-chain escrow required-input contract.
- Sync the on-chain escrow and production-readiness specs to the same fail-closed contract.

## Impact

- `onchain-escrow`: the operator-facing tool cluster is more uniformly regression-covered.
- `production-readiness`: on-chain escrow wrapper semantics now cover the full live tool surface instead of only two entrypoints.
