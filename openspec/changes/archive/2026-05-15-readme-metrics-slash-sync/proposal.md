## Why

The README internal CLI inventory still describes the metrics family with bracket shorthand, even though the rest of the inventory now favors explicit slash-separated command surfaces.

## What Changes

- update the README internal tree metrics row to `lango metrics/sessions/tools/agents/policy/history`
- sync the existing payment/metrics inventory guard and main specs with that slash-form wording

## Impact

- more consistent README inventory style
- clearer mapping from the inventory to real command paths
- stronger regression protection against stale metrics shorthand
