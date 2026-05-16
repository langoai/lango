## ADDED Requirements
### Requirement: P2P on-chain examples script-pattern guard stays executable
Repository-level regressions that reintroduce stale universal polling-loop claims into the `p2p-onchain-examples` main spec SHALL be enforced by an executable test.

#### Scenario: Universal polling claim is rejected
- **WHEN** `examples/p2p-trading/scripts/test-p2p-trading.sh` still uses a fixed `sleep 15`
- **THEN** an executable repository test SHALL fail if the `p2p-onchain-examples` main spec claims that discovery scripts universally use polling loops instead of fixed sleep
