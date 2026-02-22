### Requirement: Documentation config blocks use JSON format
All configuration examples in documentation SHALL use JSON fenced code blocks instead of YAML, matching the system's actual `lango config import/export` format.

#### Scenario: Config block format
- **WHEN** a user reads any configuration example in `docs/`
- **THEN** the code block SHALL be fenced with ` ```json ` and contain valid JSON

#### Scenario: Legitimate YAML exceptions
- **WHEN** a code block represents Docker Compose or workflow DAG definitions
- **THEN** the code block SHALL remain as YAML since these are real YAML file formats

### Requirement: TUI navigation hints on config blocks
Each configuration JSON block in documentation SHALL be preceded by a TUI navigation hint showing how to reach that setting.

#### Scenario: Navigation hint format
- **WHEN** a config JSON block is displayed
- **THEN** a blockquote hint in the format `> **Settings:** lango settings -> <MenuName>` SHALL appear immediately before it

### Requirement: No YAML file references in config documentation
Documentation SHALL NOT contain references to `config.yaml` or suggest creating YAML configuration files.

#### Scenario: No config.yaml references
- **WHEN** a user searches documentation for `config.yaml`
- **THEN** zero matches SHALL be found in config-related documentation

### Requirement: Configuration reference includes P2P section
The docs/configuration.md SHALL include a P2P Network section with JSON example, settings table covering all P2PConfig and ZKPConfig fields, and a firewall rule entry sub-table.

#### Scenario: P2P config section present
- **WHEN** the configuration reference documentation is opened
- **THEN** it contains a "P2P Network" section between Payment and Cron with experimental warning badge

#### Scenario: P2P config table complete
- **WHEN** the P2P Network configuration table is read
- **THEN** it includes entries for: p2p.enabled, p2p.listenAddrs, p2p.bootstrapPeers, p2p.keyDir, p2p.enableRelay, p2p.enableMdns, p2p.maxPeers, p2p.handshakeTimeout, p2p.sessionTokenTtl, p2p.autoApproveKnownPeers, p2p.firewallRules, p2p.gossipInterval, p2p.zkHandshake, p2p.zkAttestation, p2p.zkp.proofCacheDir, p2p.zkp.provingScheme
