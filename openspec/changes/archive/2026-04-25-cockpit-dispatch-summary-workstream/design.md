## Design Summary

This workstream keeps the existing dead-letters cockpit page and its page-top summary strip, then extends it with:

- top 5 latest dispatch references

Rules:

- aggregate from each row's current `LatestDispatchReference`
- show the result as a compact fourth `dispatch:` line
- keep the existing placement
- recompute on:
  - initial load
  - filter apply
  - `Ctrl+R` reset
  - retry-success refresh

The slice does not add a summary pane, a summary page, grouped dispatch families, or trend views.
