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

### Requirement: Quickstart references config presets
The getting started quickstart documentation SHALL reference the `--preset` flag and link to the config presets documentation.

#### Scenario: Preset flag in quickstart
- **WHEN** a user reads `docs/getting-started/quickstart.md`
- **THEN** the `--preset` flag SHALL be mentioned with a brief preset table and link to `config-presets.md`

### Requirement: CLI index includes status command
The CLI index quick reference table SHALL include the `lango status` command.

#### Scenario: Status in CLI index
- **WHEN** a user reads `docs/cli/index.md`
- **THEN** `lango status` SHALL appear in the Quick Reference table under Getting Started

### Requirement: Quickstart installation anchor resolves
The getting started quickstart documentation SHALL link to the existing installation anchor instead of a missing fragment.

#### Scenario: Installation anchor is valid
- **WHEN** a user reads `docs/getting-started/quickstart.md`
- **THEN** the installation link SHALL target the existing installation section and its compiler setup anchor

### Requirement: Cockpit public-entry consolidation
After the hidden cockpit guides move out of `docs/`, the public cockpit documentation SHALL keep `docs/features/cockpit.md` as the single public entry for operator-facing material from the cockpit approval, channels, tasks, and troubleshooting guides.

#### Scenario: Approval guidance is on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find approval operations guidance previously split into the approval sub-guide

#### Scenario: Channel, task, and troubleshooting guidance are on the main cockpit page
- **WHEN** a user reads `docs/features/cockpit.md`
- **THEN** they SHALL find channel operations, background task operations, and troubleshooting guidance previously split into the other cockpit sub-guides

### Requirement: P2P knowledge exchange track reflects the landed identity trust reputation audit
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the identity/trust/reputation detailed audit as landed work and list the follow-on work as `reputation v2`, stronger trust-entry contracts, and runtime integration.

#### Scenario: Track follow-on list is updated
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the required follow-on plan SHALL state that the identity/trust/reputation detailed audit is now landed
- **AND** the follow-on work SHALL include `reputation v2`, stronger trust-entry contracts, and runtime integration

### Requirement: P2P knowledge exchange track reflects the landed pricing negotiation settlement audit
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the pricing/negotiation/settlement detailed audit as landed work and list the follow-on work as `runtime integration`, `settlement execution`, and `escrow lifecycle completion`.

#### Scenario: Track follow-on list is updated
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the required follow-on plan SHALL state that the pricing/negotiation/settlement detailed audit is now landed
- **AND** the follow-on work SHALL include `runtime integration`, `settlement execution`, and `escrow lifecycle completion`

### Requirement: Knowledge exchange runtime architecture page describes the first control-plane slice
The `docs/architecture/knowledge-exchange-runtime.md` page SHALL describe the first transaction-oriented runtime control-plane design slice for `knowledge exchange v1`, centered on transaction receipt and submission receipt, and SHALL list the current limits of that slice.

#### Scenario: Runtime page shows the bounded slice
- **WHEN** a user reads `docs/architecture/knowledge-exchange-runtime.md`
- **THEN** they SHALL find sections covering the runtime design slice, canonical state, current limits, and follow-on work

### Requirement: P2P knowledge exchange track links the runtime design slice
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL reference `knowledge-exchange-runtime.md` as the first transaction-oriented runtime design slice and SHALL state that the remaining work is runtime implementation and broader progression handling.

#### Scenario: Track page points to the runtime slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find the runtime design slice referenced by name and linked to `knowledge-exchange-runtime.md`
- **AND** the follow-on work SHALL be described as implementation, not redesign of the landed slice

### Requirement: Settlement progression architecture page describes the first progression slice
The `docs/architecture/settlement-progression.md` page SHALL describe the first transaction-level settlement progression slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Settlement progression page shows the bounded slice
- **WHEN** a user reads `docs/architecture/settlement-progression.md`
- **THEN** they SHALL find sections describing the current progression slice, what ships, canonical state, and current limits

### Requirement: P2P knowledge exchange track reflects landed settlement progression
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the settlement progression first slice as landed work and list the remaining work as actual settlement execution, partial-settlement rules, and dispute engine completion.

#### Scenario: Track page points to the landed settlement progression slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find settlement progression described as a landed first slice
- **AND** the remaining work SHALL be described as actual settlement execution, partial-settlement rules, and dispute engine completion

### Requirement: Actual settlement execution page describes the first direct execution slice
The `docs/architecture/actual-settlement-execution.md` page SHALL describe the first direct settlement execution slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Actual settlement execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/actual-settlement-execution.md`
- **THEN** they SHALL find sections describing the current execution slice, what ships, canonical gate, and current limits

### Requirement: P2P knowledge exchange track reflects landed actual settlement execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the actual settlement execution first slice as landed work and list the remaining work as escrow lifecycle completion and dispute engine completion.

#### Scenario: Track page points to the landed actual settlement execution slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find actual settlement execution described as a landed first slice
- **AND** the remaining work SHALL be described as escrow lifecycle completion and dispute engine completion

### Requirement: Partial settlement execution page describes the first direct partial slice
The `docs/architecture/partial-settlement-execution.md` page SHALL describe the first direct partial settlement execution slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Partial settlement execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/partial-settlement-execution.md`
- **THEN** they SHALL find sections describing the current partial slice, canonical hint model, success/failure semantics, and current limits

### Requirement: P2P knowledge exchange track reflects landed partial settlement execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the partial settlement execution first slice as landed work and list the remaining work as escrow lifecycle completion and dispute engine completion.

#### Scenario: Track page points to the landed partial settlement execution slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find partial settlement execution described as a landed first slice
- **AND** the remaining work SHALL be described as escrow lifecycle completion and dispute engine completion

### Requirement: Escrow release page describes the first funded release slice
The `docs/architecture/escrow-release.md` page SHALL describe the first escrow release slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Escrow release page shows the bounded slice
- **WHEN** a user reads `docs/architecture/escrow-release.md`
- **THEN** they SHALL find sections describing the current escrow release slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed escrow release
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the escrow release first slice as landed work and list the remaining work as refund, dispute-linked escrow handling, and milestone-aware release.

#### Scenario: Track page points to the landed escrow release slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find escrow release described as a landed first slice
- **AND** the remaining work SHALL be described as refund, dispute-linked escrow handling, and milestone-aware release

### Requirement: Escrow refund page describes the first funded refund slice
The `docs/architecture/escrow-refund.md` page SHALL describe the first escrow refund slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Escrow refund page shows the bounded slice
- **WHEN** a user reads `docs/architecture/escrow-refund.md`
- **THEN** they SHALL find sections describing the current escrow refund slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed escrow refund
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the escrow refund first slice as landed work and list the remaining work as refund terminal-state design, dispute-linked refund branching, and release-after-refund safety rules.

#### Scenario: Track page points to the landed escrow refund slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find escrow refund described as a landed first slice
- **AND** the remaining work SHALL be described as refund terminal-state design, dispute-linked refund branching, and release-after-refund safety rules

### Requirement: Dispute hold page describes the first funded dispute hold slice
The `docs/architecture/dispute-hold.md` page SHALL describe the first dispute hold slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Dispute hold page shows the bounded slice
- **WHEN** a user reads `docs/architecture/dispute-hold.md`
- **THEN** they SHALL find sections describing the current dispute hold slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed dispute hold
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the dispute hold first slice as landed work and list the remaining work as release-vs-refund adjudication, explicit held-state design, and dispute engine integration.

#### Scenario: Track page points to the landed dispute hold slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dispute hold described as a landed first slice
- **AND** the remaining work SHALL be described as release-vs-refund adjudication, explicit held-state design, and dispute engine integration

### Requirement: Release-vs-refund adjudication page describes the first post-hold branching slice
The `docs/architecture/release-vs-refund-adjudication.md` page SHALL describe the first post-hold release-vs-refund adjudication slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Adjudication page shows the bounded slice
- **WHEN** a user reads `docs/architecture/release-vs-refund-adjudication.md`
- **THEN** they SHALL find sections describing the current adjudication slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed release-vs-refund adjudication
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the release-vs-refund adjudication first slice as landed work and list the remaining work as adjudication-aware release/refund execution, keep-hold or re-escalation states, and broader dispute engine integration.

#### Scenario: Track page points to the landed adjudication slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find release-vs-refund adjudication described as a landed first slice
- **AND** the remaining work SHALL be described as adjudication-aware release/refund execution, keep-hold or re-escalation states, and broader dispute engine integration

### Requirement: Adjudication-aware release/refund execution gating page describes the first executor-contract slice
The `docs/architecture/adjudication-aware-release-refund-execution-gating.md` page SHALL describe the first slice that connects canonical escrow adjudication to release/refund execution gating, including what currently ships and the current limits of the slice.

#### Scenario: Adjudication-aware execution gating page shows the bounded slice
- **WHEN** a user reads `docs/architecture/adjudication-aware-release-refund-execution-gating.md`
- **THEN** they SHALL find sections describing the current execution-gating slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed adjudication-aware release/refund execution gating
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the adjudication-aware release/refund execution gating first slice as landed work and list the remaining work as automatic post-adjudication execution, keep-hold or re-escalation states, and broader dispute engine integration.

#### Scenario: Track page points to the landed adjudication-aware gating slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find adjudication-aware release/refund execution gating described as a landed first slice
- **AND** the remaining work SHALL be described as automatic post-adjudication execution, keep-hold or re-escalation states, and broader dispute engine integration

### Requirement: Automatic post-adjudication execution page describes the first inline orchestration slice
The `docs/architecture/automatic-post-adjudication-execution.md` page SHALL describe the first inline convenience slice after escrow adjudication, including what currently ships and the current limits of the slice.

#### Scenario: Automatic post-adjudication execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/automatic-post-adjudication-execution.md`
- **THEN** they SHALL find sections describing the current auto-execution slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed automatic post-adjudication execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the automatic post-adjudication execution first slice as landed work and list the remaining work as background execution, retry orchestration, automatic execution as policy default, and broader dispute engine integration.

#### Scenario: Track page points to the landed auto-execution slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find automatic post-adjudication execution described as a landed first slice
- **AND** the remaining work SHALL be described as background execution, retry orchestration, automatic execution as policy default, and broader dispute engine integration

### Requirement: Background post-adjudication execution page describes the first async dispatch slice
The `docs/architecture/background-post-adjudication-execution.md` page SHALL describe the first background post-adjudication execution slice for `knowledge exchange v1`, including what currently ships and the current limits of the slice.

#### Scenario: Background post-adjudication execution page shows the bounded slice
- **WHEN** a user reads `docs/architecture/background-post-adjudication-execution.md`
- **THEN** they SHALL find sections describing the current background dispatch slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed background post-adjudication execution
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the background post-adjudication execution first slice as landed work and list the remaining work as retry orchestration, dead-letter handling, dedicated status observation, and policy-driven defaults.

#### Scenario: Track page points to the landed background slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find background post-adjudication execution described as a landed first slice
- **AND** the remaining work SHALL be described as retry orchestration, dead-letter handling, dedicated status observation, and policy-driven defaults

### Requirement: Retry / dead-letter handling page describes the first bounded retry slice
The `docs/architecture/retry-dead-letter-handling.md` page SHALL describe the first retry / dead-letter slice for background post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Retry / dead-letter handling page shows the bounded slice
- **WHEN** a user reads `docs/architecture/retry-dead-letter-handling.md`
- **THEN** they SHALL find sections describing the current retry/dead-letter slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed retry / dead-letter handling
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the retry / dead-letter handling first slice as landed work and list the remaining work as operator replay, generic async retry policy, dead-letter browsing, and policy-driven backoff tuning.

#### Scenario: Track page points to the landed retry slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find retry / dead-letter handling described as a landed first slice
- **AND** the remaining work SHALL be described as operator replay, generic async retry policy, dead-letter browsing, and policy-driven backoff tuning

### Requirement: Operator replay / manual retry page describes the first replay slice
The `docs/architecture/operator-replay-manual-retry.md` page SHALL describe the first operator replay / manual retry slice for dead-lettered post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Operator replay / manual retry page shows the bounded slice
- **WHEN** a user reads `docs/architecture/operator-replay-manual-retry.md`
- **THEN** they SHALL find sections describing the current replay slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed operator replay / manual retry
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the operator replay / manual retry first slice as landed work and list the remaining work as dead-letter browsing UI, policy-driven replay controls, generic replay substrate design, and broader dispute engine integration.

#### Scenario: Track page points to the landed replay slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find operator replay / manual retry described as a landed first slice
- **AND** the remaining work SHALL be described as dead-letter browsing UI, policy-driven replay controls, generic replay substrate design, and broader dispute engine integration

### Requirement: Policy-driven replay controls page describes the first replay authorization slice
The `docs/architecture/policy-driven-replay-controls.md` page SHALL describe the first policy-driven replay controls slice for post-adjudication replay, including what currently ships and the current limits of the slice.

#### Scenario: Policy-driven replay controls page shows the bounded slice
- **WHEN** a user reads `docs/architecture/policy-driven-replay-controls.md`
- **THEN** they SHALL find sections describing the current replay-authorization slice, what currently ships, and current limits

### Requirement: P2P knowledge exchange track reflects landed policy-driven replay controls
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the policy-driven replay controls first slice as landed work and list the remaining work as richer policy classes, policy editing surfaces, per-transaction snapshots, and amount-tier replay controls.

#### Scenario: Track page points to the landed replay-policy slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find policy-driven replay controls described as a landed first slice
- **AND** the remaining work SHALL be described as richer policy classes, policy editing surfaces, per-transaction snapshots, and amount-tier replay controls

### Requirement: Dead-letter browsing / status observation page describes the first read-only visibility slice
The `docs/architecture/dead-letter-browsing-status-observation.md` page SHALL describe the first dead-letter browsing / status observation slice for post-adjudication execution, including what currently ships and the current limits of the slice.

#### Scenario: Dead-letter browsing / status observation page shows the bounded slice
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find sections describing the current read-only visibility slice, what currently ships, and current limits

#### Scenario: Dead-letter browsing page describes filtering and detail hints
- **WHEN** a user reads `docs/architecture/dead-letter-browsing-status-observation.md`
- **THEN** they SHALL find filtering and pagination described for the backlog list
- **AND** they SHALL find actor/time-based list filters described
- **AND** they SHALL find dead-letter reason and dispatch-reference filters described
- **AND** they SHALL find subtype/count filters and alternate sort modes described
- **AND** they SHALL find total retry-count and subtype-family filters described
- **AND** they SHALL find any-match family grouping described
- **AND** they SHALL find dominant family described
- **AND** they SHALL find transaction-global retry count and family grouping described
- **AND** they SHALL find transaction-global dominant family described
- **AND** they SHALL find compact per-submission breakdown described
- **AND** they SHALL find the optional detail-view raw background-task bridge described
- **AND** they SHALL find the cockpit dead-letter master-detail read surface described
- **AND** they SHALL find the thin cockpit filter bar described
- **AND** they SHALL find cockpit subtype filtering described
- **AND** they SHALL find cockpit actor/time filtering described
- **AND** they SHALL find cockpit latest-family filtering described
- **AND** they SHALL find cockpit any-match-family filtering described
- **AND** they SHALL find the cockpit detail-pane `Retry` action described
- **AND** they SHALL find inline confirm and success-refresh recovery UX described
- **AND** they SHALL find retry running/failure feedback described
- **AND** they SHALL find detail navigation hints described for per-transaction status

### Requirement: P2P knowledge exchange track reflects landed dead-letter browsing / status observation
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe dead-letter browsing / status observation as landed work with transaction-global dominant family, compact per-submission breakdown, a thin raw background-task bridge on the detail view, a cockpit dead-letter read surface, a thin cockpit filter bar, cockpit subtype filtering, cockpit latest-family filtering, cockpit any-match-family filtering, cockpit actor/time filtering, a cockpit `Retry` action, confirm/refresh recovery UX, and retry loading/failure feedback, and list the remaining work as richer cockpit filters beyond latest/any-match family and higher-level CLI surfaces.

#### Scenario: Track page points to the landed status slice
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** they SHALL find dead-letter browsing / status observation described as a landed first slice
- **AND** they SHALL find compact per-submission breakdown described as landed work
- **AND** they SHALL find the thin detail-view raw background-task bridge described as landed work
- **AND** they SHALL find the cockpit dead-letter read surface described as landed work
- **AND** they SHALL find the thin cockpit filter bar described as landed work
- **AND** they SHALL find cockpit subtype filtering described as landed work
- **AND** they SHALL find cockpit actor/time filtering described as landed work
- **AND** they SHALL find cockpit latest-family filtering described as landed work
- **AND** they SHALL find cockpit any-match-family filtering described as landed work
- **AND** they SHALL find the cockpit `Retry` action described as landed work
- **AND** they SHALL find confirm/refresh recovery UX described as landed work
- **AND** they SHALL find retry loading/failure feedback described as landed work
- **AND** the remaining work SHALL be described as richer cockpit filters beyond latest/any-match family and higher-level CLI surfaces
