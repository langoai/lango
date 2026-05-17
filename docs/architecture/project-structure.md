# Project Structure

This page documents every top-level directory and internal package in the Lango codebase.

## Top-Level Layout

```
lango/
├── cmd/lango/              # Application entry point
├── internal/               # All application packages (Go internal visibility)
├── prompts/                # Default prompt .md files (embedded via go:embed)
├── skills/                 # Skill system scaffold (go:embed)
├── openspec/               # Specifications (OpenSpec workflow)
├── docs/                   # Public documentation source
├── go.mod / go.sum         # Go module definition
└── zensical.toml           # Canonical Zensical documentation configuration
```

## `cmd/lango/`

The CLI entry point. Contains `main.go` which calls the root Cobra command defined in `internal/cli/`. Follows the Go convention of `os.Exit` only in `main()` -- all other code returns errors.

## `internal/`

All application code lives under `internal/` to enforce Go's visibility boundary. Packages are organized by domain, not by technical layer.

### Core Runtime

| Package | Description |
|---------|-------------|
| `adk/` | Google ADK v1.0.0 integration. Contains `Agent` (wraps ADK runner), `ModelAdapter` (bridges `provider.ProviderProxy` to ADK `model.LLM`), `ContextAwareModelAdapter` (injects knowledge/memory/RAG into system prompt), `SessionServiceAdapter` (bridges internal session store to ADK session interface), `ChildSessionServiceAdapter` (fork/merge child sessions for sub-agent isolation), `Summarizer` (extracts key results from child sessions), and `AdaptTool()` (converts `agent.Tool` to ADK `tool.Tool`) |
| `agent/` | Core agent types: `Tool` struct (name, description, parameters, handler), `ParameterDef`, `PII Redactor` (regex + optional Presidio integration), `SecretScanner` (prevents credential leakage in model output) |
| `app/` | Application bootstrap and wiring. `app.go` defines `New()` (component initialization), `Start()`, and `Stop()`. Wiring is split across domain-specific files (`wiring_*.go`) and module files (`modules_*.go`) that create subsystems such as knowledge, memory, graph, MCP, P2P, payment, smart account, economy, observability, automation, durable missions, transient proposals, loop-reader surfaces, and collaboration-reader surfaces. `types.go` defines the `App` struct with all component fields, including durable mission store/service handles, transient proposal registry/service handles, narrow loop readers for Mission Control (`LoopMissionReader`, `LoopProposalReader`, `LoopInquiryReader`, `LoopDeadLetterReader`, `LoopCronReader`), and narrow collaboration readers (`CollaborationMissionLinkReader`, `CollaborationAgentRunReader`, `CollaborationDelegationReader`, `CollaborationRuntimeReader`). The app layer also owns mission-aware adapters that connect approval, background, and RunLedger execution flows to the mission lifecycle without pushing mission imports downward, plus the mission-attributed runtime bridge used by the Wave 5 collaboration slice |
| `bootstrap/` | Pre-application startup: opens database, initializes crypto provider, loads config profile. Returns `bootstrap.Result` with shared `DBClient` and `Crypto` provider for reuse |
| `agentregistry/` | Agent definition registry. `Registry` loads built-in agents and user-defined `AGENT.md` files from `agent.agentsDir`. Provides `Specs()` for orchestrator routing and `Active()` for runtime agent listing |
| `agentmemory/` | Per-agent persistent memory. `Store` interface with `Save()`, `Get()`, `Search()`, `Delete()`, `Prune()` operations. Scoped by agent name for cross-session context retention |
| `automation/` | Shared automation contracts package. Defines reusable runner and channel-sender interfaces plus session-context channel detection so cron, background, and workflow subsystems share one automation-facing contract surface |
| `alerting/` | Operational alerting package. `Dispatcher` watches policy decisions, recovery retries, and circuit-breaker events within a sliding window and publishes alert events, while `DeliveryRouter` fans alerts out to configured channels such as webhooks with minimum-severity filtering |
| `approvalflow/` | Canonical artifact release approval-flow package. Evaluates artifact release decisions from exportability state, override requests, artifact-label match, and high-risk conditions, returning approval decisions plus fulfillment and settlement hints |
| `archtest/` | Architecture enforcement test package. Uses `go list` and ripgrep-based repository assertions to fail on boundary violations, raw bootstrap DB-handle leaks, unapproved storage wiring, or removed façade accessors in production packages |
| `dbopen/` | Managed database-opening helpers. `OpenManaged` opens the SQLite database in read-write mode, serializes Ent schema migration to avoid Atlas concurrency hazards, and returns the shared Ent/SQL handles; `OpenReadOnly` opens a read-only Ent client without migration after header and connection validation |
| `ctxkeys/` | Context key helpers. Propagates agent identity, durable `mission_id` bindings, dynamic tool allowlists, and spawn lineage through request contexts without creating import cycles |
| `deadline/` | Extendable deadline package. Provides idle-vs-hard-ceiling timeout resolution and an extendable deadline wrapper that auto-extends on activity without exceeding a maximum absolute timeout |
| `mission/` | Durable mission lifecycle package. `Store` persists latest mission rows, append-only mission state history, and `MissionExecutionLink` records. `Service` owns durable mission creation, proposal acceptance, coarse decision/blocker transitions, execution-link attachment, and mission lookup by execution |
| `proposal/` | Transient proactive proposal package. `Registry` keeps session-scoped proposal state in memory, `DeterministicPreparer` builds source-native prepared briefs from learning-suggestion evidence, and `Service` owns proposal upsert, prepare, dismiss, accept, restore, and expiration through that transient registry |
| `loopview/` | Deterministic operator-loop projection package. `Projector` derives `LoopView` and `AgendaView` rows from real existing sources only: durable missions, pending inquiries, dead-letter backlog, cron jobs, and deterministic follow-up predicates. The current slice does not add a durable loop table and does not imply calendar, inbox, workflow-run, or external task-system integrations |
| `collabview/` | Deterministic mission-collaboration projection package for the Wave 5 local coworking slice. `Projector` derives compact mission-linked collaboration summaries from linked local execution data: participants, active owner, recent handoffs, blocked-on-approval or waiting-on-teammate state, recovery hints, budget pressure, and linked local review state. The package is projection-only and does not create a durable collaboration table |
| `exportability/` | Source-class exportability policy evaluator. `Evaluate()` returns an exportability receipt with stage, state, policy code, explanation, and source-lineage summaries derived from source classes such as public, user-exportable, and private-confidential |
| `knowledgeruntime/` | Knowledge-exchange runtime service. Opens canonical knowledge-exchange transaction receipts, verifies payment-approval state on the current submission receipt, selects the execution branch (`prepay` or `escrow`), and advances runtime status through the receipts store |
| `receipts/` | Canonical in-memory submission/transaction receipt store. Owns submission receipt creation, knowledge-exchange transaction opening, approval and settlement/runtime progression, external transaction binding, and append-only receipt events used by downstream runtime services |
| `finance/` | Shared monetary leaf utilities for USDC operations. Provides parsing/formatting helpers, micro-unit conversion, and quote-related types without depending on wallet or execution packages |
| `paymentapproval/` | Upfront-payment policy evaluator. Classifies amounts and trust context, enforces user max-prepay budget policy, and emits approve/reject/escalate outcomes with suggested settlement mode hints (`prepay`, `escrow`, or `escalate`) |
| `paymentgate/` | Direct-payment eligibility gate over canonical receipts. Verifies the current submission binding, payment approval status, and settlement hint before allowing direct settlement execution |
| `settlementprogression/` | Canonical settlement progression mapper. Translates artifact release outcomes into settlement progression states and dispute escalation transitions on top of the receipts store |
| `settlementexecution/` | Direct-payment settlement executor. Resolves final USDC amount from canonical price context, calls the direct-payment runtime, records failures, and marks canonical settlement closeout on success |
| `partialsettlementexecution/` | Partial direct-payment settlement executor. Parses partial-settlement hints, computes executed vs. remaining amount, records execution evidence, and marks partially settled closeout in canonical receipts |
| `escrowexecution/` | Escrow create/fund runtime bridge for escrow-recommended transactions. Requires approved payment state and bound escrow execution input, then records pending/created/funded progress on the canonical receipt |
| `disputehold/` | Dispute-hold executor for funded escrow transactions. Requires dispute-ready settlement state and escrow reference, invokes a hold runtime, and records hold evidence or failure against the canonical receipt |
| `escrowadjudication/` | Canonical escrow adjudication applier. Requires dispute-hold evidence on the current submission receipt, validates release/refund outcomes, and persists adjudication state through the receipts store |
| `escrowrelease/` | Escrow release executor for funded, release-adjudicated transactions. Resolves amount from canonical price context, calls the release runtime, and records settled closeout or failure |
| `escrowrefund/` | Escrow refund executor for funded, refund-adjudicated transactions. Resolves amount from canonical price context, calls the refund runtime, and records refund evidence or failure |
| `postadjudicationreplay/` | Manual post-adjudication replay dispatcher. Re-hydrates canonical adjudication snapshots from receipts, requires dead-letter evidence and actor policy permission, records manual retry intent, and dispatches background replay work |
| `postadjudicationstatus/` | Dead-letter and retry-status projection over adjudicated transactions. Builds current dead-letter backlog entries, canonical transaction status views, submission breakdown, and latest background retry linkage from receipt history |
| `storagebroker/` | Persistent stdio JSON broker protocol for encrypted storage operations. Defines request/response envelopes and typed payload contracts for DB status/open, payload encryption, config profile load/save/list, session CRUD, recall, learning, alerts, and payment history/usage flows |
| `streamx/` | Generic iterator-based stream combinator package. Defines typed `Stream[T]` and source-tagged events used by merge/race/fan-in/drain style helpers with context-aware cancellation semantics |
| `tooloutput/` | TTL-backed in-memory tool output store. Returns UUID references for stored tool output and supports full retrieval, ranged line reads, regex grep, and lifecycle-managed expiration |
| `toolparam/` | Typed dynamic tool parameter extraction helpers. Provides required/optional string, int, bool, float64, and string-slice accessors plus a structured missing-parameter error |
| `agentrt/` | Agent runtime control-plane package. Wraps the shared turn executor with delegation guard, observational budget policy, capability policy/runtime, recovery policy, task/control tools, and run-projection/store helpers without becoming a separate execution engine |
| `gatekeeper/` | Response sanitization package. `Sanitizer` strips thought tags, internal markers, large raw JSON blocks, and configured custom patterns while preserving fenced code blocks |
| `retrieval/` | Retrieval orchestration package. `RetrievalCoordinator` runs fact and temporal search agents in parallel, merges and reranks findings with authority/version/recency priority, and truncates by token budget before converting to context-layer results |
| `search/` | Domain-agnostic FTS5 search substrate. `FTS5Index` manages virtual table lifecycle and CRUD/bulk insert operations over raw row IDs and columns, while `ProbeFTS5` verifies SQLite FTS5 availability |
| `turnrunner/` | Shared turn execution runner. Owns timeout and stale-stream handling, durable trace recording, chunk/tool/delegation/thinking callbacks, retry/recovery loop integration, and final outcome classification for a single turn |
| `turntrace/` | Durable turn trace package. Defines trace/event models, append-only Ent-backed persistence, failure and retention queries, delegation/event taxonomy, and per-agent metrics summaries derived from traces/events |
| `lineio/` | Shared single-line reader helper. Preserves `bufio.Reader.ReadString('\n')` semantics, including partial-line plus EOF behavior |
| `llm/` | Minimal LLM abstraction package. Defines the provider-agnostic `TextGenerator` interface so callers can request generated text from system/user prompt pairs without coupling to a concrete provider |
| `storeutil/` | Small store-facing utility helpers. Provides generic slice/map copy helpers and JSON marshal/unmarshal wrappers so persistence layers can copy state safely and surface contextual serialization errors |
| `ontology/` | Ontology governance and tooling package. Provides schema registry and Ent-backed stores, ACL policy, action registry/executor with compensation logging, property/truth maintenance, P2P source attribution, and higher-level ontology service helpers |
| `sqlitedriver/` | Shared SQLite driver helper package. Centralizes path expansion, DB open/configuration, file-header validation, and connection setup used by managed and read-only database-opening flows |
| `storage/` | Storage facade and broker-adapter package. `Facade` composes config profiles, security state, session/provenance/run/cron/turntrace/ontology/payment stores plus runtime readers, while broker-backed adapters bridge storagebroker APIs into those persistence interfaces |
| `eventbus/` | Typed synchronous event pub/sub. `Bus` with `Subscribe()` / `Publish()`. `SubscribeTyped[T]()` generic helper for type-safe subscriptions. Events: ContentSaved, TriplesExtracted, TurnCompleted, ReputationChanged, TokenUsageEvent |
| `types/` | Shared type definitions used across packages: `ProviderType`, `Role`, `RPCSenderFunc`, `ChannelType`, `ConfidenceLevel`, `TokenUsage` |

### Presentation

| Package | Description |
|---------|-------------|
| `cli/` | Root Cobra command and subcommand packages |
| `cli/agent/` | `lango agent status`, `list`, `tools`, `hooks`, `trace list/show/metrics`, `graph` -- agent runtime inspection and diagnostics |
| `cli/a2a/` | `lango a2a card`, `lango a2a check` -- A2A protocol configuration inspection |
| `cli/approval/` | `lango approval status` -- tool approval policy and provider inspection |
| `cli/alerts/` | `lango alerts list`, `summary` -- operational alert inspection |
| `cli/bg/` | `lango bg list`, `status`, `cancel`, `result` -- background task management |
| `cli/cliboot/` | Shared bootstrap loaders that run application bootstrap once and expose reusable `BootResult` / `Config` callbacks for gateway-backed CLI commands |
| `cli/cliexit/` | Structured CLI exit-code errors returned from command packages to `cmd/lango` so process termination stays in the binary entrypoint |
| `cli/clihttp/` | Shared HTTP/JSON helpers for gateway-backed CLI commands, including bounded JSON fetches, `table|json` output validation, and common pretty-JSON rendering |
| `cli/chat/` | `lango chat` -- focused chat TUI |
| `cli/clitypes/` | Shared CLI type definitions (`ProviderMetadata` for provider display) |
| `cli/cockpit/` | Explicit `lango cockpit` multi-panel operator dashboard. Owns Mission Control page rendering, shared pending approval ownership, activity buffers, compatibility learning-buffer fallback, transient proposal rendering, durable-first mission projection, deterministic agenda/loop projection, compact mission collaboration rendering, and the sidebar/detail-page shell around the shared chat model. In the current loop slice, dead-letter and cron loops are projected as operator-global rows, while mission, inquiry, and follow-up loops remain session-scoped. In the current collaboration slice, Mission Control shows mission-linked local coworking only; external P2P team UX remains secondary and is not part of the primary collaboration surface |
| `cli/workbench/` | Bare `lango` standalone mission workbench shell. Mounts Mission Control content directly without the full cockpit sidebar/context chrome while reusing the shared chat model, pending approval path, learning/activity buffers, and mission-control runtime subscriptions |
| `cli/workbenchstart/` | Context-aware starter, post-turn, and recovery prompt builders for the bare `lango` workbench. Inspects the workspace root, git branch/dirty state, and changed top-level targets to suggest the best next prompt |
| `cli/configcmd/` | `lango config list`, `create`, `use`, `delete`, `import`, `export`, `get`, `set`, `keys`, `validate` -- encrypted profile and configuration management |
| `cli/contract/` | `lango contract read`, `call`, `abi load` -- smart contract interaction |
| `cli/cron/` | `lango cron add`, `list`, `delete`, `pause`, `resume`, `history` -- cron job management |
| `cli/doctor/` | `lango doctor` -- system diagnostics and health checks |
| `cli/economy/` | `lango economy budget status`, `risk status`, `pricing status`, `negotiate status`, `escrow status/list/show/sentinel status` -- P2P economy management |
| `cli/extension/` | `lango extension inspect/install/list/remove` -- extension pack management |
| `cli/graph/` | `lango graph status`, `query`, `stats`, `clear`, `add`, `export`, `import` -- graph store management |
| `cli/learning/` | `lango learning status`, `history` -- learning and knowledge inspection |
| `cli/librarian/` | `lango librarian status`, `inquiries` -- proactive knowledge librarian inspection |
| `cli/mcp/` | `lango mcp list`, `add`, `remove`, `get`, `test`, `enable`, `disable` -- MCP server management |
| `cli/memory/` | `lango memory list`, `status`, `clear`, `agents`, `agent <name>` -- observational and per-agent memory management |
| `cli/metrics/` | `lango metrics`, `sessions`, `tools`, `agents`, `policy`, `history` -- system observability metrics |
| `cli/onboard/` | `lango onboard` -- 5-step guided setup wizard |
| `cli/p2p/` | `lango p2p status`, `peers`, `connect`, `disconnect`, `firewall list/add/remove`, `discover`, `identity`, `reputation`, `pricing`, `session list/revoke/revoke-all`, `sandbox status/test/cleanup`, `workspace create/list/status/join/leave`, `git init/log/diff/push/fetch`, `provenance push/fetch`, `team list/status/disband`, `zkp status/circuits` -- P2P network management |
| `cli/payment/` | `lango payment balance`, `history`, `limits`, `info`, `send`, `x402` -- payment operations |
| `cli/prompt/` | Interactive prompt utilities for CLI input |
| `cli/provenance/` | `lango provenance status/checkpoint list/create/show/session tree/list/attribution show/report/bundle export/import` -- session provenance and attribution management |
| `cli/run/` | `lango run list/status/journal <run-id>` -- RunLedger inspection |
| `cli/sandbox/` | `lango sandbox status/test` -- OS-level sandbox inspection |
| `cli/security/` | `lango security status`, `change-passphrase`, deprecated `migrate-passphrase`, `secrets`, `keyring store/clear/status`, `recovery setup/restore`, `kms status/test/keys/wrap/detach` plus legacy `db-migrate`/`db-decrypt` tombstones -- security operations |
| `cli/settings/` | `lango settings` -- full configuration editor |
| `cli/smartaccount/` | `lango account info`, `deploy`, `session list/create/revoke`, `module list/install`, `policy show/set`, `paymaster status/approve` -- ERC-7579 smart account management |
| `cli/status/` | `lango status`, `dead-letter-summary`, `dead-letters`, `dead-letter`, `dead-letter retry` -- unified status and dead-letter inspection |
| `cli/tuicore/` | Shared TUI components for interactive terminal sessions. `FormModel` (Bubbletea form manager), `Field` struct with input types: `InputText`, `InputInt`, `InputPassword`, `InputBool`, `InputSelect`, `InputSearchSelect` |
| `cli/tui/` | TUI styling and banner components for interactive terminal sessions |
| `cli/workflow/` | `lango workflow run`, `list`, `status`, `cancel`, `history`, `validate <file>` -- workflow management |
| `channels/` | Channel bot integrations for Telegram, Discord, and Slack. Each adapter converts platform-specific messages to the Gateway's internal format |
| `gateway/` | HTTP REST + WebSocket server built on chi router. Handles JSON-RPC over WebSocket, OIDC authentication (`AuthManager`), turn callbacks, and approval routing. Provides `Server.SetAgent()` for late-binding the agent after initialization |

### Intelligence

| Package | Description |
|---------|-------------|
| `knowledge/` | Ent-backed knowledge store. `ContextRetriever` implements 8-layer retrieval: runtime context, tool registry, user knowledge, skill patterns, external knowledge, agent learnings, pending inquiries, and conversation analysis. Exposes `SetEmbedCallback` and `SetGraphCallback` for async processing |
| `learning/` | Self-learning engine. `Engine` extracts patterns from tool execution results. `GraphEngine` extends `Engine` with graph triple generation and confidence propagation (rate 0.3). `ConversationAnalyzer` and `SessionLearner` analyze conversation history. `AnalysisBuffer` batches analysis with turn/token thresholds |
| `memory/` | Observational memory system. `Observer` extracts observations from conversation turns, `Reflector` synthesizes higher-level reflections, `Buffer` manages async processing with configurable token thresholds. `GraphHooks` generates temporal/session triples for the graph store. Supports compaction via `SetCompactor()` |
| `embedding/` | Multi-provider embedding pipeline. `Registry` manages providers (OpenAI, Google, local). `SQLiteVecStore` stores vectors. `EmbeddingBuffer` batches embed requests asynchronously. `RAGService` performs semantic retrieval with collection/distance filtering. `StoreResolver` resolves source IDs back to knowledge/memory content |
| `graph/` | BoltDB-backed triple store with SPO/POS/OSP indexes for efficient traversal. `Extractor` uses LLM to extract entities and relations from text. `GraphBuffer` batches triple insertions. `GraphRAGService` implements 2-phase hybrid retrieval (vector search + graph expansion) |
| `librarian/` | Proactive knowledge extraction. `ObservationAnalyzer` identifies knowledge gaps from conversation observations. `InquiryProcessor` generates questions and resolves them. `InquiryStore` persists pending inquiries. `ProactiveBuffer` manages the async pipeline with configurable thresholds |
| `skill/` | File-based skill system. `FileSkillStore` manages skill files on disk. `Registry` loads skills and converts active skills to `agent.Tool` instances. Skill infrastructure (FileSkillStore, Registry, GitHub importer) supports user-defined skills |

### MCP Integration

| Package | Description |
|---------|-------------|
| `mcp/` | MCP (Model Context Protocol) client integration. `ServerConnection` manages individual server lifecycles (stdio, HTTP streamable, SSE transports). `ServerManager` coordinates multiple server connections. `AdaptTools()` converts discovered MCP tools to `agent.Tool` instances using the `mcp__{serverName}__{toolName}` naming convention. Multi-scope config: profile < user (`~/.lango/mcp.json`) < project (`.lango-mcp.json`). Built on `github.com/modelcontextprotocol/go-sdk` |

### Blockchain and Smart Accounts

| Package | Description |
|---------|-------------|
| `contract/` | Generic EVM smart contract interaction. `Caller` provides `Read()` for view/pure calls and `Write()` for state-changing transactions with EIP-1559 gas pricing, nonce management, and retry logic. `ABICache` caches parsed ABI definitions |
| `smartaccount/` | ERC-7579 modular smart account management with ERC-4337 UserOp submission. `Manager` handles Safe-based account deployment and execution. Sub-packages: `bindings/` (contract ABI bindings for Safe7579, session validator, spending hook, escrow executor), `bundler/` (external bundler RPC client), `module/` (ERC-7579 module registry and ABI encoding), `paymaster/` (Alchemy, Pimlico, Circle paymaster integrations with approval and recovery), `policy/` (off-chain policy engine for session key validation), `session/` (hierarchical session key lifecycle with crypto derivation) |

### P2P Economy

| Package | Description |
|---------|-------------|
| `economy/escrow/` | Milestone-based escrow engine for P2P transactions. `Engine` manages the escrow lifecycle (Pending/Funded/Active/Completed/Released/Disputed/Expired/Refunded). `SettlementExecutor` interface for fund lock/release/refund. `sentinel/` sub-package provides fraud detection and session guard. `hub/` sub-package provides on-chain escrow vault interaction |
| `economy/pricing/` | Dynamic pricing engine with rule-based evaluation. `Engine` computes per-tool prices using base prices, reputation-weighted adjustments, and configurable rule sets. Quote expiry support |
| `economy/negotiation/` | Multi-round price negotiation between peers. `Engine` manages negotiation sessions with turn-based protocol, strategy interface, and configurable round limits |
| `economy/risk/` | Risk assessment engine using a 3-variable matrix (trust score x transaction value x output verifiability). `Assessor` interface with policy adapter integration |
| `economy/budget/` | Task-scoped budget management. `Guard` interface enforces spending limits. `Engine` tracks allocations with alert callbacks. On-chain budget verification support |

### P2P Network

| Package | Description |
|---------|-------------|
| `p2p/` | Core P2P node management. `Node` struct handles libp2p host lifecycle and node key management |
| `p2p/identity/` | DID-based peer identity management |
| `p2p/discovery/` | Peer discovery via libp2p Kademlia DHT and gossipsub. `GossipDiscovery` for pub/sub-based peer announcements with credential revocation. `AdService` manages structured agent advertisements (Context Flyer) via DHT provider records |
| `p2p/handshake/` | Authenticated handshake protocol with signed challenges (ECDSA), timestamp validation, nonce replay protection, and session management. Dual protocol support (v1.0/v1.1) |
| `p2p/firewall/` | Inbound request firewall with rule-based filtering. `OwnerShield` restricts tool access. ZK attestation verification support |
| `p2p/protocol/` | P2P message protocol. `Handler` processes inbound tool invocations with sandbox execution and security event tracking. `RemoteAgent` wraps remote peer tool invocation. Team message handling |
| `p2p/reputation/` | Peer reputation tracking. `Store` records interaction outcomes and computes trust scores with change notification callbacks |
| `p2p/zkp/` | Zero-knowledge proof system. `ProverService` with gnark circuits for attestation, capability, identity, and reputation proofs (BN254, plonk+groth16) |
| `p2p/agentpool/` | P2P agent pool with health monitoring. `Pool` manages discovered agents. `HealthChecker` runs periodic probes (Healthy/Degraded/Unhealthy/Unknown). `Selector` provides weighted agent selection based on reputation, latency, success rate, and availability |
| `p2p/team/` | P2P team coordination. `Team` manages task-scoped agent groups with roles (Leader, Worker, Reviewer, Observer). `ScopedContext` controls metadata sharing. Budget tracking via `AddSpend()`. Team lifecycle: Forming -> Active -> Completed/Disbanded |
| `p2p/settlement/` | On-chain USDC settlement for P2P tool invocations. `Service` handles EIP-3009 authorization-based transfers with exponential retry. `ReputationRecorder` interface for outcome tracking. Subscriber pattern for settlement notifications |
| `p2p/paygate/` | Payment gate between firewall and tool executor. Verifies EIP-3009 payment authorizations, checks tool pricing, and enforces payment requirements before tool execution |

### Observability

| Package | Description |
|---------|-------------|
| `observability/` | System metrics aggregation. `MetricsCollector` performs thread-safe in-memory collection of token usage, tool executions, agent metrics, and session metrics. `SystemSnapshot` provides point-in-time summaries |
| `observability/token/` | Token usage tracking. `Tracker` subscribes to `TokenUsageEvent` on the event bus and forwards data to the `MetricsCollector` and optional persistent `TokenStore` |
| `observability/health/` | Health checking framework. `Registry` manages `Checker` instances and runs aggregate health assessments. Component-level status: Healthy/Degraded/Unhealthy |
| `observability/audit/` | Audit log recording. `Recorder` subscribes to tool execution and token usage events on the event bus and writes entries to the Ent-backed `AuditLog` schema |

### Infrastructure

| Package | Description |
|---------|-------------|
| `config/` | YAML configuration loading with environment variable substitution (`${ENV_VAR}` syntax), validation, and defaults. Defines all config structs (`Config`, `AgentConfig`, `SecurityConfig`, `MCPConfig`, `DynamicPricingConfig`, `RiskConfig`, `BudgetConfig`, etc.) |
| `configstore/` | Encrypted configuration profile storage backed by Ent ORM. Allows multiple named profiles with passphrase-derived encryption |
| `security/` | Crypto providers (`LocalProvider` with passphrase-derived keys, `RPCProvider` for remote signing). `KeyRegistry` manages encryption keys. `SecretsStore` provides encrypted secret storage. `RefStore` holds opaque references so plaintext never reaches agent context. Companion discovery for distributed setups. KMS providers (AWS KMS, GCP KMS, Azure Key Vault, PKCS#11) with retry and health checking |
| `session/` | Session persistence via Ent ORM with SQLite backend. `EntStore` implements the `Store` interface with configurable TTL and max history turns. `CompactMessages()` supports memory compaction |
| `ent/` | Ent ORM schema definitions and generated code for all database entities |
| `logging/` | Structured logging via Zap. Per-package logger instances (`logging.App()`, `logging.Agent()`, `logging.Gateway()`, etc.) |
| `provider/` | Unified AI provider interface. `GenerateParams`, `StreamEvent`, streaming via `iter.Seq2`. Implementations in sub-packages |
| `provider/anthropic/` | Anthropic Claude provider |
| `provider/gemini/` | Google Gemini provider |
| `provider/openai/` | OpenAI-compatible provider (GPT, Ollama, and other OpenAI API-compatible services) |
| `supervisor/` | `Supervisor` manages provider credentials and configuration. `ProviderProxy` handles model routing with temperature, max tokens, and fallback provider chains |
| `prompt/` | Structured prompt builder. `Builder` assembles system prompts from prioritized `Section` instances. `LoadFromDir()` loads custom prompts from user directories. Sections: Identity, Safety, ConversationRules, ToolUsage, Automation, AgentIdentity |
| `approval/` | Tool execution approval system. `CompositeProvider` routes approval requests to channel-specific providers. `GatewayProvider` sends approval requests over WebSocket. `TTYProvider` prompts in terminal. `HeadlessProvider` auto-approves. `GrantStore` caches approval decisions. `ApprovalRequest` now carries optional mission and execution attribution fields so mission-aware flows can record coarse durable decision state without changing provider contracts |
| `payment/` | Blockchain payment service. `TxBuilder` constructs USDC transfer transactions. `Service` coordinates wallet, spending limiter, and transaction execution through an explicit payment transaction store |
| `wallet/` | Wallet providers: `LocalWallet` (derives keys from secrets store), `RPCWallet` (remote signing), `CompositeWallet` (fallback chain). `EntSpendingLimiter` / store-backed limiters enforce per-transaction and daily spending limits |
| `x402/` | X402 V2 payment protocol implementation. `Interceptor` handles automatic payment for 402 responses. `LocalSignerProvider` derives signing keys from secrets store. EIP-3009 signing for gasless USDC transfers |
| `cron/` | Cron scheduling system built on robfig/cron/v3. `Scheduler` manages job lifecycle. `EntStore` persists jobs and execution history. `Executor` runs agent prompts on schedule. `Delivery` routes results to channels |
| `background/` | In-memory background task manager. `Manager` enforces concurrency limits and task timeouts. `Notification` routes results to channels. `bg_submit` can call an app-supplied mission execution linker so new background work attaches to an existing durable mission at creation time |
| `runledger/` | Durable execution engine with append-only journal and PEV validation. `run_create` can call an app-supplied mission execution linker so durable mission-to-run relationships are recorded at execution creation time |
| `workflow/` | DAG-based workflow engine. `Engine` parses YAML workflow definitions, resolves step dependencies, and executes steps in parallel where possible. `StateStore` persists workflow state via Ent |
| `lifecycle/` | Component lifecycle management. `Registry` with priority-ordered startup and reverse-order shutdown. Adapters: `SimpleComponent`, `FuncComponent`, `ErrorComponent` |
| `keyring/` | Hardware keyring integration (Touch ID / TPM 2.0). `Provider` interface backed by OS keyring via go-keyring |
| `sandbox/` | Tool execution isolation. `SubprocessExecutor` for process-isolated P2P tool execution. `ContainerRuntime` interface with Docker/gVisor/native fallback chain. Optional pre-warmed container pool |
| `dbmigrate/` | Legacy database migration tombstones and remediation helpers for old SQLCipher installs |
| `toolcatalog/` | Thread-safe tool registry with category grouping. `Catalog` with `Register()`, `Get()`, `ListCategories()`, `ListTools()`. `ToolEntry` pairs tools with categories, `ToolSchema` provides tool summaries |
| `toolchain/` | HTTP-style middleware chain for tool wrapping. `Middleware` type, `Chain()` / `ChainAll()` functions. Built-in middlewares: security filter, access control, event publishing, knowledge save, approval, browser recovery. The approval middleware exposes a lower-layer observer seam so the app layer can drive mission `waiting_decision` / `active` updates without importing mission code into `toolchain` |
| `appinit/` | Declarative module initialization system. `Module` interface with `Provides` / `DependsOn` keys. `Builder` with Kahn's algorithm topological sort for dependency resolution. Foundation for ordered application bootstrap |
| `asyncbuf/` | Generic async batch processor. `BatchBuffer[T]` with configurable batch size, flush interval, and backpressure. `Start()` / `Enqueue()` / `Stop()` lifecycle. Replaces per-subsystem buffer implementations |
| `security/passphrase/` | Passphrase prompt and validation helpers for terminal input |
| `mdparse/` | Shared markdown parsing utilities. `SplitFrontmatter()` extracts YAML frontmatter and body from markdown content |
| `testutil/` | Shared test utilities and mock implementations. `TestEntClient()` (in-memory Ent client), `NopLogger()`, and mock types for crypto, embedding, graph, session, cron, and provider interfaces |
| `orchestration/` | Multi-agent orchestration. `BuildAgentTree()` creates an ADK agent hierarchy. `AgentSpec` defines agent metadata (prefixes, keywords, capabilities). `PartitionToolsDynamic()` allocates tools to agents via multi-signal matching (prefix, keyword, capability). `BuiltinSpecs()` returns default agent definitions. Sub-agents: Operator, Navigator, Vault, Librarian, Automator, Planner, Chronicler. Supports user-defined agents via `AgentRegistry` |
| `a2a/` | Agent-to-Agent protocol. `Server` exposes agent card and task endpoints. `LoadRemoteAgents()` discovers and loads remote agent capabilities |
| `tools/` | Built-in tool implementations |
| `tools/browser/` | Headless browser tool with session management |
| `tools/crypto/` | Cryptographic operation tools (encrypt, decrypt, sign, verify) |
| `tools/exec/` | Shell command execution tool |
| `tools/filesystem/` | File read/write/list tools with path allowlisting and blocklisting |
| `tools/secrets/` | Secret management tools (store, retrieve, list, delete) |
| `tools/payment/` | Payment tools (balance, send, history) |

## `prompts/`

Default system prompt sections as Markdown files, embedded into the binary via `go:embed`. The prompt builder loads these as the default sections, which can be overridden by placing custom `.md` files in a user-specified prompts directory.

## `skills/`

Skill system scaffold. The skill infrastructure (FileSkillStore, Registry, GitHub importer) remains fully functional for user-defined skills. Built-in embedded skills were removed because Lango's passphrase-protected security model makes it impractical for the agent to invoke lango CLI commands as skills.

## `openspec/`

Specification documents following the OpenSpec workflow. Used for tracking feature specifications, changes, and architectural decisions.
