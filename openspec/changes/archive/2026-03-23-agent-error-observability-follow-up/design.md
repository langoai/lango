## Context

The first reliability overhaul added `TurnRunner`, durable trace rows, and outcome classification, but the current system still loses the operator-facing reason behind failures:

- non-success channel turns are logged with `outcome` and `trace_id`, but not `cause_class` or a diagnostic summary
- `AgentError.UserMessage()` is reused as the durable trace summary, so user-safe phrasing replaces operator-meaningful detail
- a turn can still fail before the first observable runtime event, leaving a non-success trace with zero events
- trace persistence is intentionally detached from the parent context, but currently has no independent deadline
- `E003` is still too broad because `classifyError()` remains mostly heuristic

This change is intentionally scoped to observability. It does not add retries or new user-facing error detail. It makes the existing runtime explain itself.

## Goals / Non-Goals

**Goals:**
- Preserve a root-cause classification for every non-success turn.
- Ensure non-success traces always include at least one terminal failure event.
- Keep user-facing messages broad while exposing richer diagnostic detail to logs, traces, and doctor output.
- Bound trace persistence work so observability cannot hang the runtime.
- Preserve typed incomplete child-summary notes without turning raw tool output into a success summary.

**Non-Goals:**
- Provider retry or retry policy changes.
- Richer Telegram/WebSocket user-facing error messages.
- TUI diagnostics work.
- General performance work unrelated to operator observability.

## Decisions

### D1. Replace tuple-style classification with `FailureClassification`

The new `FailureClassification` struct will carry:
- `Code`
- `CauseClass`
- `CauseDetail`
- `OperatorSummary`

`AgentError` will embed these fields so the same classification travels through logs, traces, and doctor output.

**Rationale:** Tuple-style `(ErrorCode, string)` scales poorly once the system needs detail, summary, and retryability decisions. A struct is easier to evolve.

### D2. Distinguish operator diagnostics from user messaging

`UserMessage()` remains conservative and channel-safe. A separate operator-facing summary helper will provide the real diagnostic payload for logs and traces.

**Rationale:** User messaging and operator debugging have different requirements. Mixing them caused the trace summary to become too vague.

### D3. Require `terminal_error` for all non-success traces

`TurnRunner` will append a `terminal_error` event before finishing the trace whenever the outcome is not success. This includes pre-event failures.

**Rationale:** A non-success trace with zero events is effectively non-debuggable.

### D4. Use bounded detached context for trace writes

Trace persistence will use `context.WithTimeout(context.WithoutCancel(parent), 5*time.Second)`.

**Rationale:** Trace writes must survive parent cancellation long enough to persist, but they must never hang indefinitely.

### D5. Keep trace payload JSON shape stable

`payload_json` remains a single JSON string. Truncation is expressed with a separate `payload_truncated` boolean field on the trace event row.

**Rationale:** Wrapping payloads changes consumer expectations. A dedicated boolean keeps storage shape stable.

### D6. Fix lifecycle ordering before downstream callbacks

The runner order becomes:
1. classify result
2. append terminal error event if needed
3. finish trace
4. fire turn callbacks

**Rationale:** Downstream consumers should never observe a still-running trace for a turn that has already completed logically.

## Risks / Trade-offs

- [More fields in trace schema] → Mitigation: use optional/nillable columns for backward compatibility.
- [Cause detail may capture sensitive strings] → Mitigation: bound and sanitize stored details before persistence.
- [Heuristic branches still exist after refactor] → Mitigation: sentinel and runtime-guard branches run first, reducing reliance on heuristics.
- [Doctor output may become noisy] → Mitigation: limit recent failure display to the latest three traces.

## Migration Plan

1. Add schema fields and regenerate ent code.
2. Introduce `FailureClassification` and plumb it through `AgentError`.
3. Update `TurnRunner` and trace recorder for bounded trace writes and `terminal_error` events.
4. Update channel/gateway logs and doctor output.
5. Add replay fixtures and classification tests.

## Open Questions

- None for v1. The timeout, payload shape, and visibility boundaries are fixed in this change.
