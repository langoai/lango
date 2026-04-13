## Why

Bootstrap and module initialization latency is invisible -- there is no timing data to identify slow phases or modules. Adding instrumentation enables developers to diagnose startup performance bottlenecks and track regressions.

## What Changes

- Add `PhaseTimingEntry` type and `PhaseTiming` field to bootstrap `Result` so callers receive per-phase duration data.
- Instrument `Pipeline.Execute()` to record elapsed time for each bootstrap phase with structured log output.
- Add `ModuleTimingEntry` type and `ModuleTiming` field to appinit `BuildResult` so callers receive per-module duration data.
- Instrument `Builder.Build()` to record elapsed time for each module initialization with structured log output.

## Capabilities

### New Capabilities
- `startup-instrumentation`: Timing instrumentation for bootstrap phases and appinit module initialization, exposing duration data via result structs and structured logs.

### Modified Capabilities
- `bootstrap-pipeline`: Add phase timing collection to pipeline execution and expose timing in Result.
- `app-module-build`: Add module timing collection to builder and expose timing in BuildResult.

## Impact

- `internal/bootstrap/bootstrap.go` -- new `PhaseTimingEntry` type, new `PhaseTiming` field on `Result`.
- `internal/bootstrap/pipeline.go` -- timing instrumentation in `Execute()`, new `time` import.
- `internal/appinit/builder.go` -- new `ModuleTimingEntry` type, new `ModuleTiming` field on `BuildResult`, timing instrumentation in `Build()`.
- No CLI, TUI, or external API changes. Purely additive internal instrumentation.
