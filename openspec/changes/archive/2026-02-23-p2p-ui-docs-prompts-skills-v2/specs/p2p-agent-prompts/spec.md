## MODIFIED Requirements

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
