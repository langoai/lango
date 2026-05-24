## Why

The `p2p-onchain-examples` main spec already claimed a seven-example inventory, but its numbered summaries still omitted the shipped `p2p-trading` example entirely. That makes the examples index incomplete and inconsistent with the repository.

## What Changes

- add the missing `p2p-trading` summary to the numbered example list
- renumber the later example headings to preserve a truthful seven-example inventory
- add an executable guard so shipped example summaries cannot silently disappear again

## Impact

- main specs better match the actual shipped examples tree
- reduced confusion for maintainers using the examples index as a source of truth
- stronger regression protection for inventory drift
