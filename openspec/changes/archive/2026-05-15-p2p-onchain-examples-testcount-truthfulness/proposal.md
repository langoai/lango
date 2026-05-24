## Why

The `p2p-onchain-examples` main spec still summarized each example with exact `Tests (N)` counts, but the shipped scripts are evolving shells whose section and assertion counts do not cleanly match those hard-coded numbers. That makes the spec stronger than the repository evidence.

## What Changes

- replace stale exact `Tests (N)` wording in the `p2p-onchain-examples` main spec with representative check summaries
- extend the executable spec-quality guard so stale exact test-count claims cannot silently return

## Impact

- main specs better match the actual shipped example scripts
- reduced confusion for maintainers using the examples summary as a source of truth
- stronger regression protection for drifting test-count prose
