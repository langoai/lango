## 1. PhaseTiming JSONL Writer

- [x] 1.1 Create `internal/bootstrap/timing_log.go` with `AppendTimingLog(entries []PhaseTimingEntry, version string)` that writes a JSONL line to `~/.lango/diagnostics/bootstrap-timing.jsonl`
- [x] 1.2 Implement rotation: read existing file, drop oldest entries if > N (const 50), rewrite
- [x] 1.3 Handle corrupted lines gracefully: skip unparseable lines during read
- [x] 1.4 Ensure `os.MkdirAll` for diagnostics directory, file permissions 0644
- [x] 1.5 Call `AppendTimingLog` from `Pipeline.Execute` after successful completion (log-and-continue on error)
- [x] 1.6 Add unit tests: write, rotation at N+1, corrupted file recovery, write failure non-fatal

## 2. BootstrapTimingCheck

- [x] 2.1 Create `internal/cli/doctor/checks/bootstrap_timing.go` implementing `BootstrapAwareCheck`
- [x] 2.2 Current values from `boot.PhaseTiming`, baseline from JSONL file (exclude current run if already appended)
- [x] 2.3 Compare per-phase: Pass if ≤ 2x baseline median, Warn if above, Skip if < 3 records or missing file
- [x] 2.4 Register in `checks.go` `AllChecks()`
- [x] 2.5 Add unit tests: no-boot skip, no-baseline adaptive, fallback skip, median computation

## 3. Doctor long description

- [x] 3.1 Update `internal/cli/doctor/doctor.go` long description: add "Bootstrap Timing" to check list, increment count from 25 to 26

## 4. OpenSpec delta specs

- [x] 4.1 Verify `specs/bootstrap-timing-diagnostics/spec.md` covers all scenarios
- [x] 4.2 Verify `specs/run-ledger/spec.md` delta documents the wake boundary with no runtime changes

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./...` passes (bootstrap, doctor/checks)
- [x] 5.3 Manual: `lango doctor` shows Bootstrap Timing check (verified by user — regression detected in 2 phases as expected)
