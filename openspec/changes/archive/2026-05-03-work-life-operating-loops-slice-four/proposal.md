## Why

Mission Control can already show durable missions, proposals, and live decisions, but it still leaves users to infer the broader operational picture themselves. Slice 4 is needed to surface recurring work, unresolved follow-ups, scheduled activity, and manual-recovery loops in a deterministic operator-facing agenda without pretending unsupported calendar, inbox, or third-party task integrations already exist.

## What Changes

- add a loop-and-agenda projection over existing real sources instead of introducing a new durable loop table
- extend Mission Control so it can surface operator loops alongside the existing mission/proposal/decision surfaces
- use deterministic agenda ordering so the highest-attention loops are always ranked the same way for the same input state
- limit scheduled automation loops in the first slice to cron-job sources only
- generate follow-up loops only from explicit deterministic predicates over current mission/proposal/session state
- keep calendar, inbox, and external task-system integrations as explicit non-goals for this slice

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `mission-control-tui`: Mission Control gains deterministic loop and agenda projection while keeping durable missions as the primary owned work surface

## Impact

- Affected code: `internal/loopview`, `internal/app`, `internal/cli/cockpit`, `cmd/lango`
- Affected behavior: Mission Control can show loop and agenda rows built from durable missions, pending inquiries, dead-letter backlog, cron-based scheduled automation, and deterministic follow-up signals
- Explicit exclusions: no new calendar, inbox, or external task integration contracts in this change
