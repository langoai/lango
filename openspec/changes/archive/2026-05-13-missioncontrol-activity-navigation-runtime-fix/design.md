## Context

Mission Control uses the composer focus as the shared focus state for both the activity lane header and the inline composer. The help bar already treats that focus state as navigable when multiple activity rows exist, and the page already carries an `activityOffset` cursor, but the current key handling sends `↑/↓` into the composer before cursor movement can happen.

## Goals / Non-Goals

**Goals:**

- Make activity-lane `↑/↓` handling match the existing help surface.
- Keep the change local to Mission Control key routing and regression coverage.
- Preserve existing proposal, decision, and composer submit behavior.

**Non-Goals:**

- Redesign Mission Control focus semantics.
- Add new global shortcuts or change sidebar/content focus behavior.
- Refactor the composer widget itself.

## Decisions

- Prioritize activity navigation keys before forwarding keys to the composer when the composer/activity lane is focused and another activity row exists.
  - Rationale: the page already exposes `activityOffset`, renders help for this path, and documents row navigation for the focused lane.
  - Alternative considered: hide `↑/↓` help on the composer/activity lane. Rejected because it would preserve a dead internal cursor path instead of making the existing surface truthful.

- Add a focused regression that asserts `activityOffset` changes under `↓`.
  - Rationale: this directly proves the key-routing bug is fixed without depending on brittle full-view snapshots.

## Risks / Trade-offs

- [Composer arrow keys no longer reach the inline composer while activity navigation is actionable] → Limit interception to the composer/activity focus state and only the explicit `↑/k` and `↓/j` navigation bindings.
- [Future composer behavior changes could reintroduce routing drift] → Keep a direct Mission Control regression over `activityOffset`.
