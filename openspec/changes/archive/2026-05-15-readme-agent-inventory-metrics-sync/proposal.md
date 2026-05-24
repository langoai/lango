## Why

The README internal tree still summarized the `agent/` package as `status/list/tools/hooks/trace/graph`, which hid the shipped `trace metrics` surface even after the actual CLI and architecture inventory were synced.

## What Changes

- update the README internal tree `agent/` row to include `trace list-show-metrics/graph`
- update the existing A2A/agent inventory guard and corresponding main-spec wording to enforce that current README inventory contract

## Impact

- README inventory better matches the shipped agent diagnostics surface
- reduced confusion when readers inspect command coverage from the project tree
- stronger regression protection for README agent inventory drift
