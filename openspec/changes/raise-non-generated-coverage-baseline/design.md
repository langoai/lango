# Design

## Coverage Definition
The coverage target applies to Go statements in repository-owned, non-generated code. Generated files are excluded when they are under known generated directories such as `internal/ent/` or contain standard generated-file markers such as `Code generated` with `DO NOT EDIT`.

The target is:

```text
covered_non_generated_statements / total_non_generated_statements >= 90%
```

## Measurement
A repeatable command or tool will create a coverage profile and report:

- overall non-generated coverage percentage
- covered statements
- total statements
- uncovered statements
- top files by uncovered statement count

The top-files list matters because low-percentage small files do not move the global number enough. Coverage work should prioritize high uncovered-statement files first unless a smaller file is blocking correctness or quality.

## Baseline Hotspots
The first baseline identified these largest non-generated files by uncovered statement count:

| Rank | File | Coverage | Uncovered | Total |
| --- | --- | ---: | ---: | ---: |
| 1 | `internal/cli/tuicore/state_update.go` | 0.00% | 624 | 624 |
| 2 | `internal/storagebroker/server.go` | 18.98% | 431 | 532 |
| 3 | `internal/adk/agent.go` | 10.02% | 377 | 419 |
| 4 | `internal/app/wiring_p2p.go` | 0.92% | 324 | 327 |
| 5 | `internal/knowledge/store.go` | 53.12% | 323 | 689 |
| 6 | `internal/cli/settings/editor.go` | 16.45% | 320 | 383 |
| 7 | `internal/app/tools_meta.go` | 62.11% | 280 | 739 |
| 8 | `internal/storage/facade.go` | 6.20% | 227 | 242 |
| 9 | `internal/app/wiring.go` | 42.64% | 226 | 394 |
| 10 | `internal/p2p/handshake/handshake.go` | 8.09% | 216 | 235 |

These files are starting targets, not a license to write shallow execution tests. If a hotspot is hard to test safely, split out deterministic seams first and document the reason in the batch commit.

## Implementation Strategy
Work proceeds in batches, each producing a reviewable commit:

1. Coverage measurement tooling and baseline report.
2. Small deterministic packages and pure helpers.
3. CLI/TUI state and settings surfaces: `internal/cli/tuicore/state_update.go`, `internal/cli/settings/editor.go`, `internal/cli/settings/menu.go`, and adjacent command seams.
4. Storage, knowledge, and broker packages: `internal/storage/facade.go`, `internal/storagebroker/server.go`, `internal/storagebroker/client.go`, `internal/knowledge/store.go`.
5. Runtime, workflow, and integration-heavy packages: `internal/adk/agent.go`, `internal/workflow/engine.go`, `internal/turnrunner/runner.go`, `internal/turntrace/store.go`.
6. App wiring and P2P hotspots: `internal/app/wiring_p2p.go`, `internal/app/wiring.go`, `internal/app/tools_meta.go`, `internal/p2p/handshake/handshake.go`.
7. Final threshold enforcement once the measured coverage is at or above 90%.

Each batch should follow TDD:

- write failing tests for uncovered behavior
- verify the tests fail for the intended reason
- implement only the needed test seams or assertions
- run focused package tests
- run repository verification before committing

## Risk Management
Coverage can be gamed by shallow execution tests. Reviews must check that new tests assert behavior, error paths, boundaries, and deterministic state transitions.

Generated-code exclusion must be explicit and auditable. If a file is excluded because it is generated, the exclusion reason must be mechanically discoverable from either its path or generated-file marker.

## Enforcement
After coverage reaches at least 90%, the repository must include an executable gate that fails below the threshold. The gate should be suitable for local use and CI, and must use the same generated-code exclusion rules as the report.
