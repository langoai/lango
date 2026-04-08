## Context

PR 4 (`sandbox-escape-hardening`) landed on 2026-04-08 with full coverage across `DefaultToolPolicy`, `MCPServerPolicy`, exec/skill/MCP publish sites, Recent Sandbox Decisions in `lango sandbox status`, and a TUI field for `ExcludedCommands`. Directly after archive, a Codex review flagged four concrete regressions. Three of them are contract violations (PR 4 docs/specs claim behaviors that the code does not actually implement) and one is a UX regression (`sandbox status` prompts for the passphrase twice).

This fix PR is intentionally narrow. It does not introduce new backends, does not expand `ExcludedCommands`, does not touch Landlock/seccomp, and does not modify the `SandboxDecisionEvent` schema. It only makes PR 4 behave as documented.

## Goals / Non-Goals

**Goals:**
- Seatbelt `DenyPaths` entries actually hide their target files from sandboxed children, not just block writes.
- `.git` baseline deny does not break sandbox apply in non-repo or worktree workspaces.
- `lango sandbox status` runs exactly one `bootstrap.Run` call per invocation.
- `skill.Executor.executeScript` always publishes a `SandboxDecisionEvent` before returning, even on the fail-closed-without-isolator path.
- No new features. No schema migrations. No TUI changes.

**Non-Goals (deferred to PR 5):**
- `bwrap --version` probe false positives on hosts with `kernel.unprivileged_userns_clone=0`. The probe currently only confirms the binary exists.
- Native Landlock+seccomp backend on Linux.
- File-level deny via `--ro-bind /dev/null <file>` (would allow protecting worktree `.git` files).
- Symlink chain resolution before bwrap arg compile.
- Glob / path semantics normalization between Linux and macOS.
- Per-tool policy overrides.

## Decisions

### D1 — Seatbelt template emits read+write deny for DenyPaths

**Problem**: `seatbelt_profile.go` renders only `(deny file-write* (subpath "{{.}}"))` for `DenyPaths`. When a policy sets `ReadOnlyGlobal=true` the template also emits `(allow file-read*)` which grants blanket read across the entire filesystem, so sandboxed children still read `~/.lango/lango.db`, `.git/config`, and encrypted config tokens.

**Decision**: Add `(deny file-read* (subpath "{{.}}"))` alongside the existing write-deny rule. Seatbelt applies `(deny ...)` with precedence over blanket allow rules, so adding the read-deny line hides the path from sandboxed children.

**Why not split the policy into `DenyReadPaths` and `DenyWritePaths`**: Every PR 4 caller (`DefaultToolPolicy`, `StrictToolPolicy`, `MCPServerPolicy`) treats control-plane and `.git` as "fully invisible to the child". Splitting the slice would require updating every call site without any behavior change and would invite future confusion. The Linux path already hides the target bidirectionally via `--tmpfs <path>` (verified in `bwrap_args.go:80-93`).

**Alternatives considered**:
- Switch `ReadOnlyGlobal` to `false` and enumerate `ReadPaths` explicitly. Rejected: read-global is the sane default for tool execution and Seatbelt already supports the `(deny file-read* (subpath ...))` override natively.
- Generate multiple `(deny file*)` rules via a wildcard. Rejected: Seatbelt's `file*` operator shorthand is not documented as reliable across macOS versions; explicit `file-read*` + `file-write*` is guaranteed.

### D2 — `.git` and `dataRoot` baseline deny gated on `isDir`

**Problem**: `DefaultToolPolicy` and `MCPServerPolicy` push `.git` / `dataRoot` into `DenyPaths` without checking that the path exists as a directory. `bwrap_args.go:85-91` does `os.Stat` + `IsDir()` and rejects every missing or non-directory deny entry. Non-repo workspaces (no `.git` at all) and linked worktrees (`.git` is a file) both produce `compileBwrapArgs` errors, which propagate to `Apply` → `failClosed=true` rejects every command, `failClosed=false` silently falls back.

**Decision**: Introduce a private `isDir(p string) bool` helper in `policy.go` and gate `.git` and `dataRoot` additions on it. Missing entries are silently dropped. This keeps the policy buildable in every environment the runtime might encounter.

**Worktree trade-off**: In a linked worktree, `.git` is a file containing `gitdir: /path/to/real/.git`. After this fix, `.git` is simply dropped from `DenyPaths` in worktree checkouts. The real `.git` directory (somewhere else on disk) is still readable via the global read mount but not writable (it is not in `WritePaths`). We accept this trade-off for two reasons:
1. The control-plane mask (`~/.lango`) — the core leak protection — is unaffected and still works.
2. File-level deny (`--ro-bind /dev/null <file>`) is the proper fix and is scheduled for PR 5, which introduces the backend robustness work.

**Why not follow the `gitdir:` pointer to deny the real location**: Increases the escape surface (a malicious workspace could rewrite `.git` to point at a sensitive directory and trick the sandbox into denying it, breaking tools unexpectedly). Not worth the complexity.

### D3 — Skill executor publishes nil-isolator decision before temp file creation

**Problem**: `executor.go` had an early return at `isolator == nil && failClosed` that unconditionally returned `ErrSandboxRequired` without publishing the `SandboxDecisionEvent`. A second branch further down the function was supposed to handle the same case (`else if e.failClosed { publishSandboxDecision(..., "rejected", ...) }`) but was unreachable because of the early return. The `sandbox-exception-policy/spec.md` requirement "Skill fail-closed publishes" was therefore violated by PR 4 from day one.

**Decision**: Delete the early return. Move the nil-isolator decision (both `rejected` and `skipped` branches) above temp file creation. Remove the now-dead else-if branch at the bottom. This yields a single publish path per decision and has the side benefit of avoiding unnecessary temp file allocation on the reject path.

**Why move it up rather than just add a publish call to the old early return**: Two nil-isolator publish sites in the same function were the source of the bug. Consolidation is cleaner and maintainable.

### D4 — `sandbox status` uses bootLoader only

**Problem**: `cliboot.Config()` and `cliboot.BootResult()` each call `bootstrap.Run()` independently with no caching. A single `lango sandbox status` invocation called both (cfgLoader for the config section, bootLoader for the Recent Decisions section), triggering two full bootstrap passes and two passphrase prompts.

**Decision**: Change `newStatusCmd` to take only `bootLoader`. Derive `cfg` from `boot.Config`. `defer boot.DBClient.Close()` to match the convention used elsewhere in `cmd/lango/main.go`. Pass the already-resolved `*bootstrap.Result` directly to `renderRecentDecisions` instead of letting the helper re-invoke a loader. The outer `NewSandboxCmd` signature stays unchanged (`newTestCmd` still uses `cfgLoader` because smoke tests do not need the audit DB).

**Why not cache bootstrap.Run via sync.Once in cliboot**: Cleaner but more invasive (affects every CLI subcommand that calls both `Config()` and `BootResult()`, changes `Config()`'s DB close semantics). The surgical fix in sandbox.go is enough to address the reported regression.

**Resilience analysis**: The old code failed whenever `Config()` failed, regardless of whether `BootResult()` succeeded. The new code fails whenever `bootLoader()` fails. Both paths call the same `bootstrap.Run()` against the same DB file, so the failure envelope is identical.

## Risks / Trade-offs

- **Worktree `.git` is no longer denied** → worktree users lose `.git` baseline protection on both macOS and Linux. Mitigation: control-plane (`~/.lango`) deny is unaffected. Proper fix (file-level deny) is in PR 5.
- **`bwrap --version` probe still returns false positives** on hosts with `kernel.unprivileged_userns_clone=0`. Mitigation: design.md explicitly marks this as PR 5 scope; runtime fallback behavior (fail-closed rejects / fail-open warns) still kicks in when apply actually fails.
- **Test coverage for Seatbelt enforcement** stays at the template-level assertion (generated profile string contains the read-deny line). We trust macOS to enforce what sandbox-exec parses. The existing `lango sandbox test` 4 smoke tests continue to validate runtime sandbox-exec behavior end-to-end.
- **Stage 2 test rewrite** replaces string-literal paths (`/home/user/project`) with `t.TempDir()` + real `.git` directory. Slightly more complex but reflects the `isDir` guard's actual behavior.

## Migration Plan

No schema migration, no config migration, no user action required. The fix is purely code + tests. On first run after deployment:
- macOS Seatbelt profile generation picks up the new template automatically.
- Linux bwrap `compileBwrapArgs` stops rejecting non-repo workspaces.
- `sandbox status` runs bootstrap once.
- skill audit records fail-closed-without-isolator rejections.

Rollback: revert the commits. No data migration to undo.
