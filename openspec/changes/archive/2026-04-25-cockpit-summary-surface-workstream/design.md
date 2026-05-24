## Design Summary

This workstream keeps the existing dead-letters cockpit page and adds:

- a compact page-top summary strip

Rules:

- derive the summary from the currently loaded backlog rows
- recompute on:
  - initial load
  - filter apply
  - `Ctrl+R` reset
  - retry-success refresh
- show only:
  - total dead letters
  - retryable count
  - `release/refund`
  - `retry/manual-retry/dead-letter`

The slice does not add top reasons, top actors, top dispatch references, or a separate summary pane/page.
