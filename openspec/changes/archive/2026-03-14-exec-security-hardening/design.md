## Context

The exec tool (`internal/app/tools_exec.go`) delegates shell commands to the Supervisor. The existing `blockLangoExec` guard only blocks `lango` CLI invocations. The filesystem tool blocks `~/.lango/` via `BlockedPaths`, but this protection does not extend to the exec tool. The `SecurityFilterHook` checks user-configured patterns but has no defaults and can be disabled via `cfg.Hooks.SecurityFilter`.

## Goals / Non-Goals

**Goals:**
- Block exec commands that access the lango data directory or any configured data path
- Block process management commands (`kill`, `pkill`, `killall`) from exec tools
- Provide always-on default dangerous command patterns in SecurityFilterHook
- Enforce that all configurable data paths reside under a single DataRoot
- Use typed structs for tool responses instead of `map[string]interface{}`

**Non-Goals:**
- Full shell parsing or sandboxing (handled by P2P sandbox system)
- Blocking read access to non-data files (e.g., source code, /etc/hosts)
- Replacing the existing approval flow for dangerous tools

## Decisions

### Decision 1: Command Guard as a separate struct in exec package
CommandGuard lives in `internal/tools/exec/guard.go` alongside the exec tool it protects. It receives protected paths at construction time and resolves them to absolute form. The guard uses substring matching on the normalized command string rather than full shell parsing.

**Alternative considered:** Integrating guard logic directly into SecurityFilterHook. Rejected because the hook operates on pattern matching (generic substrings) while the guard needs path resolution and verb extraction (domain-specific).

### Decision 2: Pre-built strings.Replacer for command normalization
The guard pre-builds a `strings.Replacer` at construction time for `$HOME`, `${HOME}`, and tilde-at-word-boundary replacements. This avoids repeated string allocations on every exec invocation.

### Decision 3: SecurityFilterHook always active
The hook registration is moved out of the `cfg.Hooks.Enabled` conditional block. Default patterns are merged with user patterns at construction time, with case-insensitive deduplication. Patterns are pre-lowercased at construction for O(1) matching in the hot path.

### Decision 4: DataRoot with path normalization pipeline
A `DataRoot` field in Config (default `~/.lango/`) serves as the single root for all data paths. `NormalizePaths()` expands tildes and resolves relative paths under DataRoot. `ValidateDataPaths()` verifies all paths are under the root. Both run in the Load pipeline before Validate.

### Decision 5: BlockedResult struct
Replace `map[string]interface{}{"blocked": true, "message": msg}` with `BlockedResult{Blocked: true, Message: msg}` for type safety. The handler return type is `interface{}`, so the struct is fully compatible.

## Risks / Trade-offs

- **Heuristic matching** — Substring matching can have false positives (e.g., a file named `~/.lango-backup/` would be blocked). Mitigation: protected paths are resolved to absolute form, reducing ambiguity.
- **Shell escaping bypass** — A sophisticated command could encode paths to bypass string matching (e.g., hex-encoded). Mitigation: this guard is a defense-in-depth layer; the approval flow still applies to all dangerous tools.
- **DataRoot flexibility** — Users can change DataRoot (e.g., for Docker), but all sub-paths must stay under it. This prevents splitting data across directories. Mitigation: `AdditionalProtectedPaths` allows protecting extra locations.
