# Tasks

- [x] Create the Mission Lifecycle Wave 2 OpenSpec change artifacts
- [ ] Add the `mission-control-tui` delta for durable-first mission reads and durable mission creation paths
- [ ] Add the `agent-control-plane-tools` delta for lightweight task tracking and execution-site mission linkage
- [ ] Implement durable mission storage with `Mission`, `MissionStateHistory`, and `MissionExecutionLink`
- [ ] Implement mission service write paths for direct mission start and proposal acceptance
- [ ] Make Mission Control read durable mission rows first while retaining unmatched runtime overlays
- [ ] Attach mission-aware execution linkage at mission-bound execution creation sites
- [ ] Persist coarse `waiting_decision` mission state without introducing a durable approval queue
- [ ] Verify the change with `openspec validate 2026-05-03-mission-lifecycle-wave-two --strict`
