## Design Summary

This workstream is documentation truth-alignment for the replay / recovery runtime that is already landed in code.

Landed runtime behavior to describe:

- post-adjudication follow-up resolves through one execution policy:
  - `auto_execute=true` => inline
  - `background_execute=true` => background
  - omitted flags => `manual_recovery`
- automatic retry is normalized into a background retry policy shape with bounded attempts and base delay
- retry and dead-letter evidence use the shared `post_adjudication_retry` source and subtype family
- operator replay uses that same recovery substrate and appends `manual-retry-requested` evidence before requeueing background work
- replay authorization remains config-backed and fail-closed, but now sits on top of the shared recovery gate

Documentation rules:

- treat current code and tests as the source of truth
- document additive landed behavior only
- remove follow-on items that are now implemented
- keep remaining-work lists focused on actual gaps: policy editing, broader substrate reuse, and broader dispute integration
