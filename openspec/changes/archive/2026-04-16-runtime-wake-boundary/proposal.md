## Why

The Anthropic Managed Agents architecture defines a clear separation between application-layer "resume" (opt-in handshake) and runtime-layer "wake" (harness re-initialization from event log). Lango has a solid `ResumeManager` (opt-in, `confirmResume + resumeRunId`) but no documented boundary for what a process-crash recovery path would require. Additionally, `bootstrap.Result.PhaseTiming` is already collected and logged per-phase but discarded after the process exits — there is no persistence or baseline comparison for diagnosing bootstrap regression. This change defines the wake boundary and surfaces PhaseTiming as a persisted diagnostic.

## What Changes

- **Design document**: Define what state must persist for a `wake(sessionID)` to be possible without full 5-phase bootstrap. Map the gap between existing `resume` (application-layer) and a future `wake` (runtime-layer).
- **PhaseTiming persistence**: Append `Result.PhaseTiming` to a JSONL file (`~/.lango/diagnostics/bootstrap-timing.jsonl`) after each successful bootstrap. Rotate at a fixed cap (const N=50).
- **BootstrapTimingCheck**: New `doctor` check using the `BootstrapAwareCheck` pattern. Current values from `boot.PhaseTiming`, baseline from the JSONL file. Compare per-phase durations against baseline median; Warn on significant regression, Skip if insufficient data.
- **doctor.go long description**: Update check count and list to include the new check.

## Capabilities

### New Capabilities
- `bootstrap-timing-diagnostics`: PhaseTiming file persistence, JSONL rotation, and BootstrapTimingCheck in doctor

### Modified Capabilities
- `run-ledger`: Delta spec adding design-level requirements for runtime wake boundary (no runtime behavior change — design only)

## Impact

- `internal/bootstrap/` — new file for JSONL writer, minor touch to pipeline.go or bootstrap.go to call writer after Execute
- `internal/cli/doctor/checks/` — new `bootstrap_timing.go`, register in `AllChecks()`
- `internal/cli/doctor/doctor.go` — long description update (check count + list)
- `openspec/specs/run-ledger/` — delta spec with design-level wake boundary requirements
