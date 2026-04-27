## Why

The cockpit dead-letter surface is already readable and filterable, but operators still need to leave the cockpit to trigger the first recovery action.

## What Changes

- add a detail-pane `Retry` action to the cockpit dead-letter page
- reuse the existing `retry_post_adjudication_execution` meta tool
- document the cockpit write-control slice in public docs and main OpenSpec specs

## Impact

- first operator recovery action inside cockpit
- no new recovery backend
- no confirm prompt or auto-refresh in this slice
