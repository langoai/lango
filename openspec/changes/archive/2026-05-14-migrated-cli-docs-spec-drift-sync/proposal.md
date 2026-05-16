## Why

Several migrated CLI surfaces already use explicit `--output table|json` contracts in code, but a few public docs and older main specs still referenced `--json`. That drift makes the product look less coherent than the implementation actually is and weakens the executable docs guard story.

## What Changes

- sync stale public docs and legacy main specs to the current output-format contracts
- update executable docs guards so these stale `--json` references do not reappear

## Impact

- public docs better match the current CLI UX
- less confusion for operators copying examples
- stronger regression protection for migrated output contracts
