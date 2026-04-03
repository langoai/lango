## Why

Phase 1-3 established a trustworthy, verified, and observable runtime. However, the public-facing surface still leads with "sovereign economic stack" messaging rather than the runtime's actual strengths, experimental features lack clear maturity indicators in the TUI, and installation documentation is missing platform-specific prerequisite guidance. Phase 4 reduces adoption friction so new users can understand what is stable, what is experimental, and how to get started quickly.

## What Changes

- Restructure README.md: shift lead message to "trustworthy multi-agent runtime", move early-stage warning to top, reorder "Why Lango?" to lead with trust/orchestration/observability, condense inline CLI reference from 180 lines to 8-command summary with docs link
- Enhance installation documentation with platform-specific C compiler setup (macOS, Ubuntu, Fedora, Alpine) and `go install` vs `make build` differences
- Add `[EXP]` badge to TUI settings menu for experimental feature categories, with `@experimental` search filter and drift-prevention test
- Update roadmap documentation to reflect Phase 1-3 completion status and Phase 4 progress

## Capabilities

### New Capabilities

_(none — this change enhances existing surfaces without introducing new spec-level capabilities)_

### Modified Capabilities

_(no spec-level requirement changes — all modifications are documentation, TUI presentation, and developer experience improvements that don't alter behavioral contracts)_

## Impact

- `README.md` — lead message, warning placement, Why Lango order, CLI block reduction (~170 lines removed)
- `docs/getting-started/installation.md` — platform-specific prerequisite additions
- `docs/getting-started/quickstart.md` — cross-link enhancement
- `docs/development/roadmap.md` — Phase 1-3 completion status, backlog updates
- `internal/cli/tui/styles.go` — new `BadgeExperimentalStyle`
- `internal/cli/settings/menu.go` — `ExperimentalCategories` map, `renderItem` badge, `@experimental` filter
- `internal/cli/settings/menu_test.go` — drift-prevention test
