## MODIFIED Requirements

### Requirement: CLI command group organization
The CLI SHALL organize commands into five user-intent groups: "Getting Started", "AI & Knowledge", "Automation", "Network & Economy", "Security & System".

#### Scenario: Getting Started group
- **WHEN** user runs `lango --help`
- **THEN** Getting Started section contains: serve, onboard, doctor, settings, status, version

#### Scenario: AI & Knowledge group
- **WHEN** user runs `lango --help`
- **THEN** AI & Knowledge section contains: agent, memory, learning, graph, librarian, a2a, metrics

#### Scenario: Automation group
- **WHEN** user runs `lango --help`
- **THEN** Automation section contains: cron, workflow, bg

#### Scenario: Network & Economy group
- **WHEN** user runs `lango --help`
- **THEN** Network & Economy section contains: p2p, payment, economy, contract, account, mcp

#### Scenario: Security & System group
- **WHEN** user runs `lango --help`
- **THEN** Security & System section contains: security, approval, health, config
