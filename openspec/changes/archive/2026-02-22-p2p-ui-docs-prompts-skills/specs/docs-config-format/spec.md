## MODIFIED Requirements

### Requirement: Configuration reference includes P2P section
The docs/configuration.md SHALL include a P2P Network section with JSON example, settings table covering all P2PConfig and ZKPConfig fields, and a firewall rule entry sub-table.

#### Scenario: P2P config section present
- **WHEN** the configuration reference documentation is opened
- **THEN** it contains a "P2P Network" section between Payment and Cron with experimental warning badge

#### Scenario: P2P config table complete
- **WHEN** the P2P Network configuration table is read
- **THEN** it includes entries for: p2p.enabled, p2p.listenAddrs, p2p.bootstrapPeers, p2p.keyDir, p2p.enableRelay, p2p.enableMdns, p2p.maxPeers, p2p.handshakeTimeout, p2p.sessionTokenTtl, p2p.autoApproveKnownPeers, p2p.firewallRules, p2p.gossipInterval, p2p.zkHandshake, p2p.zkAttestation, p2p.zkp.proofCacheDir, p2p.zkp.provingScheme
