## Design Summary

This workstream keeps the existing dead-letters cockpit page and its page-top summary strip, then extends it with:

- top 5 latest manual replay actors

Rules:

- aggregate from each row's current `LatestManualReplayActor`
- show the result as a compact third `actors:` line
- keep the existing placement
- recompute on:
  - initial load
  - filter apply
  - `Ctrl+R` reset
  - retry-success refresh

The slice does not add top dispatch references, a summary pane, or a summary page.
