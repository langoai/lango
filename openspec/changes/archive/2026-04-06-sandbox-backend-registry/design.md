## Context

PR 1 (`linux-sandbox-hardening`) decoupled the probe path from concrete types and added `Reason()` to `OSIsolator`. PR 2 builds on that foundation by introducing a backend selection layer and fixing the inconsistent fail-closed enforcement across exec/skill/MCP paths.

The pre-PR-2 wiring had three independent code paths each calling `sandboxos.NewOSIsolator()` directly, with different fail-closed semantics:
- `supervisor.go` — passes isolator + FailClosed to exec.Tool, but only when `Available()=true`
- `wiring_knowledge.go` — passes isolator to skill registry only when `Available()=true`, no FailClosed
- `wiring_mcp.go` — passes isolator to MCP manager only when `Available()=true`, no FailClosed

This meant `failClosed=true` was honored only for exec tool. Skill scripts and MCP stdio servers ignored it entirely.

## Goals / Non-Goals

**Goals:**
- Single source of truth for backend selection (`SelectBackend` + `PlatformBackendCandidates`)
- Typed backend identity (no `Name()` string matching)
- Fail-closed enforcement in all three process-launching paths
- Honest CLI/TUI display of backend availability and opt-out state
- Validation of `sandbox.backend` at startup (no silent fallback)
- Treat `backend=none` as explicit opt-out distinct from "no backend available"

**Non-Goals:**
- Implementing actual `bwrap` or `native` backends (PR 3)
- Real Linux kernel probes for Landlock/seccomp (PR 3)
- Policy compiler / normalization (PR 3)
- Exception model / command exclusion (PR 4)
- Adding Backend Availability or seccomp profile fields to `PlatformCapabilities`

## Decisions

### D1: Typed `BackendCandidate` instead of `[]OSIsolator`

**Decision:** `SelectBackend(mode, []BackendCandidate)` where `BackendCandidate{Mode, Isolator}` provides explicit identity.

**Why:** Identifying backends by `Name()` string matching is brittle — typos and renames silently break selection. `Mode` enum is compiler-checked.

**Alternative rejected:** Pass `[]OSIsolator` and match on `Name()` — fragile.

### D2: `SetFailClosed(bool)` as additive method

**Decision:** Add separate `SetFailClosed(bool)` to `Executor`/`Registry`/`ServerConnection`/`ServerManager` instead of changing `SetOSIsolator` signature.

**Why:** Avoids touching `skill/registry.go`'s call site in unrelated paths and preserves write-set independence for parallel implementation. Each component can adopt fail-closed without coordinated signature changes.

### D3: `backend=none` as explicit opt-out (not "unavailable backend")

**Decision:** `wiring_sandbox.go` and `supervisor.go` short-circuit when `mode == BackendNone`, returning nil/skipping wiring. Fail-closed does not apply.

**Why:** The config doc explicitly states `none` "disables OS isolation even when sandbox.enabled is true". Treating it as a degenerate unavailable backend would force users running `enabled=true + failClosed=true + backend=none` to lose all execution — contradicting their explicit opt-out.

**Alternative rejected:** Make `backend=none` return a `disabledIsolator` variant that bypasses fail-closed. More complex than short-circuiting in wiring.

### D4: Explicit-mode selection preserves identity (no noop fallback)

**Decision:** When user selects `backend=bwrap` explicitly, `SelectBackend` returns the bwrap stub isolator AS-IS (even if `Available()=false`), not a noop. Only `auto` and `none` return noop.

**Why:** Explicit selection means the user expects feedback about THAT specific backend. If we silently fall back to noop, the CLI `resolved` line lies about which backend the runtime would attempt. With identity preserved, fail-closed correctly rejects execution and the reason field surfaces "bwrap backend not yet implemented".

### D5: Auto fallback aggregates candidate reasons

**Decision:** When `auto` finds no available backend, the noop fallback's `Reason()` is built from all candidate reasons (e.g., `"seatbelt: sandbox-exec not found in PATH; bwrap: bwrap backend not yet implemented"`).

**Why:** A generic "no available backend" string hides the actionable information users need to fix their setup. Aggregating per-candidate reasons gives a complete diagnostic picture.

### D6: Startup validation, not silent fallback

**Decision:** `config.Validate()` calls `ParseBackendMode()` and rejects unknown values with a startup error. `wiring_sandbox.go` / `supervisor.go` discard the parse error since validation has already run.

**Why:** Silent coercion of typos (e.g., `seatbeltt` → `auto`) lets users believe they configured one backend while another is running. Loud failure at startup matches the contract documented in the config field comment.

### D7: Shared `PlatformBackendCandidates()` helper

**Decision:** Both wiring (`initOSSandbox`, `supervisor.New`) and CLI (`sandbox status`, `sandbox test`) call `PlatformBackendCandidates()` from `registry.go` to build their candidate list.

**Why:** If wiring and CLI assemble candidates independently, they will eventually drift. A shared helper guarantees they always see the same set.

### D8: `lango sandbox test` honors configured backend

**Decision:** `newTestCmd()` accepts `cfgLoader` and uses `SelectBackend(mode, candidates)` instead of `NewOSIsolator()`.

**Why:** Otherwise users running `backend=none` or `backend=bwrap` see Seatbelt smoke test results, which contradict the runtime's actual behavior.

## Risks / Trade-offs

- **[Breaking config validation]** Existing configs with typos in `sandbox.backend` will fail to start → Mitigated: empty string defaults to `"auto"`, only non-empty unknown values fail; clear error message points to valid values
- **[Behavioral change for failClosed users]** Skills/MCP previously ran unsandboxed even with `failClosed=true` → This is the intended fix; users relying on the buggy behavior will discover it through clear `ErrSandboxRequired` errors
- **[Stub backends in public surface]** `bwrap` and `native` appear in TUI/CLI options before they work → Mitigated: stubs always report `Reason()="X backend not yet implemented"` and CLI status section explicitly labels them; `auto` mode skips them gracefully
- **[Dual code paths]** Wiring layer has both `if iso != nil` (skill/MCP) and explicit `if mode == BackendNone` (supervisor) checks → Accepted; supervisor builds its own exec config and cannot reuse `initOSSandbox` directly without a refactor
