## ADDED Requirements

### Requirement: Comprehensive disabled category registration
Every tool subsystem SHALL register a disabled category with the tool catalog when it is not active, so that builtin_health diagnostics can report the full system state. The disabled category SHALL include the relevant configKey.

#### Scenario: Disabled subsystems appear in catalog
- **WHEN** a subsystem (browser, crypto, secrets, meta, graph, rag, memory, agent_memory, payment, p2p, librarian, economy, mcp, observability, contract, workspace) is disabled
- **THEN** a disabled category is registered with Name, Description containing "(disabled)", ConfigKey, and Enabled=false

#### Scenario: builtin_health reports disabled subsystems
- **WHEN** builtin_health runs diagnostics
- **THEN** all disabled subsystems appear in the disabled list of the tool registration summary

#### Scenario: P2P disabled with payment dependency
- **WHEN** p2p.enabled is true but payment is disabled
- **THEN** p2p disabled category description includes "(disabled — payment required)"
