## Context

The OS sandbox package (`internal/sandbox/os/`) provides process isolation via macOS Seatbelt. Linux support was scaffolded with build-tag-gated files (`isolator_linux.go`, `landlock_stub.go`, `seccomp_stub.go`) but actual implementations were never created, causing `GOOS=linux go build` to fail. The probe path uses `ll.(*landlockIsolator).abiVersion` concrete type-cast, coupling probe logic to implementation types. Documentation across 13+ locations claims Linux enforcement exists.

## Goals / Non-Goals

**Goals:**
- Fix Linux cross-compilation (`GOOS=linux go build ./...` passes)
- Remove concrete type-cast coupling in probe path
- Add `Reason()` to `OSIsolator` for honest unavailability reporting
- Model "unknown/not probed" distinct from "unavailable" in `PlatformCapabilities`
- Prevent nil deref when sandbox is disabled (`disabledIsolator`)
- Fix all misleading documentation about Linux enforcement
- Enrich CLI/TUI status output with reasons and fail-mode explanation

**Non-Goals:**
- Actual Linux kernel isolation (Landlock/seccomp syscalls) — PR 3
- Backend registry with selection logic (auto/bwrap/native/none) — PR 2
- Policy compiler/normalization — PR 3
- Exception model (command exclusion, session bypass) — PR 4
- Real Landlock/seccomp kernel probes — PR 3

## Decisions

### D1: Separate "backend status" from "primitive probe"

**Decision:** `PlatformCapabilities` with standalone probe functions for kernel detection, `OSIsolator.Available()/Reason()` for runtime backend status.

**Why:** Merging them into one model (e.g., `BackendCapabilities.Available`) loses primitive-level visibility when using a single noop backend. CLI needs to show "Landlock: unknown (probe not yet implemented)" independently of "Isolator: noop".

**Alternative rejected:** Single `BackendCapabilities` struct — conflates two concerns, loses primitive granularity.

### D2: `Available()` single source of truth

**Decision:** `Reason()` added to `OSIsolator` interface. No `Available` field in `SandboxStatus` — always delegate to `Isolator.Available()`.

**Why:** Two sources (interface method + struct field) can diverge, breaking the "honest status" goal.

### D3: "unknown" vs "unavailable" probe states

**Decision:** `PlatformCapabilities` gains `LandlockReason`, `SeccompReason`, `SeatbeltReason` string fields. Linux probes return `false` with `"probe not yet implemented"` reason, displayed as "unknown" in CLI.

**Why:** Returning `false` with no context reads as "actually unavailable" which is dishonest on kernels that do support these primitives.

### D4: `disabledIsolator` for nil safety

**Decision:** New `disabledIsolator` type returned by `NewSandboxStatus()` when isolator is nil (sandbox disabled by config).

**Why:** Current `initOSSandbox()` returns nil when disabled. Any `SandboxStatus.Isolator.Available()` call would nil-deref without this guard.

### D5: Rewrite `isolator_linux.go` instead of adding stubs

**Decision:** Remove `NewLandlockIsolator()`/`NewSeccompIsolator()` calls entirely from `isolator_linux.go`. Use standalone probe functions and return `noopIsolator` directly.

**Why:** Adding `landlock_linux.go`/`seccomp_linux.go` stubs would fix compilation but preserve the concrete type-cast structural flaw. Rewriting eliminates the root cause.

### D6: Enforce fail-closed even when OSIsolator is nil

**Decision:** `exec.Tool.applySandbox()` checks `FailClosed` before early-returning on nil isolator. Skill and MCP wiring paths log warnings when fail-closed + unavailable.

**Why:** Codex review identified that supervisor wires `OSIsolator` only when `Available()=true`. When nil, `applySandbox()` silently skipped sandbox — breaking fail-closed guarantee. Skill/MCP paths lack a `FailClosed` field, so full enforcement requires a separate change; warning logs are added now.

### D7: Distinguish "unknown" from "unavailable" in capability display

**Decision:** `capabilityReasonStatus()` shows "unknown" only for reasons containing "not yet implemented". Definitive negatives (e.g., "sandbox-exec not found in PATH") show "unavailable (reason)".

**Why:** Codex review identified that all non-empty reasons were displayed as "unknown", which is misleading for definitive probe failures.

### D8: Hide fail-mode when sandbox is disabled

**Decision:** CLI status omits "Fail-Closed" line when `sandbox.enabled=false`.

**Why:** Showing "fail-closed (execution rejected)" when sandbox is disabled contradicts reality.

## Risks / Trade-offs

- **[Interface breaking change]** Adding `Reason()` to `OSIsolator` breaks 3 test mocks and 5 implementations → Mitigated: all internal, updated in same PR
- **[Dead code]** `landlock_stub.go`, `seccomp_stub.go`, `composite.go` become unused after rewrite → Accepted: kept for future backend implementations
- **[Incomplete Linux probe]** `probeLandlockKernel()`/`probeSeccompKernel()` return false stubs → Mitigated: "probe not yet implemented" reason prevents misinterpretation; real probes in PR 3
- **[Partial fail-closed on skill/MCP]** Skill executor and MCP manager lack `FailClosed` fields — warning logged but execution not blocked → Tracked for PR 2
