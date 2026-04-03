## Context

Three documentation files referenced different ADK versions (v0.4.0, v0.5.0) while `go.mod` had the actual version. Fast-moving projects accumulate this drift naturally, but config-heavy runtimes lose credibility when documentation disagrees with itself.

## Goals / Non-Goals

**Goals:**
- All ADK version references in docs match `go.mod`
- CI automatically catches future version drift

**Non-Goals:**
- Checking all dependency versions (only ADK is referenced with explicit versions in docs)
- Automated version bumping (CI detects, humans fix)

## Decisions

1. **grep + sed extraction** — Extract version from `go.mod` via `grep "google.golang.org/adk" go.mod | sed ...` rather than parsing go.mod formally. Rationale: simple, no extra tools needed, sufficient for CI check.

2. **Fail-fast CI job** — Separate job (not a step in existing build job) so version drift failures are clearly visible. Rationale: drift is a documentation issue, not a build issue.

## Risks / Trade-offs

- [grep-based extraction may break if go.mod format changes] → Low risk; go.mod format is stable.
