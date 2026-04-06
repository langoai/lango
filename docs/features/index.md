---
title: Features
---

# Features

Lango provides a comprehensive set of features for building intelligent AI agents. This section covers each subsystem in detail.

<div class="grid cards" markdown>

-   :robot: **[AI Providers](ai-providers.md)**

    ---

    Multi-provider support for OpenAI, Anthropic, Gemini, and Ollama with a unified interface and automatic fallback.

    [:octicons-arrow-right-24: Learn more](ai-providers.md)

-   :speech_balloon: **[Channels](channels.md)**

    ---

    Connect your agent to Telegram, Discord, and Slack. Manage conversations across channels from a single instance.

    [:octicons-arrow-right-24: Learn more](channels.md)

-   :brain: **[Knowledge System](knowledge.md)**

    ---

    Self-learning knowledge store with 8-layer context retrieval, pattern recognition, and agent learning tools.

    [:octicons-arrow-right-24: Learn more](knowledge.md)

-   :eyes: **[Observational Memory](observational-memory.md)**

    ---

    Automatic conversation compression through observations and reflections for long-running sessions.

    [:octicons-arrow-right-24: Learn more](observational-memory.md)

-   :mag: **[Embedding & RAG](embedding-rag.md)**

    ---

    Vector embeddings with OpenAI, Google, or local providers. Retrieval-augmented generation for semantic context injection.

    [:octicons-arrow-right-24: Learn more](embedding-rag.md)

-   :globe_with_meridians: **[Knowledge Graph](knowledge-graph.md)** :material-flask-outline:{ title="Experimental" }

    ---

    BoltDB-backed triple store with hybrid vector + graph retrieval for deep contextual understanding.

    [:octicons-arrow-right-24: Learn more](knowledge-graph.md)

-   :dna: **[Knowledge Ontology](ontology.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Typed knowledge ontology with schema lifecycle governance, temporal truth maintenance, entity resolution, ACL, and P2P exchange.

    [:octicons-arrow-right-24: Learn more](ontology.md)

-   :busts_in_silhouette: **[Multi-Agent Orchestration](multi-agent.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Hierarchical sub-agents (Operator, Navigator, Vault, Librarian, Automator, Planner, Chronicler) working together on complex tasks.

    [:octicons-arrow-right-24: Learn more](multi-agent.md)

-   :satellite: **[A2A Protocol](a2a-protocol.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Agent-to-Agent protocol for remote agent discovery and inter-agent communication.

    [:octicons-arrow-right-24: Learn more](a2a-protocol.md)

-   :globe_with_meridians: **[P2P Network](p2p-network.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Decentralized agent-to-agent connectivity via libp2p with DID identity, knowledge firewall, and ZK-enhanced handshake.

    [:octicons-arrow-right-24: Learn more](p2p-network.md)

-   :moneybag: **[P2P Economy](economy.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Budget management, risk assessment, dynamic pricing, P2P negotiation, and milestone-based escrow for agent commerce.

    [:octicons-arrow-right-24: Learn more](economy.md)

-   :page_facing_up: **[Smart Contracts](contracts.md)** :material-flask-outline:{ title="Experimental" }

    ---

    EVM smart contract interaction with ABI caching, view/pure reads, and state-changing calls.

    [:octicons-arrow-right-24: Learn more](contracts.md)

-   :bank: **[Smart Accounts](smart-accounts.md)** :material-flask-outline:{ title="Experimental" }

    ---

    ERC-7579 modular smart accounts with session keys, ERC-4337 paymaster support, and on-chain policy enforcement.

    [:octicons-arrow-right-24: Learn more](smart-accounts.md)

-   :bar_chart: **[Observability](observability.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Token usage tracking, health monitoring, audit logging, and metrics endpoints for operational visibility.

    [:octicons-arrow-right-24: Learn more](observability.md)

-   :brain: **[Agent Memory](../cli/agent-memory.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Per-agent persistent memory for cross-session context retention and experience accumulation.

    [:octicons-arrow-right-24: Learn more](../cli/agent-memory.md)

-   :toolbox: **[Skill System](skills.md)**

    ---

    File-based skills with import from URLs and GitHub repositories. Extend agent capabilities without code changes.

    [:octicons-arrow-right-24: Learn more](skills.md)

-   :books: **[Proactive Librarian](librarian.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Autonomous knowledge agent that observes conversations and proactively curates the knowledge base.

    [:octicons-arrow-right-24: Learn more](librarian.md)

-   :scroll: **[System Prompts](system-prompts.md)**

    ---

    Customizable prompt sections for agent personality, safety rules, and behavior tuning.

    [:octicons-arrow-right-24: Learn more](system-prompts.md)

-   :office: **[P2P Workspaces](p2p-network.md#collaborative-workspaces)** :material-flask-outline:{ title="Experimental" }

    ---

    Collaborative environments where multiple agents share code, messages, and context with git bundle exchange and contribution tracking.

    [:octicons-arrow-right-24: Learn more](p2p-network.md#collaborative-workspaces)

-   :handshake: **[P2P Teams](p2p-network.md#p2p-team-coordination)** :material-flask-outline:{ title="Experimental" }

    ---

    Task-scoped multi-agent collaboration with role assignment, conflict resolution, budget tracking, and payment coordination.

    [:octicons-arrow-right-24: Learn more](p2p-network.md#p2p-team-coordination)

-   :bell: **[Operational Alerting](alerting.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Threshold-based operational alerting with policy block rate, recovery retry, and circuit breaker monitors.

    [:octicons-arrow-right-24: Learn more](alerting.md)

-   :lock: **[Exec Safety](exec-safety.md)**

    ---

    Policy-based command safety evaluation with shell wrapper unwrapping, opaque pattern detection, and catastrophic pattern blocking.

    [:octicons-arrow-right-24: Learn more](exec-safety.md)

-   :desktop_computer: **[Cockpit TUI](cockpit.md)**

    ---

    Multi-panel terminal dashboard with sidebar navigation, live metrics, and page-based UI for chat, settings, tools, and status.

    [:octicons-arrow-right-24: Learn more](cockpit.md)

-   :shield: **[OS-level Sandbox](../configuration.md#sandbox)** :material-flask-outline:{ title="Experimental" }

    ---

    Process isolation via macOS Seatbelt (Linux: planned, not yet enforced) for tool execution with network deny and workspace-scoped access.

    [:octicons-arrow-right-24: Learn more](../configuration.md#sandbox)

-   :construction: **[Response Gatekeeper](../configuration.md#gatekeeper)**

    ---

    Output sanitization that strips thought tags, internal markers, raw JSON, and custom patterns from agent responses.

    [:octicons-arrow-right-24: Learn more](../configuration.md#gatekeeper)

-   :dart: **[Context Engineering](../configuration.md#context-profile)**

    ---

    Token-budget-aware context allocation with retrieval coordinator, config profiles (off/lite/balanced/full), and relevance auto-adjustment.

    [:octicons-arrow-right-24: Learn more](../configuration.md#context-profile)

-   :package: **[Config Presets](config-presets.md)**

    ---

    Pre-built configuration templates for common deployment scenarios. Quick-start your agent with sensible defaults.

    [:octicons-arrow-right-24: Learn more](config-presets.md)

-   :electric_plug: **[MCP Integration](mcp-integration.md)**

    ---

    Connect to external MCP servers for stdio, HTTP, and SSE transports. Extend agent tooling with the Model Context Protocol.

    [:octicons-arrow-right-24: Learn more](mcp-integration.md)

-   :ledger: **[RunLedger (Task OS)](run-ledger.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Durable execution engine for multi-agent task orchestration with append-only journal, PEV verification, and typed validators.

    [:octicons-arrow-right-24: Learn more](run-ledger.md)

-   :bookmark_tabs: **[Session Provenance](provenance.md)** :material-flask-outline:{ title="Experimental" }

    ---

    Persistent checkpoints, session lineage, git-aware attribution, and signed provenance bundle export/import for auditable multi-agent work.

    [:octicons-arrow-right-24: Learn more](provenance.md)

</div>

## Feature Status

| Feature | Status | Config Key |
|---------|--------|------------|
| AI Providers | Stable | `agent.provider` |
| Channels | Stable | `channels.*` |
| Knowledge System | Stable | `knowledge.enabled` |
| Observational Memory | Stable | `observationalMemory.enabled` |
| Embedding & RAG | Stable | `embedding.*` |
| Knowledge Graph | Experimental | `graph.enabled` |
| Knowledge Ontology | Experimental | `ontology.enabled` |
| Multi-Agent Orchestration | Experimental | `agent.multiAgent` |
| A2A Protocol | Experimental | `a2a.enabled` |
| P2P Network | Experimental | `p2p.enabled` |
| P2P Economy | Experimental | `economy.enabled` |
| Smart Contracts | Experimental | `payment.enabled` |
| Smart Accounts | Experimental | `smartAccount.enabled` |
| Observability | Experimental | `observability.enabled` |
| Skill System | Stable | `skill.enabled` |
| Proactive Librarian | Experimental | `librarian.enabled` |
| System Prompts | Stable | `agent.promptsDir` |
| Agent Memory | Experimental | `agentMemory.enabled` |
| P2P Workspaces | Experimental | `p2p.workspace.enabled` |
| P2P Teams | Experimental | `p2p.enabled` + team coordination |
| Config Presets | Stable | `lango onboard --preset` |
| MCP Integration | Stable | `mcp.enabled` |
| RunLedger (Task OS) | Experimental | `runLedger.enabled` |
| Session Provenance | Experimental | `provenance.enabled` |
| Operational Alerting | Experimental | `alerting.enabled` |
| Exec Safety | Stable | `hooks.blockedCommands` |
| Cockpit TUI | Stable | — |
| OS-level Sandbox | Experimental | `sandbox.enabled` |
| Response Gatekeeper | Stable | `gatekeeper.enabled` |
| Context Engineering | Stable | `context.*`, `retrieval.*`, `contextProfile` |
| Tool Hooks | Experimental | `hooks.enabled` |
| Tool Catalog | Internal | — |
| Event Bus | Internal | — |

!!! note "Experimental Features"

    Features marked as **Experimental** are under active development. Their APIs, configuration keys, and behavior may change between releases. Enable them explicitly via their config flags.
