## Context

Bootstrap (`internal/bootstrap`) and application initialization (`internal/appinit`) run sequentially at startup. Currently neither subsystem records how long each phase or module takes, making it impossible to diagnose slow startups or track performance regressions.

Both subsystems already use structured logging via `logging.SubsystemSugar()`. The bootstrap pipeline uses a `State` struct that carries a `Result` between phases, and appinit's `Builder.Build()` returns a `BuildResult`.

## Goals / Non-Goals

**Goals:**
- Record wall-clock duration for every bootstrap phase and every appinit module initialization.
- Expose timing data in the existing `Result` / `BuildResult` structs so callers can aggregate or display it.
- Emit structured log lines with `duration_ms` for each phase and module.

**Non-Goals:**
- Changing how `app.go` consumes the results (out of scope for this unit).
- Adding CLI flags or TUI views for timing data.
- Sub-phase or sub-module granularity (e.g., timing individual DB queries within a phase).

## Decisions

1. **Timing types are simple structs, not interfaces.** Each entry holds a name (`string`) and a `time.Duration`. This is sufficient for logging and downstream aggregation. No need for an abstract timing interface at this stage.

2. **Timing is appended inside the existing loop, not via middleware.** The bootstrap `Execute()` already iterates over phases, and `Build()` already iterates over sorted modules. Wrapping `time.Now()` / `time.Since()` around each call is minimal, zero-allocation instrumentation that avoids adding a timing middleware layer.

3. **Timing is stored on the result struct, not on State directly.** For bootstrap, `PhaseTiming` is added to `Result` (which is returned to callers). The `Execute()` method writes timing into `state.Result.PhaseTiming`. For appinit, `ModuleTiming` is added to `BuildResult`.

4. **Logging uses existing subsystem sugar loggers.** Bootstrap pipeline adds a `logging.SubsystemSugar("bootstrap")` call. Appinit already has one.

## Risks / Trade-offs

- [Negligible overhead] `time.Now()` calls add ~20ns per phase/module. Acceptable for startup paths that take milliseconds or more.
- [Duration JSON encoding] `time.Duration` serializes as nanoseconds in JSON by default. This is acceptable for structured internal data; human-readable `duration_ms` is emitted via logs.
