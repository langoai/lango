## Why

`sentinel_acknowledge` mutates alert state, but its tool capability metadata is still marked as `ActivityQuery` and `ReadOnly=true`. That is a real classification bug: routing, policy, and observability layers should treat it as a dangerous management mutation, not a read-only query.

## What Changes

- Reclassify `sentinel_acknowledge` as a management/write tool instead of a query/read-only tool.
- Add a regression that locks the updated capability metadata.
- Sync prompt/docs/spec wording so the dangerous write-path contract is explicit.

## Impact

- `escrow-sentinel`: capability metadata matches actual mutation behavior.
- Operator/runtime reasoning: dangerous alert acknowledgment is no longer mislabeled as a read-only query.
