## Overview

The tracing initialization API remains unchanged. The implementation will
replace the hard-coded `os.Stderr` writer used by the stdout trace exporter
with an unexported package seam initialized to the same default writer.

## Design

- Add an unexported `io.Writer` variable initialized to `os.Stderr`.
- Use that writer when constructing the stdout trace exporter.
- Keep `InitTracer` semantics unchanged for disabled, stdout, none, and
  unsupported exporter modes.
- Do not add a new exported option until a caller needs to configure tracing
  output outside tests.

## Error Handling

Exporter construction errors and unsupported exporter errors keep their current
messages and wrapping behavior.

## Testing

Add a non-parallel test that temporarily replaces the writer seam with a buffer,
initializes stdout tracing, emits and ends a span, calls shutdown, and asserts
the buffer contains the emitted span name. The test is intentionally serialized
because it mutates a package-level seam and the global OpenTelemetry provider.
