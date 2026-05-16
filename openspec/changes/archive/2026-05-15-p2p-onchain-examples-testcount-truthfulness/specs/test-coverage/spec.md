## ADDED Requirements
### Requirement: P2P on-chain examples exact-count guard stays executable
Repository-level regressions that reintroduce stale exact `Tests (N)` claims into the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Stale exact test-count claims are rejected
- **WHEN** shipped example scripts evolve independently
- **THEN** an executable repository test SHALL fail if the `p2p-onchain-examples` main spec claims hard-coded `Tests (N)` totals
