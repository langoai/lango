## MODIFIED Requirements

### Requirement: P2P tool category in agent identity
The AGENTS.md prompt SHALL include P2P Network as part of thirteen tool categories. The identity section SHALL reference "thirteen tool categories" and include Economy, Contract, and Observability bullets alongside the existing P2P Network bullet.

#### Scenario: Agent identity includes economy, contract, observability
- **WHEN** the agent system prompt is built
- **THEN** the identity section references "thirteen tool categories" and includes Economy, Contract, and Observability bullets

## ADDED Requirements

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
