## ADDED Requirements

### Requirement: Architecture landing and track docs reference dead-letter browsing / status observation
The architecture landing page and P2P knowledge-exchange track doc SHALL reference the landed dead-letter browsing / status observation slice.

#### Scenario: Landing page links dead-letter browsing / status observation
- **WHEN** a reader opens `docs/architecture/index.md`
- **THEN** they SHALL see the Dead-Letter Browsing / Status Observation page listed with the other architecture pages

#### Scenario: Track doc reflects landed dead-letter browsing / status observation
- **WHEN** a reader opens `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** dead-letter browsing / status observation SHALL be described as landed slice work with richer filtering, raw background-task bridges, and higher-level cockpit or CLI surfaces still remaining
