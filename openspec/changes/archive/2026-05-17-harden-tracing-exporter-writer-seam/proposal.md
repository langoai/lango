## Why

OpenTelemetry stdout tracing currently binds the exporter writer directly to
process-global stderr. That makes span output harder to capture in embedded
runtimes and leaves the tracing boundary less testable than other stream-aware
subsystems.

## What Changes

- Route the stdout trace exporter through an explicit package writer seam.
- Add a regression test that emits a span and verifies it is written through the
  injected writer.
- Preserve the existing user-facing behavior: the default stdout trace exporter
  still writes to stderr.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `observability`: OpenTelemetry stdout tracing gains a seam-aware exporter
  writer guarantee.
- `test-coverage`: Standard tests gain executable coverage for the tracing
  exporter writer seam.

## Impact

- Affected code: `internal/observability/tracing.go`.
- Affected tests: `internal/observability/tracing_test.go`.
- Affected specs: `observability`, `test-coverage`.
- No CLI, config, public API, protocol, or dependency changes.
