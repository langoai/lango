## ADDED Requirements

### Requirement: README P2P parent-summary guard stays executable
Repository-level regressions that reintroduce the stale narrower `p2p/` summary into the README internal package tree SHALL be enforced by an executable test.

#### Scenario: Narrow legacy summary is rejected
- **WHEN** the README internal package tree still documents the shipped `internal/p2p` subtree
- **THEN** an executable repository test SHALL fail if the parent `p2p/` summary falls back to the older narrower wording
