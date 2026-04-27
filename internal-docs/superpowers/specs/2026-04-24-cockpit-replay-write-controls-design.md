# Cockpit Replay / Write Controls Design

## Purpose / Scope

This design adds the first operator recovery action to the landed cockpit dead-letter surface.

The slice covers:

- selected transaction detail pane
- a single `Retry` action
- reuse of the existing replay path

The slice directly includes:

- detail-pane `Retry` action
- key binding `r`
- success/failure status message
- reuse of `retry_post_adjudication_execution`

The slice does not directly include:

- confirm prompt
- auto refresh after replay
- action history
- multiple write actions
- background polling
- backlog-row write controls

## Action Model

The action model is intentionally simple.

1. operator selects a dead-lettered transaction
2. detail pane is visible
3. user presses `r`
4. cockpit invokes `retry_post_adjudication_execution`

The cockpit does not introduce a new recovery backend. It only exposes the already landed replay path through the operator UI.

## Enablement Rule

The `Retry` action is enabled only when:

- selected detail has `can_retry = true`

This keeps UI enablement aligned with the existing read-model navigation hint while leaving the final gate to the backend replay path.

## Interaction / Feedback Model

The first-slice UX remains minimal:

- key binding: `r`
- action location: selected detail pane
- result: success/failure status message only

This slice intentionally does not add:

- confirmation prompt
- auto refresh
- action history
- richer loading states

## Data / Control Reuse

The cockpit action reuses the existing control plane:

- `retry_post_adjudication_execution`

That means the following are inherited without duplication:

- dead-letter evidence gate
- canonical adjudication gate
- policy-driven replay controls
- append-only replay evidence behavior

The cockpit is only an operator-facing invocation surface.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add retry key binding handling
  - add transient status message state
  - dispatch the replay action through a cockpit bridge

- extend the cockpit bridge layer
  - call `retry_post_adjudication_execution`

- enable the action only when:
  - selected detail exists
  - `can_retry == true`

Expected flow:

1. select a row
2. load detail
3. if retryable, `r` becomes available
4. invoke replay
5. show success/failure message

## Follow-On Inputs

Natural follow-on work after this slice:

1. better recovery UX
- confirm prompt
- loading/disabled states
- auto refresh after success

2. more cockpit actions
- additional recovery or repair actions

3. richer operator workflow
- replay result panel
- action history
- bulk recovery
