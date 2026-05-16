## Why

After the recent inventory cleanup, every shipped top-level `internal/` package now appears in the public docs, but that parity is still protected only by many narrow slice-specific guards. A new top-level package could still be added later and silently appear in only one document.

## What Changes

- add the remaining missing `automation/`, `deadline/`, and `llm/` rows to `docs/architecture/project-structure.md`
- add a generalized executable guard that requires every shipped top-level `internal/` package to appear in both `README.md` and `docs/architecture/project-structure.md`
- sync the main docs-only and test-coverage specs

## Impact

- closes the last known top-level package gaps in the architecture inventory
- replaces manual spot-checking with a system-level parity invariant
- reduces future doc drift when new top-level packages are added
