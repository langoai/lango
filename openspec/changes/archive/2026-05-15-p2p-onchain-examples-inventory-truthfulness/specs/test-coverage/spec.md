## ADDED Requirements
### Requirement: P2P on-chain examples inventory guard stays executable
Repository-level regressions that omit shipped example summaries from the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Missing shipped example summary is rejected
- **WHEN** the repository still ships `examples/p2p-trading`
- **THEN** an executable repository test SHALL fail if the `p2p-onchain-examples` main spec omits the corresponding numbered example summary
