# Proposal: Restore dev code lost during worktree file copy

## Problem

Worktree workers based on `origin/main` produced files missing `dev`-only code. Copying these files to the feature branch overwrote critical subsystem code: LoadResult struct, 35 TUI field handlers, 3 form factory cases, and migrate.go LoadResult handling.

## Fix

Restored 4 files from `dev` branch as base, then re-applied intentional additions (Context Engineering settings forms + handlers) on top.

### Restored code:
- `loader.go`: LoadResult struct, Load() returning (*LoadResult, error), context profile/auto-enable logic, Retrieval defaults, contextProfile validation
- `state_update.go`: 35 field handlers (Orchestration 8, RunLedger 9, Provenance 5, OS Sandbox 9, TraceStore 4)
- `setup_flow.go`: 3 form creator cases (runledger, provenance, os_sandbox)
- `migrate.go`: LoadResult-based config migration

### Preserved additions:
- Context allocation defaults in loader.go
- 22 new field handlers (Context Profile, Retrieval, Auto-Adjust, Context Budget)
- 4 new form factory cases
