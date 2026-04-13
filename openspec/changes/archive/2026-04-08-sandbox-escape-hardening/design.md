## Context

PR 1 (`linux-sandbox-hardening`, archived 2026-04-05) restored the Linux build, refactored isolators around a `Reason()` interface, and rewrote `lango sandbox status` to be honest about what each platform can do. PR 2 (`sandbox-backend-registry`, archived 2026-04-06) introduced a typed backend registry, brought MCP and skill executor under the same fail-closed contract, and exposed backend selection to CLI/TUI. PR 3 (`linux-bwrap-backend`, archived 2026-04-07) shipped the first real Linux isolation backend (`bubblewrap`), wrote a Policy → bwrap argv compiler, replaced kernel probe stubs with real `golang.org/x/sys/unix` syscalls, and built a 4-case smoke test suite (`lango sandbox test`) that runs identically on macOS and Linux.

What remained after PR 3 was a set of operational gaps in the *existing* sandbox layer rather than missing backends:

1. The `DefaultToolPolicy` and `MCPServerPolicy` helpers used `ReadOnlyGlobal: true`, which mounts the entire host read-only into the sandboxed child. Every sandboxed child — agent shell exec, skill scripts, MCP stdio servers — could therefore read `~/.lango/lango.json` (encrypted config + secret tokens), `~/.lango/lango.db` (session and audit database), `~/.lango/skills/` (executable skill files), and `~/.lango/workflows/` (state). The leak was readable secrets and history rather than mutability, but it was still a leak.
2. `Sandbox.AllowedWritePaths` had been a config field since PR 0 but no code path actually read it. Users who set it saw no effect.
3. There was no controlled bypass mechanism for commands that simply don't work inside the sandbox (e.g. `git status` cannot read `.git` once it's denied; `docker run` needs the docker socket). Users either had to disable the sandbox globally or go without the tool.
4. Fail-open fallback was silent — `exec.Tool.applySandbox` logged a zap warning when the isolator failed to apply, but nothing reached stderr or audit, so the user could be running commands unsandboxed for hours without noticing.
5. The audit recorder had no event for sandbox decisions. There was no way to answer "which commands ran sandboxed in this session?" or "why did the last fall back?".

This change closes those five gaps in one PR so the operational state of the sandbox layer is honest end-to-end: the control-plane is unreachable to children, the user has a clearly-audited bypass path for the commands they trust, and every apply / skip / reject / exclude decision is recorded.

## Goals / Non-Goals

**Goals:**

- Mask the lango control-plane (`~/.lango`) from every sandboxed child process produced by `exec.Tool`, `skill.Executor`, and `mcp.ServerConnection`, with one consistent mechanism.
- Make `.git` denial part of the baseline `DefaultToolPolicy` so the wired path actually protects git metadata (it was `StrictToolPolicy`-only, but the wired path uses default).
- Wire `Sandbox.AllowedWritePaths` so it actually grants write access in the exec tool's policy.
- Provide a basename-based `ExcludedCommands` bypass with mandatory audit recording, scoped to the exec tool only (skill and MCP have no per-call bypass semantics).
- Surface fail-open fallback to the user via a one-shot stderr warning AND a per-decision audit row.
- Add `SandboxDecisionEvent` to the event bus, subscribe in audit recorder, publish from all three sandbox call sites.
- Render the most recent N=10 sandbox decisions in `lango sandbox status`, with optional `--session <prefix>` filtering and graceful degradation when the audit DB is unavailable.
- Add `os_sandbox_excluded_commands` to the TUI form.
- Update README, configuration docs, sandbox CLI docs, and the agent prompts (TOOL_USAGE.md, SAFETY.md) so the documented contract matches the wired contract.
- Single OpenSpec change, single PR, ship behind the existing sandbox feature flag (`sandbox.enabled`).

**Non-Goals (deferred to PR 5+):**

- Native Landlock+seccomp backend implementation. PR 5 work; status still says "planned".
- File-level deny via `--ro-bind /dev/null <file>`. bwrap supports it but it requires per-file plumbing in the args compiler.
- Symlink chain resolution before bwrap arg compile. A user-created symlink that escapes the workspace will currently be honored by bwrap.
- Glob/path semantics normalization between Linux bwrap and macOS Seatbelt subpath matching.
- Per-tool policy override (each tool builds its own policy from scratch). Today every exec invocation shares one supervisor-built `SandboxPolicy`.
- `--unshare-user` for bwrap. Distro user-namespace policies are a minefield (`unprivileged_userns_clone`, AppArmor profiles); we keep relying on bwrap's setuid binary path.
- Glob/regex matching for `ExcludedCommands`. Basename-only is conservative on purpose.
- A TUI menu badge indicating "fail-open fallback is currently active" — would require an audit DB live query on every menu render. Visibility is provided via stderr warn-once + status section instead; revisit if user feedback shows it's needed.

## Decisions

### D1: Slice support in NormalizePaths via a separate helper

`normalizePath(p *string, ...)` is the existing single-pointer contract. Adding slice support inline would change its signature and ripple through the seven existing call sites. Instead, add a tiny `normalizePathSlice([]string, dataRoot, home) []string` helper that allocates a new slice and applies the existing helper to each entry. Empty entries are preserved so callers can distinguish "explicitly empty" from absent. The empty `WorkspacePath` case is preserved end-to-end: supervisor falls back to `os.Getwd()` only when `WorkspacePath == ""`, so we MUST keep that string empty when the user did not set it.

**Alternative considered**: collapse `normalizePath` to take `any` (`*string` or `[]string`) via type switch. Rejected — it makes the call sites less readable and complicates testing.

### D2: One control-plane mask, three policy helpers, one signature shape

The cleanest model is "sandboxed child cannot see lango's data directory at all". Implementation: append `cfg.DataRoot` to `Filesystem.DenyPaths`. On bwrap this becomes `--tmpfs <path>` which masks the directory with an empty tmpfs. On macOS Seatbelt it becomes `(deny file* (subpath "<path>"))`. Both block read AND write.

To enforce this consistently, all three policy helpers gain a `dataRoot string` parameter:

| Before | After |
|--------|-------|
| `DefaultToolPolicy(workDir string) Policy` | `DefaultToolPolicy(workDir, dataRoot string) Policy` |
| `StrictToolPolicy(workDir string) Policy` | `StrictToolPolicy(workDir, dataRoot string) Policy` |
| `MCPServerPolicy() Policy` | `MCPServerPolicy(dataRoot string) Policy` |

Empty `dataRoot` is allowed and skips the deny — this is the path used by isolated unit tests (`sandbox/os/policy_test.go`) so a test does not need to fabricate a real directory just to satisfy bwrap's "deny path must exist" check.

`StrictToolPolicy` becomes a thin wrapper around `DefaultToolPolicy` since the only previous strict-only feature (`.git` denial) is now baseline. The function is preserved as a separate symbol so future strict-only options can branch later without another signature migration.

`.git` denial is moved into the baseline `DefaultToolPolicy` because the actually-wired path uses default (supervisor.go calls `DefaultToolPolicy`, not strict). Keeping `.git` strict-only meant it was never actually applied.

**Alternative considered**: a `SandboxBundle` struct that all three call sites share. Rejected for PR 4 — it would require touching every wiring entry to thread the bundle, and we already have to touch the wiring for the bus parameter. Revisit in PR 5+.

### D3: Wire `AllowedWritePaths` in the supervisor only

`supervisor.go` builds the exec tool's policy. After calling `DefaultToolPolicy`, it appends `cfg.Sandbox.AllowedWritePaths` to `policy.Filesystem.WritePaths`. Skill and MCP do NOT consume `AllowedWritePaths` because their policies are application-specific (skill has its own workspace path, MCP only needs `/tmp`). If a user adds an entry that falls under `cfg.DataRoot`, the deny path wins (bwrap's last-mount-wins semantics, verified by a new test in `bwrap_args_test.go`).

### D4: ExcludedCommands matches the user command, not `cmd.Args[0]`

⚠️ **Critical implementation detail**: every `exec.Tool` execution path (`Run` line 133, `RunWithPTY` line 182, `StartBackground` line 245) wraps the user command in `exec.CommandContext(ctx, "sh", "-c", resolved)`. So `cmd.Args = ["sh", "-c", "<command>"]` and `filepath.Base(cmd.Args[0])` is **always `"sh"`**. If `applySandbox` matches on `cmd.Args[0]`, `ExcludedCommands: ["git"]` will never trigger and `ExcludedCommands: ["sh"]` will silently bypass everything.

The fix: change `applySandbox` signature to accept the raw user command string as a third parameter. Each caller passes the pre-`sh -c`, pre-secret-resolution `command` string. The matcher splits on whitespace, takes the first token, and compares its basename:

```go
func excludedMatch(userCommand string, patterns []string) (matched, pattern string) {
    if len(patterns) == 0 { return "", "" }
    fields := strings.Fields(userCommand)
    if len(fields) == 0 { return "", "" }
    base := filepath.Base(fields[0])
    for _, p := range patterns {
        if base == p { return base, p }
    }
    return "", ""
}
```

Examples:
- `git status` → first token `git` → basename `git` → match
- `/usr/bin/git push` → first token `/usr/bin/git` → basename `git` → match
- `cd /tmp && git status` → first token `cd` → no match (sandbox stays applied — safe direction)
- `git status | grep foo` → first token `git` → basename `git` → match

A regression test (`TestApplySandbox_ExcludedDoesNotMatchSh`) pins this semantic: `ExcludedCommands: ["sh"]` must NOT bypass an `echo hello` call. If a future refactor accidentally falls back to `cmd.Args[0]`, this test will fail loudly.

**Alternative considered**: glob/regex matching. Rejected — explosion radius is too large for a security carveout.

**Scope decision**: ExcludedCommands is wired in `exec.Tool` only. Skill executor takes its instructions from a skill definition (the user's intent is encoded there) and MCP is per-server-startup, not per-call, so neither has meaningful "per-command bypass" semantics.

### D5: SessionKey is derived from runtime ctx, not stored on the call site

The plan's first draft stored `SessionKey` as a field on `exec.Tool`/`skill.Executor`/`mcp.ServerConnection` and added `SetSessionKey` setters. This was wrong because the wiring runs at app startup before any session exists, and the lango codebase already has a clean ctx-based pattern: `internal/session/context.go` provides `WithSessionKey`/`SessionKeyFromContext`, used by `internal/tools/exec/policy.go:406`, `internal/adk/context_model.go:174`, `internal/automation/runner.go:29`, `internal/toolchain/mw_*.go`, and many more.

So `applySandbox(ctx, cmd, userCommand)` and `executeScript(ctx, ...)` derive the session key on-demand. `mcp.ServerConnection.createTransport()` is called at process startup with no session — its `SandboxDecisionEvent.SessionKey` is intentionally empty, and `audit/recorder.go.handleSandboxDecision` conditionally calls `SetSessionKey` so the `actor.NotEmpty()` validator on the ent schema does not fail.

### D6: One canonical SandboxDecisionEvent schema

```go
type SandboxDecisionEvent struct {
    SessionKey string    // session.SessionKeyFromContext(ctx); empty for MCP startup
    Source     string    // "exec" | "skill" | "mcp"
    Command    string    // user command, skill name, or MCP server name
    Decision   string    // "applied" | "skipped" | "rejected" | "excluded"
    Backend    string    // "bwrap" | "seatbelt" | "noop" | ""
    Reason     string    // empty for "applied", populated otherwise
    Pattern    string    // populated for "excluded" only
    Timestamp  time.Time // PublishSandboxDecision sets if zero
}
```

The four `Decision` values map onto the only four outcomes the sandbox layer can produce: the isolator was applied successfully (`applied`), the isolator could not be applied and we proceeded anyway because fail-closed=false (`skipped`), the isolator could not be applied and fail-closed=true rejected the call (`rejected`), or the user listed the command in `ExcludedCommands` (`excluded`). Audit recorder writes one row per decision with `actor=Source`, `target=Command`, and `details={decision, source, backend, reason, pattern}`.

A small `PublishSandboxDecision` helper handles the "bus may be nil" check and timestamp default so each call site is one line.

### D7: All three sandbox call sites publish — wiring complexity is paid up-front

Past sandbox bugs tended to fix one call site and miss the other two. To avoid that, every site that calls `OSIsolator.Apply` MUST also publish `SandboxDecisionEvent`. Stage 3 of the implementation includes a `Grep` inventory at start AND end to verify the count is consistent.

The three sites are: `internal/tools/exec/exec.go:applySandbox`, `internal/skill/executor.go:executeScript`, and `internal/mcp/connection.go:createTransport`. Each one gets a `bus *eventbus.Bus` field (or `Config.Bus` for exec) plus a `SetEventBus` setter. The wiring entry points are: `app.go:populateAppFields` callsite for exec via `Supervisor.SetEventBus(bus)`, `wiring_knowledge.go:initSkills` for skill via `Registry.SetEventBus(bus)` pass-through, and `wiring_mcp.go:initMCP` for MCP via `ServerManager.SetEventBus(bus)`.

`initSkills` and `initMCP` gain a `bus *eventbus.Bus` parameter (caller passes `m.bus` from the relevant module struct).

### D8: `lango sandbox status` Recent Decisions — graceful degradation by default

`lango sandbox status` is a standalone CLI: it loads config and probes the sandbox layer without booting the supervisor. The plan's first draft tried to use a "current session" filter, but standalone CLI has no current session. The fix:

- The default view is **global last 10** decisions across all sessions.
- Optional `--session <prefix>` flag filters by session-key prefix.
- Output shows an 8-character session-prefix in brackets so cross-session views are still readable; empty session keys (MCP startup) render as `--------`.

The audit DB is opened via the existing `cliboot.BootResult` helper, which is the same path the other CLI subcommands (learning history, etc.) use. `NewSandboxCmd` gains an optional `BootLoader` parameter; passing `nil` (or having the loader fail) silently skips the section so the diagnostic remains usable as a pure sandbox-layer inspection tool when the DB is locked, signed-out, or missing.

The DB client returned by the loader is **not** closed in the helper — it is owned by the bootstrap result and the cobra root is responsible for the process lifecycle. Closing it inside `renderRecentDecisions` would break subsequent commands sharing the same boot.

### D9: Documentation downstream sync ships in the same PR

Per CLAUDE.md, any change to `internal/` must update README/docs/TUI/prompts in the same response. PR 4 updates:

- `README.md`: feature list line + sandbox config table (new rows for `workspacePath`, `allowedWritePaths`, `excludedCommands`; new note on fail-open warning visibility).
- `docs/configuration.md`: top-of-section paragraph on control-plane masking + fail-open visibility, JSON example with `excludedCommands`, table rows for the new fields.
- `docs/cli/sandbox.md`: Recent Sandbox Decisions section with example output, ExcludedCommands semantics paragraph, fail-open warning paragraph.
- `prompts/TOOL_USAGE.md`: agent-facing "OS sandbox awareness" bullet under Exec Tool, including the explicit "control-plane is denied" rule and "do not invent shell tricks to bypass" rule.
- `prompts/SAFETY.md`: new "Control-plane is off-limits" bullet that names config / session DB / secret tokens / skills as denied surfaces.

### D10: One OpenSpec change, one new capability

`sandbox-exception-policy` is the only new capability. Its requirements cover ExcludedCommands semantics, the SandboxDecisionEvent schema, the four publish-site contract, the fail-open user-visible warning rule, and the actor mapping in audit recorder. The other deltas (Policy signatures, control-plane mask, status Recent Decisions, MCP wiring) modify existing capabilities (`os-sandbox-core`, `os-sandbox-cli`, `os-sandbox-integration`, `mcp-integration`).

### D11: Ship behind the existing flag, no migration drama

`sandbox.enabled=false` is still the default. Users who have not turned on the sandbox see no behavior change — same dead code path. Users who have turned it on get the control-plane mask immediately, plus a one-shot stderr warning if their config falls into fail-open and the isolator can't apply. The ent schema migration is an enum addition only (forward compatible).

## Risks / Trade-offs

- **`.git` denial in baseline default may break agent-driven git workflows.** Today the wired default policy denies the workspace's `.git` directory, which means `git status`, `git log`, etc. fail with "not a git repository" inside the sandbox. → Mitigation: the user can add `git` to `sandbox.excludedCommands`. The design change to make ExcludedCommands easy and auditable was specifically motivated by this. We considered making `.git` read-only-allow instead of fully denied, but bwrap's mount model makes it expensive to express "read-only inside an otherwise-writable parent" without also writing fragile per-platform bind sequences.
- **`/tmp` writability is required, so `TMPDIR=~/.lango/tmp` would self-conflict.** If a user sets `TMPDIR` to a path under `~/.lango`, the skill executor's `os.CreateTemp("", ...)` would create the script in a denied location and the sandboxed child could not read it. → Mitigation: documented as a gotcha in PR 5 risk; for now we assume `TMPDIR` defaults to `/tmp` which is in `WritePaths`. A follow-up could force-reset `TMPDIR=/tmp` in the policy environment.
- **MCP server binaries that live under `~/.lango` (e.g. bundled servers) would fail to launch.** → Mitigation: the manager could detect this and emit a clear validation error, but PR 4 keeps the policy simple. Documented as a known limitation.
- **Three publish sites create wiring drift risk.** → Mitigation: Stage 3 of the implementation runs `Grep "isolator.Apply"` at start AND end of the stage to verify every match has a matching publish call. The compile-time signature change on `applySandbox` (third parameter) also forces every caller to acknowledge the change.
- **`sync.Once` for the fallback warning is per-Tool, not per-process.** If a future build creates multiple `exec.Tool` instances, each one will print its own warning. → Mitigation: tool-local once is correct for now; promote to a package-level `sync.Once` only if user feedback shows noise. Test isolation is easier with tool-local state.
- **Stale spec drift from PR 1~3.** The active main specs (`os-sandbox-core`, `os-sandbox-integration`, `mcp-integration`) may still document the old function signatures. → Stage 6 of the implementation greps for `DefaultToolPolicy(workDir)` etc. in `openspec/specs/` and patches the deltas.

## Migration Plan

1. **Stage 1** (NormalizePaths): mechanical loader.go fix, fully covered by table-driven tests. Stop at commit boundary.
2. **Stage 2** (Policy signature + control-plane denylist + 3 sandbox sites): compiler-driven refactor — every call site updates because the signature changed. Tests for empty/non-empty dataRoot, deny-vs-write order.
3. **Stage 3** (ExcludedCommands + audit publish + 3 sites): largest stage. ent schema regen + eventbus event + recorder subscribe + 3 publish sites + 3 wiring entries. Stop at commit boundary.
4. **Stage 4** (status Recent Decisions + TUI): non-invasive additions. New unit tests for `truncateSessionKey` and graceful degradation.
5. **Stage 5** (docs): README/docs/prompts only. No code changes.
6. **Stage 6** (this document): OpenSpec change archived, deltas synced into main specs.

Each stage ends at a commit boundary (the user commits manually) and the next stage runs against the same branch.

**Rollback**: every stage is independently revertable via `git revert`. The ent schema enum addition is forward-compatible, so reverting Stage 3 does not corrupt existing rows (rows with `action="sandbox_decision"` simply become invalid for the validator and would need to be deleted manually if the Stage 3 commit is reverted).
