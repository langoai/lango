## Context

The current approval system has two persistence levels only:
- one-shot `Approve`, which applies to exactly one tool call
- `Always Allow`, which writes to the session-wide `GrantStore`

That leaves a gap for repeated identical calls within the same agent turn. When a browser action times out in approval or is denied, the model can issue the same call again and trigger another prompt. Since approval failures are returned as generic tool errors, the model has no strong signal that the failure is policy/user-driven rather than transient.

## Goals / Non-Goals

**Goals:**
- Prevent duplicate approval prompts for identical retries within the same request.
- Preserve the existing distinction between `Approve` and `Always Allow`.
- Make approval failures machine-distinguishable and easier to diagnose from logs.

**Non-Goals:**
- Persisting one-shot approvals across requests
- Changing approval UX/button layout
- Changing approval policy selection semantics

## Decisions

### D1: Add request-scoped turn-local approval cache

Each request gets an in-memory approval state attached to context. Entries are keyed by `tool name + canonical params JSON`.

Stored outcomes:
- approved once
- denied
- timed out
- unavailable

This state lives only for the current request and is discarded afterward.

### D2: `Approve` becomes a turn-local positive grant

A normal `Approve` will allow identical retries of the same `tool + params` within the same request without opening another approval prompt. `Always Allow` remains the only session-wide persistent grant.

### D3: Negative outcomes are replay-blocking within the same request

If an identical call has already resulted in deny/timeout/unavailable during the current request, the middleware will return the same structured outcome immediately without contacting the provider again.

### D4: Introduce approval sentinel errors with provider metadata

The approval package will expose sentinel errors for:
- denied
- timeout
- unavailable

Wrapped approval errors carry provider metadata so middleware logs can attribute the outcome.

### D5: Add explicit structured approval logs

Approval flow logs will include:
- request emission
- callback receipt
- granted / denied / expired
- turn-local bypass
- replay-blocked negative outcome

Common fields include session, request ID, tool, summary, params hash, grant scope, outcome, and provider.

## Risks / Trade-offs

- [Turn-local cache key mismatch] → Mitigation: use canonical JSON for params and include tool name.
- [Model still retries with slightly different params] → Mitigation: replay guard is exact-match only; prompt guidance covers the behavioral side.
- [Extra approval logs increase noise] → Mitigation: keep them at debug/info/warn according to outcome severity.

## Migration Plan

1. Add approval state + sentinel errors.
2. Update approval middleware and providers.
3. Add logs and tests.
4. Update prompts/specs.
5. Verify, sync, archive.
