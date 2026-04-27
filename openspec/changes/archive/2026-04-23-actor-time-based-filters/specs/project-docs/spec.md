## MODIFIED Requirements

### Requirement: Architecture landing and track docs reference dead-letter browsing / status observation
The architecture landing page and P2P knowledge-exchange track doc SHALL describe the actor/time-aware dead-letter browsing surface.

#### Scenario: Landing page frames actor/time-aware visibility
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see dead-letter browsing / status observation framed as actor/time-aware read-only visibility

#### Scenario: Track doc reflects the richer landed slice
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** dead-letter browsing / status observation SHALL be described as landed work with actor/time-based filters
- **AND** the remaining work SHALL be described as richer reason/dispatch filters, raw background-task bridges, and higher-level cockpit or CLI surfaces
