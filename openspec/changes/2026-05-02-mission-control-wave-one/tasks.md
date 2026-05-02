# Tasks

- [x] Create the Mission Control Wave 1 OpenSpec change artifacts
- [ ] Add the `mission-control-tui` capability spec for the new default surface
- [ ] Update `cockpit-shell` for default Mission Control routing and cockpit-lifetime shared state ownership
- [ ] Update `cockpit-pages` for Mission Control page registration and routing with minimal shortcut churn
- [ ] Update `interactive-tui-chat` so bare `lango` no longer claims direct chat and cockpit chat reuses the shared pending approval owner
- [ ] Update `tui-approval-tiers` so active approvals resolve through the shared pending response path
- [ ] Expose optional runtime readers needed for honest projection into cockpit deps
- [ ] Add cockpit-lifetime buffers and subscriptions for pending approvals, activity, and learning suggestions
- [ ] Implement the Mission Control projector and page using deterministic Wave 1 rules only
- [ ] Verify loading, empty, degraded, narrow-terminal, and approval-routing behavior with focused tests
- [ ] Run `openspec validate 2026-05-02-mission-control-wave-one --strict`
