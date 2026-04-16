## Context

Lango's bootstrap is a 5-phase pipeline (`pipeline.go:88-115`): directory → DB/migration → passphrase/crypto → ent schema → optional PQ keys. `Result` carries `Config`, `DBClient`, `Crypto`, `IdentityKey`, etc. The pipeline already collects `PhaseTimingEntry{Phase, Duration}` per phase and logs `duration_ms`, but timing data is lost when the process exits.

`ResumeManager` (`runledger/resume.go`) is application-layer: user explicitly triggers `confirmResume + resumeRunId` via gateway. It replays RunLedger journal events but does NOT skip or lighten bootstrap. A process crash requires a full bootstrap regardless.

The `doctor` command already runs bootstrap once and supports `BootstrapAwareCheck` (`checks.go:82-85`) where checks receive `boot *bootstrap.Result` directly.

## Goals / Non-Goals

**Goals:**
- Define what state would need to persist for a lightweight `wake(sessionID)` path (design document only — no runtime implementation)
- Persist `PhaseTiming` to disk so successive runs can be compared
- Add a `BootstrapTimingCheck` to `doctor` that compares current timing to baseline

**Non-Goals:**
- Implementing the actual `wake` path (P2 — requires this design + data)
- Adding a config field for rotation cap N (fixed constant for now)
- Modifying `ResumeManager` or gateway handling

## Decisions

### D1: JSONL file for PhaseTiming persistence

**Choice**: Append one JSON line per bootstrap to `~/.lango/diagnostics/bootstrap-timing.jsonl`. Schema: `{"ts": "<RFC3339>", "version": "<version>", "phases": [{"name": "...", "durationMs": N}]}`. Rotate at const N=50 (read → drop oldest → rewrite).

**Why not SQLite?** Bootstrap timing is a diagnostic aid, not application state. A plain JSONL file avoids coupling to the DB lifecycle (which is itself a bootstrap phase). File corruption degrades gracefully (warn, don't crash).

**Why not provenance/checkpoint?** Provenance runs after full app init. The timing writer must fire immediately after `Pipeline.Execute` — before app wiring.

### D2: BootstrapAwareCheck pattern

**Choice**: New `BootstrapTimingCheck` implements `BootstrapAwareCheck`. Current values come from `boot.PhaseTiming` (already available). Baseline comes from reading the JSONL file (excluding the current run's entry if already appended). Compare each phase's current duration to baseline median. Pass if ≤ 2x median, Warn if above, Skip if < 3 baseline records.

### D3: Writer is fail-safe

**Choice**: JSONL writer errors are logged and swallowed. Bootstrap must never fail due to a diagnostic file write error. `os.MkdirAll` for the diagnostics directory; file permissions 0644.

### D4: Wake boundary is design-only

**Choice**: The delta spec for `run-ledger` adds design-level requirements describing what must persist for wake. No runtime behavior changes. This ensures the boundary is documented before any implementation work in P2.

**Wake boundary analysis** (for design.md / delta spec):
- **In-flight tool call state**: Currently lost on crash. Would need journal event for tool-start/tool-complete.
- **Pending approval state**: Gateway approval is synchronous per-turn. Lost on crash but re-requestable.
- **Supervisor handle / ADK session bridge**: Ephemeral. ADK sessions are in-memory. Would need session snapshot serialization.
- **What resume covers**: RunLedger journal replay (goal, steps, progress). Does NOT cover tool-level state, context window, or ADK session.
- **What wake would additionally need**: DB handle without full migration check (version pin), crypto provider re-init from cached envelope, ADK session reconstruction from journal.

## Risks / Trade-offs

- **[Risk] JSONL file grows on high-frequency bootstrap** → Capped at 50 entries; rewrite on each append keeps file small.
- **[Risk] Concurrent bootstrap instances** → File locking not implemented; accept last-writer-wins for diagnostics.
- **[Risk] Clock skew in timing comparison** → Comparing durations (not timestamps), so system clock drift is irrelevant.
