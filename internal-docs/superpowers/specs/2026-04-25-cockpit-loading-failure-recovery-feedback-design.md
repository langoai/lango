# Cockpit Loading / Failure Recovery Feedback Design

## Purpose / Scope

This design upgrades the landed cockpit retry UX so operators can better understand when a retry is actively running and why it failed.

The slice covers:

- retry in-flight loading state
- failure message detail

This slice directly includes:

- action-state rendering for retry
- input guarding while retry is running
- failure error string surfacing
- post-failure reset semantics

This slice does not directly include:

- richer success banner
- structured failure panel
- recent action history
- spinner widgets
- background polling after retry

## Loading Model

When retry execution starts, the detail pane `Retry action` line moves into a running state.

Minimum action states:

- `idle`
- `confirm`
- `running`

The first-slice rendering is text-first:

- `Retry action: running...`

No modal or separate progress panel is introduced.

## Failure Feedback Model

Failure feedback reuses the existing backend/meta-tool error string as-is.

The cockpit does not define a new structured error model in this slice.

The failure string is shown through the existing status-message path and may also be reflected in the detail-pane action state text if needed.

## Post-Failure Reset Semantics

After a retry failure:

- loading state is cleared
- action state returns to retryable idle
- failure status message remains visible

The UI does not enter a cooldown or sticky lock state.

## Interaction Guarding

While retry is running:

- replay-trigger input is blocked
- duplicate retry/confirm triggers are not allowed

This slice only guarantees action-path guarding. It does not attempt a larger global interaction lock.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - add retry action state
    - idle / confirm / running
  - render running state in the detail pane
  - block replay-trigger input while running
  - surface failure string through status messaging
- extend `internal/cli/cockpit/pages/deadletters_test.go`
  - cover running-state rendering
  - cover duplicate retry guard
  - cover failure reset behavior

The replay bridge and backend path stay unchanged.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer success feedback
- explicit success banner
- refreshed-state explanation

2. richer failure presentation
- structured failure section
- action history

3. broader operator polish
- spinner
- polling
- replay result pane
