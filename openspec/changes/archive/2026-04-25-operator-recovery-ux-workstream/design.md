## Design Summary

This workstream keeps the existing retry path and improves only the operator-facing recovery UX.

Key design points:

- CLI retry output distinguishes:
  - precheck rejection
  - retry-request failure
  - retry-request acceptance
- CLI `json` output returns a clearer structured retry result
- cockpit retry state labels move through:
  - confirm
  - requesting
  - accepted / failed
- success wording explicitly means request acceptance, not completed execution

The workstream does not add polling, action history, or new recovery actions.
