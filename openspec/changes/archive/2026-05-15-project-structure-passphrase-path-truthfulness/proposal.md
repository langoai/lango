## Why

The architecture project-structure inventory still contains the deleted top-level `passphrase/` package path even though passphrase helpers now live under `internal/security/passphrase`.

## What Changes

- replace the stale `passphrase/` row with `security/passphrase/` in the architecture inventory
- add an executable guard that rejects reintroduction of the deleted package path
- sync the main docs-only and test-coverage specs

## Impact

- more truthful package-path documentation
- less confusion for maintainers navigating the codebase
- stronger regression protection against stale consolidated-package paths
