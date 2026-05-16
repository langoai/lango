## Why

The `p2p-onchain-examples` main spec described a universal discovery-script polling-loop pattern, but the shipped `p2p-trading` example still uses a fixed `sleep 15` warm-up before checking peers. That makes the spec materially stronger than the current implementation.

## What Changes

- sync the `p2p-onchain-examples` main spec to the current mixed discovery-script reality
- add an executable guard so the stale universal polling claim cannot silently return

## Impact

- main specs better match the actual example scripts
- reduced confusion for maintainers reading the examples summary
- stronger regression protection for example-pattern drift
