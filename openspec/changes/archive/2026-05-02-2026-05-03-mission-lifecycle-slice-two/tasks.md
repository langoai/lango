# Tasks

- [x] Create the Mission Lifecycle Slice 2 OpenSpec change artifacts
- [x] Add the `mission-control-tui` delta for durable-first mission reads and durable mission creation paths
- [x] Add the `agent-control-plane-tools` delta for lightweight task tracking and execution-site mission linkage
- [x] Implement durable mission storage with `Mission`, `MissionStateHistory`, and `MissionExecutionLink`
- [x] Implement mission service write paths for direct mission start and proposal acceptance
- [x] Make Mission Control read durable mission rows first while retaining unmatched runtime overlays
- [x] Attach mission-aware execution linkage at mission-bound execution creation sites
- [x] Persist coarse `waiting_decision` mission state without introducing a durable approval queue
- [x] Verify the change with `openspec validate 2026-05-03-mission-lifecycle-slice-two --strict`
