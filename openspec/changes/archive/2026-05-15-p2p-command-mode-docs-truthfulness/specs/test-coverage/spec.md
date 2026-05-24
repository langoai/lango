## ADDED Requirements
### Requirement: P2P feature mixed-command-mode guard stays executable
Repository-level regressions that reintroduce stale all-ephemeral wording into public P2P feature docs SHALL be enforced by an executable test.

#### Scenario: Stale all-ephemeral intro is rejected
- **WHEN** `team`, `workspace`, and `git` command families remain server-backed guidance surfaces
- **THEN** an executable repository test SHALL fail if the feature-page CLI intro describes the entire `lango p2p` surface as ephemeral-node execution
