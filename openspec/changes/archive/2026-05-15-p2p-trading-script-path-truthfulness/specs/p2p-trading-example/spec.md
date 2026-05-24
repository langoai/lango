## ADDED Requirements
### Requirement: P2P trading example spec reflects the current test script path
The `p2p-trading-example` main spec SHALL point to the current repository path for the shipped integration test script.

#### Scenario: Stale integration test script path is rejected
- **WHEN** the repository keeps the integration test runner at `examples/p2p-trading/scripts/test-p2p-trading.sh`
- **THEN** the `p2p-trading-example` main spec SHALL not claim `scripts/test-p2p-trading.sh` as the current script path
