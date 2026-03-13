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

-   :brain: **Agent Memory** :material-flask-outline:{ title="Experimental" }

    ---

    Per-agent persistent memory for cross-session context retention and experience accumulation.

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

-   :package: **[Config Presets](config-presets.md)**

    ---

    Pre-built configuration templates for common deployment scenarios. Quick-start your agent with sensible defaults.

    [:octicons-arrow-right-24: Learn more](config-presets.md)

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
| Tool Hooks | Experimental | `hooks.enabled` |
| Tool Catalog | Internal | — |
| Event Bus | Internal | — |

!!! note "Experimental Features"

    Features marked as **Experimental** are under active development. Their APIs, configuration keys, and behavior may change between releases. Enable them explicitly via their config flags.
