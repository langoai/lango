## ADDED Requirements
### Requirement: Main specs avoid stale P2P trading contract paths
Main specs SHALL not keep stale single-file contract references after repository paths move.

#### Scenario: Stale MockUSDC spec path is rejected
- **WHEN** a maintainer updates the `p2p-trading-example` main spec
- **THEN** it SHALL not claim `contracts/MockUSDC.sol` as the current mock contract path
