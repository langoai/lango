## ADDED Requirements
### Requirement: P2P identity docs truthfulness guard stays executable
Repository-level regressions that reintroduce stale `lango p2p identity` wording into public P2P feature docs SHALL be enforced by an executable test.

#### Scenario: Stale DID-output wording is rejected
- **WHEN** `docs/features/p2p-network.md` claims that `lango p2p identity` does not print the active DID directly
- **THEN** an executable repository test SHALL fail
