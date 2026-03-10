# Tasks: P2P Workspace Downstream Artifacts

## WU-1: TUI Settings — Workspace Config Form
- [x] 1.1 Add `NewP2PWorkspaceForm()` to `forms_p2p.go` with 7 fields and VisibleWhen pattern
- [x] 1.2 Add `p2p_workspace` menu entry to `menu.go` under P2P Network section
- [x] 1.3 Add `case "p2p_workspace":` to `editor.go` handleMenuSelection

## WU-2: Doctor Check — Workspace Diagnostic
- [x] 2.1 Create `workspace.go` with `WorkspaceCheck` (Name, Run, Fix)
- [x] 2.2 Register `&WorkspaceCheck{}` in `AllChecks()`

## WU-3: Tool Catalog — Workspace Category
- [x] 3.1 Add "workspace" entry to `buildToolCategories()` in `catalog.go`

## WU-4: Docs — CLI P2P Reference
- [x] 4.1 Add `lango p2p workspace` section with 5 subcommands (create, list, status, join, leave)
- [x] 4.2 Add `lango p2p git` section with 5 subcommands (init, log, diff, push, fetch)

## WU-5: Docs — P2P Network Feature Overview
- [x] 5.1 Add "Collaborative Workspaces" section (lifecycle, members, messages, chronicler, contributions)
- [x] 5.2 Add "Git Bundle Exchange" section (bare repos, protocol, workflow, DAG leaves)
- [x] 5.3 Add workspace config block to Configuration section
- [x] 5.4 Add workspace/git CLI commands to CLI Commands section

## WU-6: README.md Update
- [x] 6.1 Add P2P Workspaces feature to features list
- [x] 6.2 Add 10 workspace/git CLI commands

## WU-7: Prompts Update
- [x] 7.1 Add "P2P Workspace Tool" section to TOOL_USAGE.md (12 tools + workflow)
- [x] 7.2 Update tool category count in AGENTS.md ("thirteen" → "fourteen")
- [x] 7.3 Add P2P Workspace category bullet to AGENTS.md

## WU-8: Unit Tests — Core Packages
- [x] 8.1 Create `manager_test.go` (16 tests: CRUD, lifecycle, messaging, members)
- [x] 8.2 Create `contribution_test.go` (5 tests: record, get, list, remove)
- [x] 8.3 Create `chronicler_test.go` (5 tests: record, parent, metadata, nil adder, error)
- [x] 8.4 Create `store_test.go` (7 tests: init, idempotent, repo, not found, path, list, remove)
- [x] 8.5 Create `bundle_test.go` (4 tests: init, log empty, leaves empty, bundle empty)
- [x] 8.6 Fix CreateBundle locale-insensitive error detection (exit code 128 fallback)

## WU-9: Docker & Makefile
- [x] 9.1 Add commented workspace volume and env var to docker-compose.yml
- [x] 9.2 Add `test-workspace` Makefile target
- [x] 9.3 Update `.PHONY` with `test-workspace`
