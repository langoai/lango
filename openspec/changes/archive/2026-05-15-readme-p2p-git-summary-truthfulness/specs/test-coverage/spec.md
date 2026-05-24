## ADDED Requirements
### Requirement: README P2P git-summary guard stays executable
Repository-level regressions that reintroduce stale `lango p2p git push` or `lango p2p git fetch` summary wording into the README quick reference SHALL be enforced by an executable test.

#### Scenario: Stale README git summaries are rejected
- **WHEN** the current CLI still exposes guidance surfaces for `git push` and `git fetch`
- **THEN** an executable repository test SHALL fail if `README.md` describes those commands as direct live bundle actions
