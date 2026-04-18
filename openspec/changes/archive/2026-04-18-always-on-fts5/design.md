## Context

The runtime already imports `modernc.org/sqlite` through `internal/sqlitedriver`, and targeted package tests show FTS5 works without an explicit build tag in the current environment. The remaining `fts5` tag usage lives in build commands, container build instructions, contributor documentation, and test skip messages that still frame FTS5 as optional.

The desired end state is simple: normal `go build ./...` and `go test ./...` represent the default FTS5-capable runtime. Optional `vec` integration remains tag-gated because it still depends on the legacy sqlite-vec path.

## Goals / Non-Goals

**Goals:**
- Make FTS5 part of the default build/test/runtime contract.
- Remove `fts5` from default build tooling and documentation.
- Keep optional `vec` guidance accurate.
- Preserve current runtime behavior and tests.

**Non-Goals:**
- Remove `vec` integration.
- Rewrite embedding/vector architecture.
- Change KMS tag behavior.
- Remove runtime FTS5 probing and fallback logic in this slice.

## Decisions

### 1. Treat FTS5 as a default runtime contract

Default Makefile and Docker builds stop passing `-tags "fts5"`. Contributor and installation docs also stop requiring it.

### 2. Keep runtime probe logic for safety

Even though FTS5 is now considered always-on for supported builds, runtime probe/fallback logic stays in place. That avoids coupling this cleanup change to a larger failure-mode redesign.

### 3. Keep `vec` as the only optional legacy build tag

Docs and examples shift from `fts5,vec` to `vec` for vector-only legacy builds. This keeps the optional path explicit without implying FTS5 is optional too.

## Risks / Trade-offs

- **[Risk] Some environments may still lack FTS5 support unexpectedly** → Mitigation: keep probe/fallback logic and avoid removing runtime checks in this slice.
- **[Risk] Docs drift if all examples are not updated together** → Mitigation: update README, development docs, install docs, and Docker instructions in the same change.
- **[Risk] Tests still mention tag-gated behavior** → Mitigation: update skip/failure messaging to reflect default runtime expectations.
