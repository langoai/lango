## 1. Bootstrap Timing Types

- [x] 1.1 Add `PhaseTimingEntry` struct to `internal/bootstrap/bootstrap.go` with `Phase` (string) and `Duration` (time.Duration) fields, JSON-tagged
- [x] 1.2 Add `PhaseTiming []PhaseTimingEntry` field to `Result` struct with `json:"phaseTiming,omitempty"` tag

## 2. Bootstrap Pipeline Instrumentation

- [x] 2.1 Add timing instrumentation to `Pipeline.Execute()` in `internal/bootstrap/pipeline.go`: wrap each phase with `time.Now()` / `time.Since()`, append `PhaseTimingEntry`, emit structured log with `duration_ms`
- [x] 2.2 Store accumulated timing in `state.Result.PhaseTiming` before returning

## 3. Appinit Timing Types

- [x] 3.1 Add `ModuleTimingEntry` struct to `internal/appinit/builder.go` with `Module` (string) and `Duration` (time.Duration) fields, JSON-tagged
- [x] 3.2 Add `ModuleTiming []ModuleTimingEntry` field to `BuildResult` struct with `json:"moduleTiming,omitempty"` tag

## 4. Appinit Builder Instrumentation

- [x] 4.1 Add timing instrumentation to `Builder.Build()` in `internal/appinit/builder.go`: wrap each module `Init()` with `time.Now()` / `time.Since()`, append `ModuleTimingEntry`, update log line with `duration_ms`
- [x] 4.2 Include accumulated `ModuleTiming` in the returned `BuildResult`

## 5. Verification

- [x] 5.1 Run `go build ./internal/bootstrap/...` and `go build ./internal/appinit/...` -- verify no compile errors
- [x] 5.2 Run `go test ./internal/bootstrap/...` and `go test ./internal/appinit/...` -- verify all tests pass
