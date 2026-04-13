## Why

ADK version references in docs did not match `go.mod`. `docs/architecture/index.md` said v0.4.0 while `go.mod` had v0.5.0+. This 3-way version mismatch undermines documentation credibility for a config-heavy runtime where small version differences matter.

## What Changes

- Fix stale ADK version reference in `docs/architecture/index.md` (v0.4.0 → v0.5.0)
- Add CI `docs-version-check` job that extracts ADK version from `go.mod` and fails if any docs reference a different version
- Scan for other dependency version drift in docs (none found)

## Capabilities

### New Capabilities

- `ci-docs-version-check`: Automated CI job that detects ADK version drift between `go.mod` and documentation

### Modified Capabilities

- `docs-architecture`: Corrected ADK version reference

## Impact

- `docs/architecture/index.md` — version fix (1 line)
- `.github/workflows/ci.yml` — new `docs-version-check` job (~18 lines)
