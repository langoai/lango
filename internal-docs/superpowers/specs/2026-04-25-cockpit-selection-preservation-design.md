# Cockpit Selection Preservation Design

## Purpose / Scope

This design upgrades the landed cockpit dead-letter page so operators can keep looking at the same transaction across reload-triggering actions whenever that transaction still exists in the refreshed result set.

This slice adds:

- unified selection preservation

The target is the existing cockpit dead-letter page.

This slice directly includes:

- filter-apply selection preservation
- reset selection preservation
- retry-success refresh selection preservation
- deterministic fallback behavior

This slice does not directly include:

- empty-state transition messaging
- stale-detail banners
- per-action visual diff
- selection history

## Preservation Model

The rule is simple:

- when a reload happens
- if the current `selected_transaction_receipt_id` still exists in the new result set
- preserve that selection

This applies uniformly to:

- `Enter` apply
- `Ctrl+R` reset
- retry success refresh

The slice turns selection preservation into a unified dead-letter page reload rule rather than a one-off exception for a single action.

## Fallback Model

Selection preservation can fail when the selected transaction disappears from the refreshed result set.

Fallback rules:

- if the new result set is not empty:
  - select the first row
  - reload detail from that row
- if the new result set is empty:
  - clear selection
  - clear detail

This keeps the page deterministic and avoids ambiguous stale-detail states.

## Apply / Reset / Retry Refresh Semantics

After this slice, the page uses the same rule for all reload-triggering actions:

1. `Enter` apply
- preserve current selection if present
- otherwise first-row fallback

2. `Ctrl+R` reset
- preserve current selection if present after reset
- otherwise first-row fallback

3. retry success refresh
- preserve current selection if present after refresh
- otherwise first-row fallback

This keeps reload semantics consistent from the operator's point of view.

## Implementation Shape

Recommended implementation:

- extend `internal/cli/cockpit/pages/deadletters.go`
  - unify backlog reload helper around optional selected-ID preservation
  - preserve `selectedID` on:
    - apply
    - reset
    - retry success

- extend `internal/cli/cockpit/pages/deadletters_test.go`
  - cover preservation on apply
  - cover preservation on reset
  - cover preservation on retry success
  - cover fallback when the selected transaction disappears

No backend or bridge contract change is required.

## Follow-On Inputs

Natural follow-on work after this slice:

1. richer selection UX
- stale-selection messaging
- empty-state transition messaging

2. richer filter UX
- per-field clear
- result highlighting

3. higher-level operator surfaces
- additional CLI views
- broader cockpit workflow polish
