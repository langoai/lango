## Context

Phase 2 makes RunLedger durable and write-through, but read paths still need to converge on the
ledger. Resume also remains a helper API rather than an integrated user flow.

## Goals

- Make RunLedger snapshots authoritative for all run reads.
- Inject active run summaries into agent command context.
- Provide explicit, opt-in resume orchestration with candidate selection.
- Keep pause/resume semantics compatible with turn-limit and escalation behavior.

## Non-Goals

- Full workspace isolation activation.
- Tool-profile narrowing in orchestrator runtime.

## Decisions

### 1. Reads converge on snapshots

Once authoritative-read mode is enabled, workflow/background/gateway must read run state from
RunLedger snapshots instead of local mirrors.

### 2. Resume remains opt-in

The system may detect candidates, but it must not silently resume. User intent and explicit
confirmation remain mandatory.

### 3. Command Context uses summaries, not full journals

Phase 3 injects compact run summaries into prompt assembly. Full journals remain on-demand to
avoid blowing the token budget.
