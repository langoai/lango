## ADDED Requirements

### Requirement: README P2P package-subtree guard stays executable
Repository-level regressions that let the README internal package tree omit shipped `internal/p2p` subpackages SHALL be enforced by an executable test.

#### Scenario: P2P package subtree remains visible
- **WHEN** the repository still ships the current `internal/p2p` subpackages
- **THEN** an executable repository test SHALL fail if the README internal tree drops one of those package rows
