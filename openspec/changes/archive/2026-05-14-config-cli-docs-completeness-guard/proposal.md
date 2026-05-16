## Why

The public config CLI docs were recently expanded to include `get`, `set`, and `keys`, but that completeness can regress easily because the commands are spread across README and CLI docs.

## What Changes

- add an executable docs guard that requires `config get`, `config set`, and `config keys` to remain present in README and public CLI docs
- sync docs-only and test-coverage specs to describe the guard

## Impact

- keeps the documented config CLI surface aligned with the implemented command set
- prevents silent discoverability regressions
