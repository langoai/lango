## ADDED Requirements

### Requirement: README agent inventory uses the current trace slash form
The README internal CLI inventory SHALL describe the agent diagnostics slice using the current slash-separated form `trace list/show/metrics/graph`.

#### Scenario: Stale hyphen shorthand stays removed
- **WHEN** a maintainer updates the README internal tree inventory
- **THEN** it SHALL describe the agent diagnostics slice as `trace list/show/metrics/graph` instead of `trace list-show-metrics/graph`
