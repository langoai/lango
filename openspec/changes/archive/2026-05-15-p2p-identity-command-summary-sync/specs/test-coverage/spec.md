## ADDED Requirements
### Requirement: P2P identity command-summary guard stays executable
Repository-level regressions that reintroduce stale `lango p2p identity` command-summary wording into the public P2P feature page SHALL be enforced by an executable test.

#### Scenario: Stale summary wording is rejected
- **WHEN** `docs/features/p2p-network.md` describes `lango p2p identity` as showing only peer identity and listen addresses
- **THEN** an executable repository test SHALL fail
