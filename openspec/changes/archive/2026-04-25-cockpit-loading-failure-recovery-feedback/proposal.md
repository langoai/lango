## Why

The cockpit dead-letter page already supports retry, inline confirm, and success refresh, but operators still need clearer action-state feedback while replay is running and better immediate failure visibility.

## What Changes

- surface retry `running...` state in the dead-letter detail pane
- guard duplicate retry triggers while replay is in flight
- surface backend failure strings through the existing status-message path
- document the landed loading/failure-feedback slice in public docs and main OpenSpec specs

## Impact

- clearer operator feedback during replay
- lower accidental duplicate retry risk
- no new backend endpoints or replay contracts
