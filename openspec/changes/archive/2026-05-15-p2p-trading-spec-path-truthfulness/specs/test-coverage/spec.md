## ADDED Requirements
### Requirement: P2P trading broken-path guard stays executable
Repository-level regressions that reintroduce stale MockUSDC path claims into the `p2p-trading-example` main spec SHALL be enforced by an executable test.

#### Scenario: Stale MockUSDC path is rejected
- **WHEN** the shipped mock contract lives at `contracts/test/mocks/MockUSDC.sol`
- **THEN** an executable repository test SHALL fail if the `p2p-trading-example` main spec claims `contracts/MockUSDC.sol` is the current path
