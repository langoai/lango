## MODIFIED Requirements

### Requirement: Settings help lists all category groups
The `lango settings --help` output SHALL display all 7 group sections (Core, AI & Knowledge, Automation, Payment & Account, P2P & Economy, Integrations, Security) with their constituent categories.

#### Scenario: User views settings help
- **WHEN** user runs `lango settings --help`
- **THEN** the output lists Core (Providers, Agent, Channels, Tools, Server, Session, Logging, Gatekeeper, Output Manager), AI & Knowledge (Context Profile, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph, Librarian, Retrieval, Auto-Adjust, Context Budget, Agent Memory, Multi-Agent, A2A Protocol, Hooks, Ontology), Automation (Cron, Background, Workflow, RunLedger, Provenance), Payment & Account (Payment, Smart Account), P2P & Economy (P2P Network, P2P Workspace, P2P ZKP, P2P Pricing, P2P Owner, P2P Sandbox, Economy, Risk, Negotiation, Escrow, On-Chain Escrow, Pricing), Integrations (MCP, Observability, Alerting), and Security (Security, Auth, Legacy DB Encryption, KMS, OS Sandbox)
