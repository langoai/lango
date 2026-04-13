## ADDED Requirements

### Requirement: Tracing middleware
The toolchain MUST provide a `WithTracing(tracer)` middleware that wraps each tool invocation in an OpenTelemetry span. The span MUST record tool name, parameter count, and any error.

#### Scenario: Successful tool call traced
- **WHEN** a tool call succeeds
- **THEN** a span named `tool/<name>` SHALL be created with status OK

#### Scenario: Failed tool call traced
- **WHEN** a tool call returns an error
- **THEN** the span SHALL record the error and set status to Error

## MODIFIED Requirements

### Requirement: Middleware chain order
The production middleware chain MUST be: **Tracing** (outermost) → ExecPolicy → Approval → Principal → Hooks → OutputManager → Learning (innermost) → Handler. Tracing is outermost so that blocked calls are also traced.

#### Scenario: Blocked call produces span
- **WHEN** ExecPolicy blocks a tool call
- **THEN** the Tracing middleware SHALL still produce a span with the error recorded
