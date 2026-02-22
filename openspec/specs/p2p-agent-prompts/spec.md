## ADDED Requirements

### Requirement: P2P tool category in agent identity
The AGENTS.md prompt SHALL include P2P Network as the 10th tool category describing peer connectivity, firewall ACL management, remote agent querying, capability-based discovery, and peer payments with Noise encryption and DID identity verification.

#### Scenario: Agent identity includes P2P
- **WHEN** the agent system prompt is built
- **THEN** the identity section references "ten tool categories" and includes a P2P Network bullet

### Requirement: P2P tool usage guidelines
The TOOL_USAGE.md prompt SHALL include a "P2P Networking Tool" section documenting all P2P tools: p2p_status, p2p_connect, p2p_disconnect, p2p_peers, p2p_query, p2p_discover, p2p_firewall_rules, p2p_firewall_add, p2p_firewall_remove, p2p_pay.

#### Scenario: Tool usage includes P2P section
- **WHEN** the agent system prompt is built
- **THEN** the tool usage section includes P2P Networking Tool guidelines with session token and firewall deny behavior notes

### Requirement: Vault agent P2P role
The vault agent IDENTITY.md SHALL include P2P peer management and firewall rule management as part of its responsibilities.

#### Scenario: Vault identity covers P2P
- **WHEN** the vault sub-agent prompt is built
- **THEN** the identity mentions P2P networking alongside crypto, secrets, and payment operations
