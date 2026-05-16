## Why

The workbench starter prompts already know the repository, branch, and dirty-state, but they still stop short of pointing at the concrete files or directories that changed. That leaves the quick-start path one abstraction layer away from the operator's actual edit surface.

## What Changes

- Parse lightweight `git status --short` output to identify the top changed files or directories.
- Feed that summary into the ready-profile dirty-repository starter prompt.
- Keep the existing Git-aware fallback when no changed targets can be summarized.
- Update tests, docs, and specs to reflect the changed-target-aware quick-start behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Dirty-repository starter prompts now mention the most obvious changed files or directories when available.
- `downstream-docs-sync`: Public docs now describe changed-target-aware starter prompts.

## Impact

- Affected code: `internal/cli/workbench/starter_prompts.go`
- Affected tests: `internal/cli/workbench/starter_prompts_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
