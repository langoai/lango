# Proposal: P2P Workspace & Git Bundle Downstream Artifacts

## Summary

Update all downstream artifacts to reflect the P2P workspace management and git bundle exchange features added in the `feature/p2p-agent-cowork` branch. The core implementation (~3,300 lines across 36 files) is complete but downstream artifacts (TUI settings, doctor checks, docs, README, prompts, tests, Docker, Makefile, tool catalog) were not yet updated.

## Motivation

The P2P workspace and git bundle features are fully implemented in `internal/p2p/workspace/` and `internal/p2p/gitbundle/`, with CLI commands in `internal/cli/p2p/`, config types in `internal/config/types_p2p.go`, and wiring in `internal/app/`. However, users cannot discover or configure these features through the TUI settings editor, diagnose issues via the doctor command, or find documentation about them. Tests are also missing for the core packages.

## Scope

9 independent work units:

1. **TUI Settings** — `NewP2PWorkspaceForm` with 7 fields, menu entry, editor routing
2. **Doctor Check** — `WorkspaceCheck` with git binary, data dir, and config validation
3. **Tool Catalog** — "workspace" category entry in `buildToolCategories()`
4. **Docs — CLI Reference** — `lango p2p workspace` (5 subcommands) + `lango p2p git` (5 subcommands)
5. **Docs — Feature Overview** — Collaborative Workspaces + Git Bundle Exchange sections
6. **README.md** — Feature list entry + 10 CLI commands
7. **Prompts** — 12 workspace/git tool descriptions + category count update
8. **Unit Tests** — 5 test files, 37 tests covering Manager, ContributionTracker, Chronicler, BareRepoStore, Service
9. **Docker & Makefile** — Commented workspace env/volume + `test-workspace` target

## Non-Goals

- No new core features or API changes
- No runtime integration tests (all changes are compile-time verifiable + unit-testable)
- No changes to the workspace/gitbundle core packages (except a locale-insensitive bugfix in CreateBundle)
