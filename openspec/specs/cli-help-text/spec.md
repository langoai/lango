## Purpose

Defines requirements for accurate and complete `--help` text across Lango CLI commands (settings, doctor, onboard).

## Requirements

### Requirement: Settings help lists all category groups
The `lango settings --help` output SHALL display all 7 group sections (Core, AI & Knowledge, Automation, Payment & Account, P2P & Economy, Integrations, Security) with their constituent categories.

#### Scenario: User views settings help
- **WHEN** user runs `lango settings --help`
- **THEN** the output lists Core (Providers, Agent, Channels, Tools, Server, Session, Logging, Gatekeeper, Output Manager), AI & Knowledge (Context Profile, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph, Librarian, Retrieval, Auto-Adjust, Context Budget, Agent Memory, Multi-Agent, A2A Protocol, Hooks, Ontology), Automation (Cron, Background, Workflow, RunLedger, Provenance), Payment & Account (Payment, Smart Account), P2P & Economy (P2P Network, P2P Workspace, P2P ZKP, P2P Pricing, P2P Owner, P2P Sandbox, Economy, Risk, Negotiation, Escrow, On-Chain Escrow, Pricing), Integrations (MCP, Observability, Alerting), and Security (Security, Auth, Legacy DB Encryption, KMS, OS Sandbox)

### Requirement: Settings help mentions keyword search
The `lango settings --help` output SHALL mention the `/` key for keyword search across categories.

#### Scenario: Search feature documented
- **WHEN** user runs `lango settings --help`
- **THEN** the output includes instruction to press `/` to search across all categories by keyword

### Requirement: Doctor help lists all current check families
The `lango doctor --help` output SHALL list the current diagnostic check families and the current total count.

#### Scenario: User views doctor help
- **WHEN** user runs `lango doctor --help`
- **THEN** the output includes `Checks performed (27 total):`
- **AND** it lists the current families: `Core Configuration`, `Security`, `Context Engineering`, `Memory & Scanning`, `Embedding / RAG`, `Graph / Multi-Agent / A2A`, `Tool Hooks & Agent Management`, `Execution`, `Economy / Contract / Observability`, and `P2P Workspace`

### Requirement: Doctor help documents fix and output flags
The `lango doctor --help` output SHALL describe the `--fix` and `--output` flags in the Long description.

#### Scenario: Flags documented in description
- **WHEN** user runs `lango doctor --help`
- **THEN** the Long description includes usage guidance for `--fix` (automatic repair) and `--output json` (machine-readable output)

### Requirement: Onboard help reflects current provider list
The `lango onboard --help` output SHALL list all supported providers including GitHub in step 1.

#### Scenario: GitHub provider listed
- **WHEN** user runs `lango onboard --help`
- **THEN** step 1 lists Anthropic, OpenAI, Gemini, Ollama, and GitHub as provider choices

### Requirement: Onboard help reflects model auto-fetch
The `lango onboard --help` output SHALL mention that models are auto-fetched from the provider in step 2.

#### Scenario: Auto-fetch mentioned
- **WHEN** user runs `lango onboard --help`
- **THEN** step 2 description includes that model selection uses auto-fetched models from the provider

### Requirement: Onboard help reflects approval policy
The `lango onboard --help` output SHALL mention approval policy in step 4.

#### Scenario: Approval policy mentioned
- **WHEN** user runs `lango onboard --help`
- **THEN** step 4 description includes approval policy alongside privacy interceptor and PII redaction
