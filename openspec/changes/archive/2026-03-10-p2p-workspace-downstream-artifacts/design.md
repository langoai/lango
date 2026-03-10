# Design: P2P Workspace Downstream Artifacts

## Approach

All 9 work units are independent and follow existing patterns in the codebase. No new abstractions or architectural decisions are needed.

## Key Decisions

### TUI Form Pattern
Follow the `NewP2PSandboxForm` pattern with `VisibleWhen` closures for conditional field visibility. The workspace enabled field controls visibility of chronicler, auto-sandbox, and contribution tracking fields.

### Doctor Check Pattern
Follow the `A2ACheck` pattern. The check skips when workspace is disabled, warns when git is missing, and offers auto-fix for missing data directories. Uses `os.MkdirAll` with `0o700` permissions.

### Test Strategy
Table-driven tests with `t.TempDir()` for BoltDB and bare git repos. Git-binary-dependent tests use `skipIfNoGit` helper. Mock `TripleAdder` for chronicler tests.

### Bundle Locale Bug Fix
The `CreateBundle` method's error detection for empty repos relied on English error messages. Added fallback detection using `exec.ExitError` with exit code 128 for locale-independent behavior.

## File Changes

| Area | Files Modified | Files Created |
|------|---------------|---------------|
| TUI | forms_p2p.go, menu.go, editor.go | — |
| Doctor | checks.go | workspace.go |
| Catalog | catalog.go | — |
| Docs | docs/cli/p2p.md, docs/features/p2p-network.md | — |
| README | README.md | — |
| Prompts | TOOL_USAGE.md, AGENTS.md | — |
| Tests | — | manager_test.go, contribution_test.go, chronicler_test.go, store_test.go, bundle_test.go |
| Infra | docker-compose.yml, Makefile | — |
| Bugfix | bundle.go | — |
