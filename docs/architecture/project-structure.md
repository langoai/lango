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
| `adk/` | Google ADK v0.4.0 integration. Contains `Agent` (wraps ADK runner), `ModelAdapter` (bridges `provider.ProviderProxy` to ADK `model.LLM`), `ContextAwareModelAdapter` (injects knowledge/memory/RAG into system prompt), `SessionServiceAdapter` (bridges internal session store to ADK session interface), and `AdaptTool()` (converts `agent.Tool` to ADK `tool.Tool`) |
| `agent/` | Core agent types: `Tool` struct (name, description, parameters, handler), `ParameterDef`, `PII Redactor` (regex + optional Presidio integration), `SecretScanner` (prevents credential leakage in model output) |
| `app/` | Application bootstrap and wiring. `app.go` defines `New()` (component initialization), `Start()`, and `Stop()`. `wiring.go` contains all `init*` functions that create individual subsystems. `types.go` defines the `App` struct with all component fields. `tools.go` builds tool collections. `sender.go` provides `channelSender` adapter for delivery |
| `bootstrap/` | Pre-application startup: opens database, initializes crypto provider, loads config profile. Returns `bootstrap.Result` with shared `DBClient` and `Crypto` provider for reuse |

### Presentation

| Package | Description |
|---------|-------------|
| `cli/` | Root Cobra command and subcommand packages |
| `cli/agent/` | `lango agent status`, `lango agent list` -- agent runtime inspection |
| `cli/common/` | Shared CLI helpers (output formatting, error display) |
| `cli/doctor/` | `lango doctor` -- system diagnostics and health checks |
| `cli/graph/` | `lango graph status`, `query`, `stats`, `clear` -- graph store management |
| `cli/memory/` | `lango memory list`, `status`, `clear` -- observational memory management |
| `cli/onboard/` | `lango onboard` -- 5-step guided setup wizard |
| `cli/settings/` | `lango settings` -- full configuration editor |
| `cli/payment/` | `lango payment balance`, `history`, `limits`, `info`, `send` -- payment operations |
| `cli/cron/` | `lango cron add`, `list`, `delete`, `pause`, `resume`, `history` -- cron job management |
| `cli/bg/` | `lango bg list`, `status`, `cancel`, `result` -- background task management |
| `cli/workflow/` | `lango workflow run`, `list`, `status`, `cancel`, `history` -- workflow management |
| `cli/prompt/` | Interactive prompt utilities for CLI input |
| `cli/security/` | `lango security status`, `secrets`, `migrate-passphrase`, `keyring store/clear/status`, `db-migrate`, `db-decrypt`, `kms status/test/keys` -- security operations |
| `cli/p2p/` | `lango p2p status`, `peers`, `connect`, `disconnect`, `firewall list/add/remove`, `discover`, `identity`, `reputation`, `pricing`, `session list/revoke/revoke-all`, `sandbox status/test/cleanup` -- P2P network management |
| `cli/tui/` | TUI components and views for interactive terminal sessions |
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
| `skill/` | File-based skill system. `FileSkillStore` manages skill files on disk. `Registry` loads skills, deploys embedded defaults from `skills/` via `go:embed`, and converts active skills to `agent.Tool` instances |

### Infrastructure

| Package | Description |
|---------|-------------|
| `config/` | YAML configuration loading with environment variable substitution (`${ENV_VAR}` syntax), validation, and defaults. Defines all config structs (`Config`, `AgentConfig`, `SecurityConfig`, etc.) |
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
| `passphrase/` | Passphrase prompt and validation helpers for terminal input |
| `orchestration/` | Multi-agent orchestration. `BuildAgentTree()` creates an ADK agent hierarchy with sub-agents: Operator (tool execution), Navigator (research), Vault (security), Librarian (knowledge), Automator (cron/bg/workflow), Planner (task planning), Chronicler (memory) |
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

Skill system scaffold. Previously included ~30 built-in skills as SKILL.md files deployed via go:embed, but these were removed because Lango's passphrase-protected security model makes it impractical for the agent to invoke lango CLI commands as skills. The skill infrastructure (FileSkillStore, Registry, GitHub importer) remains fully functional for user-defined skills.

## `openspec/`

Specification documents following the OpenSpec workflow. Used for tracking feature specifications, changes, and architectural decisions.
