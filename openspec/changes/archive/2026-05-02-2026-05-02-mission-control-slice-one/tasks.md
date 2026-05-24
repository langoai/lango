# Tasks

- [x] Create the Mission Control Slice 1 OpenSpec change artifacts
- [x] Add the `mission-control-tui` capability spec for the new default surface
- [x] Update `cockpit-shell` for default Mission Control routing and cockpit-lifetime shared state ownership
- [x] Update `cockpit-pages` for Mission Control page registration and routing with minimal shortcut churn
- [x] Update `interactive-tui-chat` so bare `lango` no longer claims direct chat and cockpit chat reuses the shared pending approval owner
- [x] Update `tui-approval-tiers` so active approvals resolve through the shared pending response path
- [x] Expose optional runtime readers needed for honest projection into cockpit deps
- [x] Add cockpit-lifetime buffers and subscriptions for pending approvals, activity, and learning suggestions
- [x] Implement the Mission Control projector and page using deterministic Slice 1 rules only
- [x] Verify loading, empty, degraded, narrow-terminal, and approval-routing behavior with focused tests
- [x] Run `openspec validate 2026-05-02-mission-control-slice-one --strict`
