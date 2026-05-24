## ADDED Requirements
### Requirement: P2P trading script-path guard stays executable
Repository-level regressions that reintroduce stale integration test script path claims into the `p2p-trading-example` main spec SHALL be enforced by an executable test.

#### Scenario: Stale integration test script path is rejected
- **WHEN** the shipped integration test runner lives at `examples/p2p-trading/scripts/test-p2p-trading.sh`
- **THEN** an executable repository test SHALL fail if the `p2p-trading-example` main spec claims `scripts/test-p2p-trading.sh` is the current path
