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
├── docs/                   # MkDocs documentation source
├── go.mod / go.sum         # Go module definition
└── mkdocs.yml              # MkDocs configuration
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
| `app/` | Application bootstrap and wiring. `app.go` defines `New()` (component initialization), `Start()`, and `Stop()`. Wiring is split across domain-specific files (`wiring_*.go`) that create individual subsystems (knowledge, memory, embedding, graph, MCP, P2P, payment, smart account, economy, observability, automation, contract). `types.go` defines the `App` struct with all component fields. `tools.go` builds tool collections. `sender.go` provides `channelSender` adapter for delivery |
| `bootstrap/` | Pre-application startup: opens database, initializes crypto provider, loads config profile. Returns `bootstrap.Result` with shared `DBClient` and `Crypto` provider for reuse |
| `agentregistry/` | Agent definition registry. `Registry` loads built-in agents and user-defined `AGENT.md` files from `agent.agentsDir`. Provides `Specs()` for orchestrator routing and `Active()` for runtime agent listing |
| `agentmemory/` | Per-agent persistent memory. `Store` interface with `Save()`, `Get()`, `Search()`, `Delete()`, `Prune()` operations. Scoped by agent name for cross-session context retention |
| `ctxkeys/` | Context key helpers. `WithAgentName()` / `AgentNameFromContext()` for propagating agent identity through request contexts |
| `eventbus/` | Typed synchronous event pub/sub. `Bus` with `Subscribe()` / `Publish()`. `SubscribeTyped[T]()` generic helper for type-safe subscriptions. Events: ContentSaved, TriplesExtracted, TurnCompleted, ReputationChanged, TokenUsageEvent |
| `types/` | Shared type definitions used across packages: `ProviderType`, `Role`, `RPCSenderFunc`, `ChannelType`, `ConfidenceLevel`, `TokenUsage` |

### Presentation

| Package | Description |
|---------|-------------|
| `cli/` | Root Cobra command and subcommand packages |
| `cli/agent/` | `lango agent status`, `lango agent list` -- agent runtime inspection |
| `cli/a2a/` | `lango a2a card`, `lango a2a check` -- A2A protocol configuration inspection |
| `cli/approval/` | `lango approval status` -- tool approval policy and provider inspection |
| `cli/bg/` | `lango bg list`, `status`, `cancel`, `result` -- background task management |
| `cli/clitypes/` | Shared CLI type definitions (`ProviderMetadata` for provider display) |
| `cli/contract/` | `lango contract read`, `call`, `abi load` -- smart contract interaction |
| `cli/cron/` | `lango cron add`, `list`, `delete`, `pause`, `resume`, `history` -- cron job management |
| `cli/doctor/` | `lango doctor` -- system diagnostics and health checks |
| `cli/economy/` | `lango economy budget`, `risk`, `pricing`, `negotiate`, `escrow` -- P2P economy management |
| `cli/graph/` | `lango graph status`, `query`, `stats`, `clear` -- graph store management |
| `cli/learning/` | `lango learning status`, `history` -- learning and knowledge inspection |
| `cli/librarian/` | `lango librarian status`, `inquiries` -- proactive knowledge librarian inspection |
| `cli/mcp/` | `lango mcp list`, `add`, `remove`, `get`, `test`, `enable`, `disable` -- MCP server management |
| `cli/memory/` | `lango memory list`, `status`, `clear` -- observational memory management |
| `cli/metrics/` | `lango metrics`, `sessions`, `tools`, `agents`, `history` -- system observability metrics |
| `cli/onboard/` | `lango onboard` -- 5-step guided setup wizard |
| `cli/p2p/` | `lango p2p status`, `peers`, `connect`, `disconnect`, `firewall list/add/remove`, `discover`, `identity`, `reputation`, `pricing`, `session list/revoke/revoke-all`, `sandbox status/test/cleanup` -- P2P network management |
| `cli/payment/` | `lango payment balance`, `history`, `limits`, `info`, `send` -- payment operations |
| `cli/prompt/` | Interactive prompt utilities for CLI input |
| `cli/security/` | `lango security status`, `secrets`, `migrate-passphrase`, `keyring store/clear/status`, `db-migrate`, `db-decrypt`, `kms status/test/keys` -- security operations |
| `cli/settings/` | `lango settings` -- full configuration editor |
| `cli/smartaccount/` | `lango account info`, `deploy`, `session list`, `module list`, `policy show`, `paymaster` -- ERC-7579 smart account management |
| `cli/tuicore/` | Shared TUI components for interactive terminal sessions. `FormModel` (Bubbletea form manager), `Field` struct with input types: `InputText`, `InputInt`, `InputPassword`, `InputBool`, `InputSelect`, `InputSearchSelect` |
| `cli/tui/` | TUI styling and banner components for interactive terminal sessions |
| `cli/workflow/` | `lango workflow run`, `list`, `status`, `cancel`, `history` -- workflow management |
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
| `approval/` | Tool execution approval system. `CompositeProvider` routes approval requests to channel-specific providers. `GatewayProvider` sends approval requests over WebSocket. `TTYProvider` prompts in terminal. `HeadlessProvider` auto-approves. `GrantStore` caches approval decisions |
| `payment/` | Blockchain payment service. `TxBuilder` constructs USDC transfer transactions. `Service` coordinates wallet, spending limiter, and transaction execution |
| `wallet/` | Wallet providers: `LocalWallet` (derives keys from secrets store), `RPCWallet` (remote signing), `CompositeWallet` (fallback chain). `EntSpendingLimiter` enforces per-transaction and daily spending limits |
| `x402/` | X402 V2 payment protocol implementation. `Interceptor` handles automatic payment for 402 responses. `LocalSignerProvider` derives signing keys from secrets store. EIP-3009 signing for gasless USDC transfers |
| `cron/` | Cron scheduling system built on robfig/cron/v3. `Scheduler` manages job lifecycle. `EntStore` persists jobs and execution history. `Executor` runs agent prompts on schedule. `Delivery` routes results to channels |
| `background/` | In-memory background task manager. `Manager` enforces concurrency limits and task timeouts. `Notification` routes results to channels |
| `workflow/` | DAG-based workflow engine. `Engine` parses YAML workflow definitions, resolves step dependencies, and executes steps in parallel where possible. `StateStore` persists workflow state via Ent |
| `lifecycle/` | Component lifecycle management. `Registry` with priority-ordered startup and reverse-order shutdown. Adapters: `SimpleComponent`, `FuncComponent`, `ErrorComponent` |
| `keyring/` | Hardware keyring integration (Touch ID / TPM 2.0). `Provider` interface backed by OS keyring via go-keyring |
| `sandbox/` | Tool execution isolation. `SubprocessExecutor` for process-isolated P2P tool execution. `ContainerRuntime` interface with Docker/gVisor/native fallback chain. Optional pre-warmed container pool |
| `dbmigrate/` | Database encryption migration. `MigrateToEncrypted` / `DecryptToPlaintext` for SQLCipher transitions. `IsEncrypted` detection and `secureDeleteFile` cleanup |
| `toolcatalog/` | Thread-safe tool registry with category grouping. `Catalog` with `Register()`, `Get()`, `ListCategories()`, `ListTools()`. `ToolEntry` pairs tools with categories, `ToolSchema` provides tool summaries |
| `toolchain/` | HTTP-style middleware chain for tool wrapping. `Middleware` type, `Chain()` / `ChainAll()` functions. Built-in middlewares: security filter, access control, event publishing, knowledge save, approval, browser recovery |
| `appinit/` | Declarative module initialization system. `Module` interface with `Provides` / `DependsOn` keys. `Builder` with Kahn's algorithm topological sort for dependency resolution. Foundation for ordered application bootstrap |
| `asyncbuf/` | Generic async batch processor. `BatchBuffer[T]` with configurable batch size, flush interval, and backpressure. `Start()` / `Enqueue()` / `Stop()` lifecycle. Replaces per-subsystem buffer implementations |
| `passphrase/` | Passphrase prompt and validation helpers for terminal input |
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
