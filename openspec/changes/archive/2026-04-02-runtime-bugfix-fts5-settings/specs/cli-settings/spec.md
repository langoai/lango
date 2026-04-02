## MODIFIED Requirements

### Requirement: Settings menu category list
The settings menu SHALL organize categories into named sections. The sections SHALL be, in order:
1. **Core** — Providers, Agent, Channels, Tools, Server (advanced), Session (advanced), Logging (advanced), Gatekeeper (advanced), Output Manager (advanced)
2. **AI & Knowledge** — Context Profile, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store (advanced), Librarian (advanced), Retrieval (advanced), Auto-Adjust (advanced), Context Budget (advanced), Agent Memory (advanced), Multi-Agent (advanced), A2A Protocol (advanced), Hooks (advanced), Ontology (advanced)
3. **Automation** — Cron Scheduler, Background Tasks (advanced), Workflow Engine (advanced), RunLedger (advanced), Provenance (advanced)
4. **Payment & Account** — Payment, Smart Account (advanced), SA Session Keys (advanced), SA Paymaster (advanced), SA Modules (advanced)
5. **P2P & Economy** — P2P Network (advanced), P2P Workspace (advanced), P2P ZKP (advanced), P2P Pricing (advanced), P2P Owner Protection (advanced), P2P Sandbox (advanced), Economy (advanced), Economy Risk (advanced), Economy Negotiation (advanced), Economy Escrow (advanced), On-Chain Escrow (advanced), Economy Pricing (advanced)
6. **Integrations** — MCP Settings, MCP Server List (advanced), Observability (advanced), Alerting (advanced)
7. **Security** — Security, Auth (advanced), Security DB Encryption (advanced), Security KMS (advanced), OS Sandbox (advanced)
8. *(untitled)* — Save & Exit, Cancel

#### Scenario: Level 1 section list displayed
- **WHEN** user views the settings menu at Level 1
- **THEN** 7 named section items SHALL be rendered with category counts, followed by a separator and Save & Cancel items
