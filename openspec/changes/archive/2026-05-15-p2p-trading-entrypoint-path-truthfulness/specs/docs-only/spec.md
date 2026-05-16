## ADDED Requirements
### Requirement: Main specs avoid stale P2P trading entrypoint paths
Main specs SHALL not keep stale single-file example-entrypoint references after repository paths move.

#### Scenario: Stale Docker entrypoint path is rejected
- **WHEN** a maintainer updates the `p2p-trading-example` main spec
- **THEN** it SHALL not claim `docker-entrypoint-p2p.sh` as the current Docker entrypoint path
