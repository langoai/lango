## MODIFIED Requirements

### Requirement: Architecture landing and track docs reference dead-letter browsing / status observation
The architecture landing page and P2P knowledge-exchange track doc SHALL describe the richer dead-letter browsing / status observation surface.

#### Scenario: Landing page frames the richer visibility slice
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see dead-letter browsing / status observation framed as filtered read-only visibility

#### Scenario: Track doc reflects the richer landed slice
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** dead-letter browsing / status observation SHALL be described as landed work with richer filtering and pagination
- **AND** the remaining work SHALL be described as actor/time-based filters, raw background-task bridges, and higher-level cockpit or CLI surfaces
