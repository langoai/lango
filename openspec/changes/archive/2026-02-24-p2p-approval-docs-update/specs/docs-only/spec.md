## ADDED Requirements

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
