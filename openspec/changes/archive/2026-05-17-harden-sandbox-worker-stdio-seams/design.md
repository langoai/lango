# Design

## Approach

Introduce two unexported seams in `internal/sandbox`:

- `workerStdin io.Reader = os.Stdin`
- `workerStdout io.Writer = os.Stdout`

`RunWorker` will delegate to `RunWorkerWithIO(registry, workerStdin, workerStdout)`. This preserves production defaults while making the public wrapper directly testable without replacing process-global stdio.

`RunWorkerWithIO` remains the single protocol implementation and continues to own JSON request decoding, handler dispatch, JSON result encoding, and exit-code behavior.

## Test Strategy

Add a non-parallel regression test that temporarily replaces the worker seams, calls `RunWorker`, and decodes exactly one `ExecutionResult` from the injected stdout buffer. The helper restores package globals with `t.Cleanup`.

The test must not call `t.Parallel()` because seam overrides are package-global state.

## Non-Goals

- Do not change the worker JSON protocol.
- Do not change subprocess launch behavior.
- Do not add exported test hooks.
