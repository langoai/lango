## Design Summary

This workstream is documentation truth-alignment for the already-landed operator-surface consolidation slice.

Landed behavior to describe:

- CLI dead-letter list parity now includes `--any-match-family`
- CLI summary now includes grouped `by_dispatch_family` buckets alongside raw top dispatch references
- CLI summary top sections are controlled by `--top`
- CLI summary trend output is controlled by `--trend-window` and `--trend-bucket`
- cockpit summary strip now includes `dispatch families:` and a compact trend line
- CLI retry now returns a structured follow-up snapshot and optional polling via `--wait`
- cockpit retry follow-up wording now explains backlog refresh and latest-status interpretation after acceptance

Documentation rules:

- treat code and tests as the source of truth
- describe additive behavior only; do not imply retry-substrate redesign
- narrow the remaining backlog to work that is still actually missing after this consolidation slice

This archive does not introduce new runtime or control-plane behavior. It records the completed docs-only alignment for the landed operator surface.
