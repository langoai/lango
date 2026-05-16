## Why

The `p2p-trading-example` main spec still points at `contracts/MockUSDC.sol`, but the repository's shipped mock contract lives at `contracts/test/mocks/MockUSDC.sol`. That stale single-file reference makes the spec materially false and can mislead maintainers who use the spec as a source of truth.

## What Changes

- sync the `p2p-trading-example` main spec to the current mock contract path
- extend the executable broken-path guard so stale `contracts/MockUSDC.sol` claims cannot silently return

## Impact

- main specs better match the actual contracts tree
- reduced confusion for maintainers following the P2P trading example
- stronger regression protection for future spec drift
