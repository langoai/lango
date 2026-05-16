## Why

The `p2p-trading-example` main spec still described the integration test runner as `scripts/test-p2p-trading.sh`, but the shipped script lives under the example tree at `examples/p2p-trading/scripts/test-p2p-trading.sh`. That stale path sends maintainers to a nonexistent location.

## What Changes

- sync the `p2p-trading-example` main spec to the current integration test script path
- extend the executable broken-path guard so the stale top-level script path cannot silently return

## Impact

- main specs better match the actual example layout
- reduced confusion when maintainers follow the P2P trading example
- stronger regression protection for example-path drift
