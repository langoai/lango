## Why

The current post-adjudication status detail is receipts-centered and explains canonical state well, but operators still cannot see the latest raw async executor state without dropping down into the background substrate directly.

## What Changes

- add an optional thin `latest_background_task` bridge to `get_post_adjudication_execution_status`
- reuse the existing background task list/read path
- document the new detail-view bridge in public docs and main OpenSpec specs

## Impact

- better operator visibility into async execution state
- no canonical receipt mutation
- no list-surface expansion in this slice
