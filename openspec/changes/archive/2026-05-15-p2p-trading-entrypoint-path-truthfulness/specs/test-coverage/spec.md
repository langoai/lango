## ADDED Requirements
### Requirement: P2P trading entrypoint-path guard stays executable
Repository-level regressions that reintroduce stale Docker entrypoint path claims into the `p2p-trading-example` main spec SHALL be enforced by an executable test.

#### Scenario: Stale Docker entrypoint path is rejected
- **WHEN** the shipped bootstrap entrypoint lives at `examples/p2p-trading/docker-entrypoint-p2p.sh`
- **THEN** an executable repository test SHALL fail if the `p2p-trading-example` main spec claims `docker-entrypoint-p2p.sh` is the current path
