## ADDED Requirements
### Requirement: CLI index P2P team-summary guard stays executable
Repository-level regressions that reintroduce stale live-control wording for `lango p2p team` into the public CLI index SHALL be enforced by an executable test.

#### Scenario: Stale P2P team summaries are rejected
- **WHEN** `docs/cli/index.md` describes `lango p2p team list/status/disband` as direct live-control commands
- **THEN** an executable repository test SHALL fail
