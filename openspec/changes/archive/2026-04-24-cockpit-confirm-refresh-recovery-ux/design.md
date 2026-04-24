## Design Summary

This slice upgrades the cockpit retry action with:

- inline confirm
- success-path backlog/detail refresh

Behavior:

- first `r` enters confirm state
- second `r` invokes replay
- `Esc`, selection change, and filter apply clear confirm
- replay success refreshes backlog and selected detail
- replay failure only updates status feedback

The existing replay backend path remains unchanged.
