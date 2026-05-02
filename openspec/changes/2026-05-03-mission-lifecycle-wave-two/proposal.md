# Mission Lifecycle Wave 2

## Why

Mission Control Wave 1 made missions visible, but the visible rows are still projections over runtime state rather than durable product records. That leaves the first screen dependent on background-task and run overlays even when the user is looking for stable work units with durable identity.

Wave 2 exists to turn missions into first-class durable records without collapsing mission lifecycle into `RunLedger` or promoting every runtime artifact into mission truth.

## What Changes

This change introduces the first durable mission lifecycle slice for Mission Control:

- add hybrid storage with a durable mission latest-state row plus append-only mission state history
- give every durable mission its own `mission_id`, separate from execution IDs
- use `MissionExecutionLink` as the durable truth for mission-to-execution relationships
- start durable mission rows at `prepared`, not `proposed`
- keep `proposed` as a transient Mission Control overlay until the user accepts it
- make direct mission start and proposal acceptance real application write paths
- make Mission Control read durable mission rows first while still overlaying unmatched runtime work
- store `waiting_decision` as a coarse durable mission state instead of a durable approval queue
- keep `TaskEntry` lightweight and outside the durable mission checklist model in this wave
- attach mission-aware execution linkage at execution creation sites rather than retrofitting all task tracking

## Scope Guardrails

Wave 2 is intentionally bounded:

- no full durable mission event journal beyond latest row plus state history
- no automatic durable mission creation for every background task or run
- no durable multi-item approval queue
- no promotion of `TaskEntry` into the authoritative mission checklist model
- no rewrite of `RunLedger` as the top-level mission store

## User Impact

Mission Control becomes durable-first. Users can start a mission directly, accept a proposed mission into a real durable row, and come back to a stable mission record with its own lifecycle and linked executions. Runtime overlays still remain visible when work exists outside the durable mission graph, so Wave 2 improves truthfulness without hiding active unmatched work.
