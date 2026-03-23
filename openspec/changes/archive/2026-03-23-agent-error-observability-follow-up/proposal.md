## Why

The current turn runtime can tell operators that a request ended in `tool_error`, `timeout`, or `loop_detected`, but it still cannot reliably explain the concrete root cause behind those buckets. A recent Telegram failure produced `E003` with a trace row but zero trace events, which means the system preserved neither a terminal error event nor an operator-facing cause summary.

This makes real-world debugging slower than it should be. The next step is not more retries or more user-visible detail, but deeper operator observability: keep user-facing messages broad, while making logs, durable traces, and `lango doctor` show the precise failure class and summary.

## What Changes

- Replace broad `classifyError(err) -> ErrorCode` behavior with a richer failure-classification model that preserves `cause_class`, `cause_detail`, and `operator_summary`.
- Extend durable turn traces so non-success turns always append a terminal failure event, even when the failure happens before the first normal runtime event.
- Add bounded trace-write deadlines so durable trace persistence cannot hang forever after parent cancellation.
- Surface detailed root-cause fields in channel/gateway logs and `lango doctor`, while keeping Telegram/user-facing messages conservative.
- Add typed incomplete child-summary notes so isolated specialist failures preserve cause information without promoting raw tool output to success.
- Add targeted replay fixtures and classification tests for pre-event failures and common E003 subtypes.

## Capabilities

### New Capabilities

### Modified Capabilities
- `agent-error-handling`: Add structured failure classification metadata and operator-facing diagnostic summaries.
- `agent-turn-tracing`: Add bounded trace writes, terminal failure events, and richer trace fields for root-cause analysis.
- `cli-health-check`: Expand doctor output to show recent failed multi-agent traces with cause metadata.
- `sub-session-isolation`: Preserve typed incomplete notes when visible assistant completion is missing.
- `test-infrastructure`: Add replay fixtures and table-driven tests for new failure classifications and terminal trace behavior.

## Impact

- `internal/adk/` error classification, agent error construction, and child-session summary behavior.
- `internal/turnrunner/` and `internal/turntrace/` runtime trace lifecycle, schema, and persistence.
- Channel/gateway logging and doctor output consumers.
- New ent schema fields and regenerated ent code.
- New regression fixtures/tests around pre-event failures and trace completeness.
