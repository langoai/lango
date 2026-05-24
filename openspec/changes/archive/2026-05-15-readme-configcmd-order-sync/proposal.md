## Why

The architecture inventory already reflects the current `configcmd` command surface, but the README internal tree still uses a stale order that places `validate` before `get`, `set`, and `keys`.

## What Changes

- update the README internal tree `configcmd` row to the current command order
- extend the existing config inventory guard so it verifies the README order too
- sync the main docs-only and test-coverage specs

## Impact

- more truthful README inventory ordering
- better consistency between architecture inventory, README, and the CLI index
- stronger regression protection against stale command ordering
