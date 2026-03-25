## Context

Structured multi-agent mode currently has two independent reliability gaps that compound each other. A streaming specialist failure does not discard the active isolated child session on the iterator-error path, and the structured recovery wrapper treats post-specialist tool failures as generic same-input retries. When those failures leave an unanswered tool call in parent-visible history, OpenAI request conversion injects orphan-repair tool responses, which makes the user-visible symptom look like a provider failure instead of a runtime cleanup failure. Long turns also lose trace append/finish events because the trace recorder reuses a single five-second detached context for the entire turn.

## Goals / Non-Goals

**Goals:**
- Make streaming failure cleanup consistent with collection-based cleanup for isolated specialists.
- Prevent failed specialist turns from poisoning retries or later turns with dangling tool-call state.
- Make structured recovery reroute away from the failed specialist when a specialist tool error already occurred.
- Preserve and improve diagnostics by recording recovery attempts and by allowing long turns to finish trace persistence.

**Non-Goals:**
- Rework orchestrator routing tables, keyword matching, or specialist prompt boundaries.
- Change provider-side orphan repair semantics beyond additional diagnostics.
- Introduce new user-facing CLI or config settings.

## Decisions

### 1. Fix cleanup in the streaming runtime, not in storage annotations
The first fix point is `RunStreamingDetailed`, because that path currently returns iterator errors without discarding the active child session. Storage-layer timeout annotations are a separate concern and do not explain the observed structured retry loop.

Alternative considered:
- Fix only `AnnotateTimeout`: rejected because the observed failures are not primarily timeout outcomes and the stale isolated session can still survive iterator-error exits.

### 2. Add explicit failed-turn cleanup for dangling parent-visible tool calls
The runtime will close dangling tool calls immediately after a failed turn by appending matching synthetic tool responses exactly once. This keeps parent-visible history valid before structured retry or later turns reuse it.

Alternative considered:
- Rely on provider-side orphan repair only: rejected because it defers cleanup too late and hides the runtime bug behind provider-specific behavior.

### 3. Make structured recovery specialist-aware
The coordinating executor already tracks the last delegation target. Recovery will now carry that target into `RecoveryContext` and prefer reroute hints for `ErrToolError` after a specialist delegation has occurred. Generic pre-specialist retries stay unchanged.

Alternative considered:
- Always reroute on any tool error: rejected because pre-delegation or provider-level tool failures still benefit from same-input retry.

### 4. Use per-write detached trace contexts
Trace persistence will create a fresh detached timeout context for each create/append/finish write. This preserves the bounded timeout contract without letting early writes exhaust the whole turn’s trace budget.

Alternative considered:
- Increase the shared timeout only: rejected because long turns can still consume the single budget before the final append/finish operations.

## Risks / Trade-offs

- [Risk] Failed-turn cleanup could append duplicate synthetic tool responses if it does not detect already-closed calls. → Mitigation: compute unanswered tool-call IDs from stored history and append only missing closures.
- [Risk] Specialist-aware reroute could hide a genuinely retryable tool failure if delegation detection is wrong. → Mitigation: only switch to reroute when a non-root specialist target was actually observed in the current attempt.
- [Risk] Per-write trace contexts may increase trace-store write pressure slightly. → Mitigation: keep the same timeout budget and payload truncation behavior; only context lifecycle changes.
- [Risk] Cleanup notes or synthetic closures could violate isolation if raw child history leaks into parent persistence. → Mitigation: keep discard/merge behavior unchanged and close dangling calls only from parent-visible history.

## Migration Plan

1. Update runtime cleanup and recovery behavior with regression tests.
2. Update trace recorder to use per-write detached contexts and emit recovery events.
3. Refresh multi-agent runtime docs to match the implemented behavior.
4. Validate with `go build ./...` and `go test ./...`.

Rollback:
- Revert the cleanup/recovery changes together if retries regress.
- Revert per-write trace context changes independently if trace persistence introduces unexpected store pressure.

## Open Questions

- None for implementation. The change is scoped to runtime cleanup, structured recovery, trace persistence, and documentation.
