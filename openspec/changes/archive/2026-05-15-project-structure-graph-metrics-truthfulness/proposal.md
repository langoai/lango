## Why

The public architecture project-structure reference still described `cli/graph/` and `cli/metrics/` using stale subsets of the current command surface. That no longer matched the shipped CLI after graph add/export/import and metrics policy became part of the operator-facing surface.

## What Changes

- update `docs/architecture/project-structure.md` so `cli/graph/` includes `add`, `export`, and `import`
- update the same page so `cli/metrics/` includes `policy`
- add a regression guard so those architecture rows keep the current command inventory

## Impact

- architecture docs better match the shipped graph and metrics command surface
- reduced confusion when readers inspect module ownership from the project-structure page
- stronger regression protection for architecture-doc truthfulness
