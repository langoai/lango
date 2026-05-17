# Design: Harden Logging Default Writer Seam

## Approach

Add an unexported `defaultLogWriter io.Writer` seam in `internal/logging`, initialized to `os.Stderr`. `Init` will continue to prefer `LogConfig.Writer` and `LogConfig.OutputPath`; only the fallback branch will use `defaultLogWriter`.

## Test Strategy

Add a non-parallel test because `Init` mutates package-global logger state and the test temporarily replaces `defaultLogWriter`. The test will:

- Replace `defaultLogWriter` with a `bytes.Buffer`.
- Call `Init` without `Writer` or `OutputPath`.
- Emit a log entry through `Logger()`.
- Assert that the buffer receives the log entry.

## Compatibility

The public `LogConfig` API is unchanged. The default stream moves from stdout to stderr to align logs with terminal conventions and existing tracing behavior.
