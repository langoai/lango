## ADDED Requirements

### Requirement: Vault agent prefix list covers all vault tool families
The vault AgentSpec Prefixes SHALL include prefixes for all vault-domain tool families: crypto, secrets, payment, p2p, smartaccount, economy, escrow, sentinel, and contract.

#### Scenario: Vault prefixes include smartaccount tools
- **WHEN** the vault spec's Prefixes are checked
- **THEN** they SHALL include `smart_account_`, `session_key_`, `session_execute`, `policy_check`, `module_`, `spending_`, `paymaster_`

#### Scenario: Vault prefixes include economy/escrow/sentinel/contract tools
- **WHEN** the vault spec's Prefixes are checked
- **THEN** they SHALL include `economy_`, `escrow_`, `sentinel_`, `contract_`

### Requirement: Dynamic and builtin vault specs are synchronized
The embedded AGENT.md for vault and the builtin vault spec in agentSpecs SHALL have identical prefix and keyword lists.

#### Scenario: AGENT.md prefixes match builtin spec
- **WHEN** the vault AGENT.md frontmatter prefixes are compared to agentSpecs vault Prefixes
- **THEN** the lists SHALL contain the same entries

### Requirement: capabilityMap covers all vault prefixes
Every prefix in the vault AgentSpec SHALL have a corresponding entry in capabilityMap for diagnostics.

#### Scenario: All vault prefixes have capability entries
- **WHEN** toolCapability is called with tool names starting with any vault prefix
- **THEN** it SHALL return a non-empty capability description
