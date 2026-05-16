## Why

The `p2p-onchain-examples` main spec still claimed there were six Docker Compose integration examples, but the current repository ships seven example directories. That stale inventory count makes the spec less trustworthy as a repository summary.

## What Changes

- sync the `p2p-onchain-examples` main spec to the current seven-example inventory
- extend the executable spec-quality guard so the stale count cannot silently return

## Impact

- main specs better match the actual shipped example set
- reduced confusion for maintainers using the examples summary as source of truth
- stronger regression protection for example-inventory drift
