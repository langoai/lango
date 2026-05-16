## Why

The README internal tree still described the `graph/` package using a stale subset of the current CLI surface. The shipped graph CLI already exposed `add`, `export`, and `import`, but the inventory row still stopped at `status/query/stats/clear`.

## What Changes

- update the README internal tree `graph/` row to include `add`, `export`, and `import`
- add an executable guard so the README graph inventory keeps the current command surface

## Impact

- README inventory better matches the shipped graph CLI surface
- reduced confusion when readers inspect module ownership from the project tree
- stronger regression protection for README graph inventory truthfulness
