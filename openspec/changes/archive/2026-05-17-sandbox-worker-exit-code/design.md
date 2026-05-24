## Context

The sandbox subprocess protocol launches `lango --sandbox-worker` and exchanges JSON over stdin/stdout. `SubprocessExecutor` interprets process exit failure as worker/process failure and interprets JSON `ExecutionResult.Error` as tool failure.

## Design

Introduce `RunWorkerWithIO(registry, in, out) int` in `internal/sandbox`. It reads the request from `in`, writes exactly one JSON `ExecutionResult` to `out`, and returns the intended process exit code. `RunWorker(registry)` becomes a compatibility wrapper that calls `RunWorkerWithIO` with `os.Stdin` and `os.Stdout`.

Change `cmd/lango` worker mode seam from `func()` to `func() int` so `runMain` returns the worker exit code. The real `main` function keeps the only actual process exit through `exitFn`.

Exit-code semantics remain unchanged: malformed requests and unregistered tools return code `1`; handler-level tool errors return code `0` with JSON error; successful tools return code `0` with JSON output.

## Testing

Add direct worker tests that use in-memory stdin/stdout and assert both returned exit codes and JSON results. Update `cmd/lango` worker-mode tests to prove the returned worker code is preserved and broker/root setup remains skipped.

## Risks

The main risk is changing how subprocess errors are classified. The tests should explicitly preserve the current distinction between worker failures (non-zero exit) and tool failures (zero exit with JSON error).
