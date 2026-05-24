## Design Summary

This slice extends the cockpit dead-letter retry UX with:

- explicit retry action states:
  - `idle`
  - `confirm`
  - `running`
- retry-trigger guarding while replay is running
- failure error-string surfacing through the existing status-message path

The page keeps the existing interaction model:

- first `r` enters confirm state
- second `r` executes retry
- replay success refreshes backlog and selected detail
- replay failure leaves current data in place and returns the action to idle
