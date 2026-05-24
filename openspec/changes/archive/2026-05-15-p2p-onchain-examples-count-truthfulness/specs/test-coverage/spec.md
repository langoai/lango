## ADDED Requirements
### Requirement: P2P on-chain examples count guard stays executable
Repository-level regressions that reintroduce a stale shipped-example count into the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Stale six-example claim is rejected
- **WHEN** seven example directories ship under `examples/`
- **THEN** an executable repository test SHALL fail if the `p2p-onchain-examples` main spec claims there are only six Docker Compose examples
