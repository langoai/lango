## ADDED Requirements
### Requirement: P2P trading example spec reflects the current Docker entrypoint path
The `p2p-trading-example` main spec SHALL point to the current repository path for the shipped Docker bootstrap entrypoint.

#### Scenario: Stale Docker entrypoint path is rejected
- **WHEN** the repository keeps the bootstrap entrypoint at `examples/p2p-trading/docker-entrypoint-p2p.sh`
- **THEN** the `p2p-trading-example` main spec SHALL not claim `docker-entrypoint-p2p.sh` as the current path
