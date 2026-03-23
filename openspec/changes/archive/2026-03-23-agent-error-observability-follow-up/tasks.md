## 1. Failure Classification

- [x] 1.1 Introduce `FailureClassification` and extend `AgentError` with `CauseClass`, `CauseDetail`, and `OperatorSummary`.
- [x] 1.2 Refactor existing `classifyError` call sites to build `AgentError` from `FailureClassification`.
- [x] 1.3 Add initial cause-class coverage for approval, tool lookup/validation, provider, timeout, turn limit, repeated-call, and empty-after-tool-use cases.

## 2. Durable Turn Trace

- [x] 2.1 Extend `TurnTrace` and `TurnTraceEvent` schema with cause fields and `payload_truncated`.
- [x] 2.2 Update `TurnRunner` to use bounded detached trace-write contexts.
- [x] 2.3 Guarantee that every non-success turn appends a `terminal_error` event before finishing the trace.
- [x] 2.4 Preserve stable payload JSON shape and represent truncation through metadata only.

## 3. Downstream Observability

- [x] 3.1 Log non-success channel turns at `warn` with `error_code`, `cause_class`, `summary`, and `trace_id`.
- [x] 3.2 Add the same diagnostic fields to gateway logs and `agent.error` payloads.
- [x] 3.3 Expand `Multi-Agent` doctor output to include `trace_id`, `outcome`, `error_code`, `cause_class`, and `summary`.
- [x] 3.4 Replace generic child-summary placeholders with typed incomplete notes that preserve the cause.

## 4. Verification

- [x] 4.1 Add a pre-event failure replay fixture and assert that non-success traces never end with zero events.
- [x] 4.2 Add table-driven tests for the new failure classifications and terminal trace payloads.
- [x] 4.3 Run `go build ./...` and `go test ./...` and fix regressions.
