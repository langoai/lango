# Cockpit Confirm / Refresh Recovery UX Design

## Purpose / Scope

This design upgrades the landed cockpit `Retry` action with the first safety and refresh semantics needed for real operator use.

The slice covers:

- inline confirm before retry
- success-path backlog/detail refresh

The slice directly includes:

- first `r` enters confirm state
- second `r` performs the retry
- confirm reset triggers
- success-path UI refresh
- existing simple status feedback

The slice does not directly include:

- modal confirmation
- advanced loading indicators
- auto-timeout reset
- action history
- richer failure drill-down
- multiple recovery actions

## Confirm Model

The confirm interaction is inline and page-local.

Flow:

1. selected detail is retryable
2. user presses `r`
3. page enters confirm state
4. detail pane shows a confirm hint such as `press r again to confirm`
5. user presses `r` again
6. actual replay is invoked

No popup or modal is introduced in this slice.

## Reset Semantics

The confirm state is cleared when context changes.

Reset triggers:

- `Esc`
- row selection change
- filter change or apply

This avoids stale confirmation state surviving after the operator changes focus.

## Success Refresh Model

After a successful replay:

1. reload the backlog
2. reload the currently selected transaction detail

This slice intentionally keeps refresh semantics narrow:

- success path refreshes
- failure path does not auto-refresh

## Interaction / Feedback Model

Interaction summary:

- retryable detail selected
- first `r` enters confirm state
- second `r` invokes replay
- success:
  - show success status message
  - refresh backlog and detail
- failure:
  - show failure status message
  - keep current data in place

`Esc` cancels the confirm state.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add page-local confirm state
  - add first/second `r` handling
  - clear confirm on escape/selection/filter change
  - trigger backlog/detail refresh on success

- keep the existing replay bridge
  - still reuse `retry_post_adjudication_execution`

This is a page-local UX/state-machine upgrade, not a new backend feature.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer recovery feedback
- loading indicators
- failure detail presentation

2. timeout behavior
- auto-expiring confirm state

3. broader recovery surface
- more actions than just retry
- action history
