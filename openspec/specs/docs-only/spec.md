## Purpose

Documentation accuracy requirements ensuring README.md stays in sync with codebase configuration and feature state.

## Requirements

### Requirement: README documents librarian configuration

README.md Configuration Reference table SHALL include all `librarian.*` fields matching `LibrarianConfig` in `internal/config/types.go`.

#### Scenario: Librarian config fields present
- **WHEN** a user reads the Configuration Reference in README.md
- **THEN** the table contains entries for `librarian.enabled`, `librarian.observationThreshold`, `librarian.inquiryCooldownTurns`, `librarian.maxPendingInquiries`, `librarian.autoSaveConfidence`, `librarian.provider`, `librarian.model`

### Requirement: README documents automation defaultDeliverTo

README.md Configuration Reference table SHALL include `defaultDeliverTo` fields for cron, background, and workflow sections.

#### Scenario: defaultDeliverTo fields present
- **WHEN** a user reads the Cron Scheduling, Background Execution, and Workflow Engine config sections
- **THEN** each section contains a `*.defaultDeliverTo` entry with type `[]string` and default `[]`

### Requirement: README multi-agent table reflects librarian tools

The multi-agent orchestration table SHALL list proactive knowledge extraction in the librarian role and include `librarian_pending_inquiries` and `librarian_dismiss_inquiry` in the tools column.

#### Scenario: Librarian row updated
- **WHEN** a user reads the Multi-Agent Orchestration table
- **THEN** the librarian row includes "proactive knowledge extraction" in Role and both `librarian_pending_inquiries` and `librarian_dismiss_inquiry` in Tools

### Requirement: README documents streaming in gateway feature

README.md Features list SHALL describe the Gateway as supporting real-time streaming.

#### Scenario: Gateway feature line updated
- **WHEN** a user reads the Features list in README.md
- **THEN** the Gateway bullet reads "WebSocket/HTTP server with real-time streaming"

### Requirement: README documents observational memory context limit configs

README.md Configuration Reference table SHALL include `observationalMemory.maxReflectionsInContext` and `observationalMemory.maxObservationsInContext` fields matching `ObservationalMemoryConfig` in `internal/config/types.go`.

#### Scenario: Context limit config fields present
- **WHEN** a user reads the Observational Memory config section in README.md
- **THEN** the table contains `observationalMemory.maxReflectionsInContext` (int, default `5`) and `observationalMemory.maxObservationsInContext` (int, default `20`)

### Requirement: README documents embedding cache

README.md Embedding & RAG section SHALL include an Embedding Cache subsection describing in-memory caching with 5-minute TTL and 100-entry limit.

#### Scenario: Embedding cache subsection present
- **WHEN** a user reads the Embedding & RAG section in README.md
- **THEN** there is an "Embedding Cache" heading describing automatic in-memory caching with 5-minute TTL and 100-entry limit

### Requirement: README documents observational memory context limits

README.md Observational Memory section SHALL describe context limits for reflections and observations.

#### Scenario: Context limits bullet present
- **WHEN** a user reads the Observational Memory component list in README.md
- **THEN** there is a "Context Limits" bullet describing default limits of 5 reflections and 20 observations

### Requirement: README documents WebSocket events

README.md SHALL include a WebSocket Events subsection documenting `agent.thinking`, `agent.chunk`, and `agent.done` events with their payloads.

#### Scenario: WebSocket events table present
- **WHEN** a user reads the WebSocket section in README.md
- **THEN** there is a "WebSocket Events" heading with a table listing `agent.thinking`, `agent.chunk`, and `agent.done` events

#### Scenario: Backward compatibility noted
- **WHEN** a user reads the WebSocket Events section
- **THEN** there is a note that clients not handling `agent.chunk` will still receive the full response in the RPC result

### Requirement: Documentation accuracy

Documentation, prompts, and CLI help text SHALL accurately reflect all implemented features including P2P REST API endpoints, CLI flags, and example projects.

#### Scenario: P2P REST API documented
- **WHEN** a user reads the HTTP API documentation
- **THEN** the P2P REST endpoints (`/api/p2p/status`, `/api/p2p/peers`, `/api/p2p/identity`) SHALL be documented with request/response examples

#### Scenario: Secrets --value-hex documented
- **WHEN** a user reads the secrets set CLI documentation
- **THEN** the `--value-hex` flag SHALL be documented with non-interactive usage examples

#### Scenario: P2P trading example discoverable
- **WHEN** a user reads the README
- **THEN** the `examples/p2p-trading/` directory SHALL be referenced in an Examples section

### Requirement: Approval Pipeline documentation in P2P feature docs
The `docs/features/p2p-network.md` file SHALL include an "Approval Pipeline" section describing the three-stage inbound gate (Firewall ACL → Owner Approval → Tool Execution) with a Mermaid flowchart diagram and auto-approval shortcut rules table.

#### Scenario: Approval Pipeline section present
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** there SHALL be an "Approval Pipeline" section between Knowledge Firewall and Discovery with a Mermaid diagram and descriptions of all three stages

### Requirement: Auto-Approval for Small Amounts in Paid Value Exchange docs
The Paid Value Exchange section in `docs/features/p2p-network.md` SHALL include an "Auto-Approval for Small Amounts" subsection describing the three conditions checked by `IsAutoApprovable`: threshold, maxPerTx, and maxDaily.

#### Scenario: Auto-approval subsection present
- **WHEN** a user reads the Paid Value Exchange section
- **THEN** there SHALL be a subsection documenting the three auto-approval conditions and fallback to interactive approval

### Requirement: Reputation and Pricing endpoints in REST API tables
All REST API documentation (p2p-network.md, http-api.md, README.md, examples/p2p-trading/README.md) SHALL list `GET /api/p2p/reputation` and `GET /api/p2p/pricing` with curl examples and JSON response samples.

#### Scenario: Endpoints in p2p-network.md
- **WHEN** a user reads the REST API table in `docs/features/p2p-network.md`
- **THEN** reputation and pricing endpoints SHALL be listed with curl examples

#### Scenario: Endpoints in http-api.md
- **WHEN** a user reads `docs/gateway/http-api.md`
- **THEN** there SHALL be full endpoint sections for reputation and pricing with query parameters, JSON response examples, and curl commands

### Requirement: Reputation and Pricing CLI commands documented
The CLI command listings in `docs/features/p2p-network.md` and `README.md` SHALL include `lango p2p reputation` and `lango p2p pricing` commands.

#### Scenario: CLI commands in feature docs
- **WHEN** a user reads the CLI Commands section of `docs/features/p2p-network.md`
- **THEN** reputation and pricing commands SHALL be listed

### Requirement: README P2P config fields complete
The README.md P2P configuration reference table SHALL include `p2p.autoApproveKnownPeers`, `p2p.minTrustScore`, `p2p.pricing.enabled`, and `p2p.pricing.perQuery` fields.

#### Scenario: Missing config fields added
- **WHEN** a user reads the P2P Network section of the Configuration Reference in README.md
- **THEN** all four fields SHALL be present with correct types, defaults, and descriptions

### Requirement: Tool usage prompts reflect approval behavior
The `prompts/TOOL_USAGE.md` file SHALL describe auto-approval behavior for `p2p_pay`, the remote owner's approval pipeline for `p2p_query`, and inbound tool invocation gates.

#### Scenario: p2p_pay auto-approval documented
- **WHEN** a user reads the `p2p_pay` description
- **THEN** it SHALL mention that payments below `autoApproveBelow` are auto-approved

#### Scenario: Inbound invocation gates documented
- **WHEN** a user reads the P2P Networking Tool section
- **THEN** there SHALL be a description of the three-stage inbound gate

### Requirement: USDC docs cross-reference P2P auto-approval
The `docs/payments/usdc.md` file SHALL include a P2P integration note explaining that `autoApproveBelow` applies to both outbound payments and inbound paid tool approval.

#### Scenario: P2P integration note present
- **WHEN** a user reads `docs/payments/usdc.md`
- **THEN** there SHALL be a note after the config table linking to the P2P approval pipeline

### Requirement: P2P trading example documents configuration highlights
The `examples/p2p-trading/README.md` SHALL include a "Configuration Highlights" section with a table of key approval and payment settings used in the example.

#### Scenario: Configuration highlights section present
- **WHEN** a user reads the example README
- **THEN** there SHALL be a Configuration Highlights section with autoApproveBelow, autoApproveKnownPeers, pricing settings, and a production warning

### Requirement: test-p2p Makefile target
The root `Makefile` SHALL include a `test-p2p` target that runs `go test -v -race ./internal/p2p/... ./internal/wallet/...` and SHALL be listed in the `.PHONY` declaration.

#### Scenario: test-p2p target runs successfully
- **WHEN** a user runs `make test-p2p`
- **THEN** P2P and wallet tests SHALL execute with race detector enabled
