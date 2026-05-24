## ADDED Requirements
### Requirement: Main specs avoid stale P2P trading script paths
Main specs SHALL not keep stale single-file example-script references after repository paths move.

#### Scenario: Stale integration test script path is rejected
- **WHEN** a maintainer updates the `p2p-trading-example` main spec
- **THEN** it SHALL not claim `scripts/test-p2p-trading.sh` as the current integration test script path
