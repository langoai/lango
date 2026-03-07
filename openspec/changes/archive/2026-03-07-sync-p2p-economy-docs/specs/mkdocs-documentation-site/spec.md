## MODIFIED Requirements

### Requirement: Feature documentation coverage
The documentation SHALL have dedicated pages for: AI Providers, Channels, Knowledge System, Observational Memory, Embedding & RAG, Knowledge Graph, Multi-Agent Orchestration, A2A Protocol, P2P Network, P2P Economy, Smart Contracts, Observability, Skill System, Proactive Librarian, and System Prompts.

#### Scenario: All features documented
- **WHEN** a user browses the Features section
- **THEN** each feature SHALL have its own page with configuration reference and usage examples

### Requirement: CLI reference documentation
The documentation SHALL include a complete CLI reference organized by command category: Core, Config Management, Agent & Memory, Security, Payment, P2P, Economy, Contract, Metrics, and Automation commands.

#### Scenario: CLI commands documented
- **WHEN** a user looks up a CLI command
- **THEN** they SHALL find syntax, flags, and usage examples

### Requirement: Navigation includes P2P pages
The mkdocs.yml navigation SHALL include "P2P Network: features/p2p-network.md", "P2P Economy: features/economy.md", "Smart Contracts: features/contracts.md", and "Observability: features/observability.md" in the Features section and "P2P Commands: cli/p2p.md", "Economy Commands: cli/economy.md", "Contract Commands: cli/contract.md", and "Metrics Commands: cli/metrics.md" in the CLI Reference section.

#### Scenario: Economy, contract, observability features in nav
- **WHEN** the mkdocs site is built
- **THEN** the Features navigation section includes "P2P Economy", "Smart Contracts", and "Observability" entries after "P2P Network"

#### Scenario: Economy, contract, metrics CLI in nav
- **WHEN** the mkdocs site is built
- **THEN** the CLI Reference navigation section includes "Economy Commands", "Contract Commands", and "Metrics Commands" entries after "P2P Commands"

### Requirement: Configuration reference
The documentation SHALL include a complete configuration reference page listing all configuration keys with type, default value, and description, organized by category, including Economy and Observability sections.

#### Scenario: Configuration completeness
- **WHEN** the configuration reference is viewed
- **THEN** it SHALL list all configuration keys including economy.* and observability.* sections
