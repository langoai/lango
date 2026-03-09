## ADDED Requirements

### Requirement: P2P tool category in agent identity
The AGENTS.md prompt SHALL include P2P Network as part of thirteen tool categories. The identity section SHALL reference "thirteen tool categories" and include Economy, Contract, and Observability bullets alongside the existing P2P Network bullet.

#### Scenario: Agent identity includes P2P
- **WHEN** the agent system prompt is built
- **THEN** the identity section references "thirteen tool categories" and includes a P2P Network bullet

#### Scenario: Agent identity includes economy, contract, observability
- **WHEN** the agent system prompt is built
- **THEN** the identity section references "thirteen tool categories" and includes Economy, Contract, and Observability bullets

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

### Requirement: Agent prompts include paid value exchange
The agent prompt files SHALL describe paid value exchange capabilities including pricing query, reputation checking, and owner shield protection.

#### Scenario: AGENTS.md describes paid P2P features
- **WHEN** agent loads AGENTS.md system prompt
- **THEN** P2P Network description includes pricing query, reputation tracking, owner shield, and USDC Payment Gate

#### Scenario: TOOL_USAGE.md documents new tools
- **WHEN** agent loads TOOL_USAGE.md
- **THEN** P2P section includes `p2p_price_query`, `p2p_reputation` tool descriptions and paid tool workflow guidance

#### Scenario: Vault IDENTITY.md includes new capabilities
- **WHEN** vault agent loads IDENTITY.md
- **THEN** role description includes reputation and pricing management, and REST API list includes `/api/p2p/reputation` and `/api/p2p/pricing`

### Requirement: Economy tool usage guidelines
The TOOL_USAGE.md prompt SHALL include an "Economy Tool" section documenting all 13 economy tools: economy_budget_allocate, economy_budget_status, economy_budget_close, economy_risk_assess, economy_price_quote, economy_negotiate, economy_negotiate_status, economy_escrow_create, economy_escrow_fund, economy_escrow_milestone, economy_escrow_release, economy_escrow_status, economy_escrow_dispute. The section SHALL include workflow guidance: budget → risk → pricing → negotiation → escrow.

#### Scenario: Tool usage includes Economy section
- **WHEN** the agent system prompt is built
- **THEN** the tool usage section includes Economy Tool guidelines with all 13 tools and workflow order

### Requirement: Contract tool usage guidelines
The TOOL_USAGE.md prompt SHALL include a "Contract Tool" section documenting 3 tools: contract_read (Safe), contract_call (Dangerous), contract_abi_load (Safe). The section SHALL include guidance to load ABI first, read before write.

#### Scenario: Tool usage includes Contract section
- **WHEN** the agent system prompt is built
- **THEN** the tool usage section includes Contract Tool guidelines with all 3 tools

### Requirement: Exec tool blocklist updated
The TOOL_USAGE.md exec tool blocklist SHALL include `lango economy`, `lango metrics`, and `lango contract` to prevent CLI bypass of agent tools.

#### Scenario: Blocklist includes new command groups
- **WHEN** the agent checks exec tool blocklist
- **THEN** `lango economy`, `lango metrics`, and `lango contract` SHALL be listed as blocked commands
