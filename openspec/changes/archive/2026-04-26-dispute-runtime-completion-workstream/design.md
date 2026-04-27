## Design Summary

This workstream is documentation-only. It does not change Go behavior.

Key design points:

- treat receipts as the canonical dispute-runtime source of truth
- describe dispute hold as setting `dispute_lifecycle_status = hold-active`
- describe exhausted post-adjudication retries as preserving adjudication while re-escalating settlement progression to `dispute-ready` with `dispute_lifecycle_status = re-escalated`
- describe settlement progression as the canonical transaction-level state machine for renewed disagreement, including re-entry from `review-needed`, `approved-for-settlement`, and `partially-settled`
- describe dispute-linked tool receipts as surfacing `dispute_lifecycle_status` where the runtime now returns it
- update follow-on work so the docs stop listing already-landed dispute-runtime slices as future work
