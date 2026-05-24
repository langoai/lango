## Why

The public docs now cover the full `internal/cli/` subtree, but that parity is still protected only by many narrow command-family checks. A new CLI helper or command package could still be added later and silently appear in only one inventory.

## What Changes

- add a generalized executable guard that requires every shipped `internal/cli/` subpackage to appear in both `README.md` and `docs/architecture/project-structure.md`
- sync the main docs-only and test-coverage specs

## Impact

- converts current CLI subtree completeness into a maintained invariant
- protects helper-package discoverability as well as command-family discoverability
- reduces future CLI inventory drift
