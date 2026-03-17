## Why

The exec tool allows AI agents to execute arbitrary shell commands, but lacks protection against commands that access the lango data directory (`~/.lango/`) or terminate system processes. An agent can run `sqlite3 ~/.lango/lango.db` to read/modify sensitive settings, `cat ~/.lango/keyfile` to extract the passphrase, or `kill 1` to terminate the server — none of which are caught by the existing `blockLangoExec` guard (which only blocks the `lango` CLI itself).

Additionally, the `SecurityFilterHook` has no default dangerous patterns and can be fully disabled via config, and configurable data paths can be pointed outside `~/.lango/` to escape the filesystem tool's existing protection.

## What Changes

- Add `CommandGuard` to the exec tool that blocks commands accessing protected data paths and process management commands (`kill`, `pkill`, `killall`)
- Add default catastrophic command patterns (`rm -rf /`, `mkfs`, `dd`, fork bomb, etc.) to `SecurityFilterHook` that are always active regardless of configuration
- Make `SecurityFilterHook` registration unconditional (not gated by `cfg.Hooks.SecurityFilter`)
- Add `DataRoot` config field enforcing all data paths reside under a single root directory
- Add `NormalizePaths` / `ValidateDataPaths` to the config loader pipeline
- Define `BlockedResult` struct replacing `map[string]interface{}` for typed blocked-command responses
- Add `AdditionalProtectedPaths` to `ExecToolConfig` for user-specified extra paths

## Capabilities

### New Capabilities
- `exec-command-guard`: Command-level security guard for exec tools — blocks protected path access and process management commands

### Modified Capabilities
- `tool-execution-hooks`: SecurityFilterHook now includes default blocked patterns and is always active
- `config-system`: DataRoot field added with path normalization and validation enforcement
- `tool-exec`: Exec/exec_bg handlers integrate CommandGuard and return typed BlockedResult

## Impact

- `internal/tools/exec/guard.go` — new CommandGuard with path and process verb checking
- `internal/toolchain/hook_security.go` — default patterns, pre-lowercased pattern matching
- `internal/app/tools_exec.go` — BlockedResult type, guard integration in exec/exec_bg handlers
- `internal/app/tools.go` — blockProtectedPaths helper
- `internal/app/app.go` — always-on SecurityFilterHook, CommandGuard wiring
- `internal/config/types.go` — DataRoot, AdditionalProtectedPaths fields
- `internal/config/loader.go` — expandTilde, NormalizePaths, ValidateDataPaths
