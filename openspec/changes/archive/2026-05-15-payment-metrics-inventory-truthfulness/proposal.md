## Why

The architecture project-structure page and the README internal tree still described stale subsets of the current payment and metrics CLI surfaces. `payment x402` and `metrics policy` were already shipped, but the inventory summaries omitted them.

## What Changes

- update `docs/architecture/project-structure.md` so `cli/payment/` includes `x402`
- update the README internal tree so `metrics` includes `policy` and `payment` includes `x402`
- add a regression guard so those inventory summaries keep the current command surface

## Impact

- inventory-style docs better match the shipped payment and metrics CLI surface
- reduced confusion when readers inspect module ownership and command coverage
- stronger regression protection for architecture/README truthfulness
