## Design Summary

This workstream keeps the existing dead-letters cockpit page and its page-top summary strip, then extends it with:

- top 5 latest dead-letter reasons

Rules:

- aggregate from each row's current `LatestDeadLetterReason`
- show the result as a compact second `reasons:` line
- keep the existing placement
- recompute on:
  - initial load
  - filter apply
  - `Ctrl+R` reset
  - retry-success refresh

The slice does not add top actors, top dispatch references, a summary pane, or a summary page.
