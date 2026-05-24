## Why

The project already uses a pure-Go SQLite runtime where FTS5 works in normal builds, but the build surface and docs still present `fts5` as a required opt-in tag. That is now misleading and makes installation, CI, and contributor workflows more complicated than the actual runtime requires.

## What Changes

- Remove `fts5` from default build and test commands so FTS5 is treated as always-on in the standard runtime.
- Keep `vec` as the only optional build tag for legacy sqlite-vec integration.
- Update docs, README, Docker build instructions, and test messaging to describe FTS5 as built-in rather than tag-gated.

## Capabilities

### New Capabilities
- `always-on-fts5`: The default runtime always includes FTS5 support without requiring a dedicated build tag.

### Modified Capabilities
- `fts5-search-index`: FTS5 is now part of the default runtime contract instead of a tag-gated build variant.
- `docker-deployment`: Default container builds no longer pass `-tags "fts5"`.
- `project-docs`: Build and installation docs now describe FTS5 as always enabled and `vec` as the only optional legacy tag.

## Impact

- Affected code and build surfaces:
  - [Makefile](/Users/juwonkim/GolandProjects/lango/Makefile)
  - [Dockerfile](/Users/juwonkim/GolandProjects/lango/Dockerfile)
  - FTS5-related tests and messaging under `internal/search`, `internal/session`, and `internal/knowledge`
- Affected docs:
  - [README.md](/Users/juwonkim/GolandProjects/lango/README.md)
  - [docs/getting-started/installation.md](/Users/juwonkim/GolandProjects/lango/docs/getting-started/installation.md)
  - [docs/development/build-test.md](/Users/juwonkim/GolandProjects/lango/docs/development/build-test.md)
  - [docs/development/index.md](/Users/juwonkim/GolandProjects/lango/docs/development/index.md)
