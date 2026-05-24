## Why

The on-chain escrow wrapper surface is much better covered now, but `escrow_create` and `escrow_resolve` still do not directly lock all of their required wrapper inputs. That leaves room for regressions where some required fields might drift away from exact missing-parameter behavior without immediate detection.

## What Changes

- Add exact missing-parameter regressions for `escrow_create` required `buyerDid`, `sellerDid`, and `amount`.
- Add an exact missing-parameter regression for `escrow_resolve` required `favor`.
- Sync the on-chain escrow and production-readiness specs to state that the full required-input set is covered at the wrapper boundary.

## Impact

- `onchain-escrow`: create/resolve required-input coverage is more complete.
- `production-readiness`: the spec no longer relies on partial evidence for those entrypoints.
