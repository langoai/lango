## Why

The `p2p-trading-example` main spec still described the bootstrap entrypoint as `docker-entrypoint-p2p.sh`, but the shipped entrypoint lives under the example tree at `examples/p2p-trading/docker-entrypoint-p2p.sh`. That stale path makes the spec less trustworthy for maintainers following the example.

## What Changes

- sync the `p2p-trading-example` main spec to the current Docker entrypoint path
- extend the executable broken-path guard so the stale bare entrypoint path cannot silently return

## Impact

- main specs better match the actual example layout
- reduced confusion when maintainers inspect the P2P trading bootstrap flow
- stronger regression protection for example-path drift
