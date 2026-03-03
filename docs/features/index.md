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
| Skill System | Stable | `skill.enabled` |
| Proactive Librarian | Experimental | `librarian.enabled` |
| System Prompts | Stable | `agent.promptsDir` |
| Agent Memory | Experimental | `agentMemory.enabled` |
| Tool Hooks | Experimental | `hooks.enabled` |
| Tool Catalog | Internal | — |
| Event Bus | Internal | — |

!!! note "Experimental Features"

    Features marked as **Experimental** are under active development. Their APIs, configuration keys, and behavior may change between releases. Enable them explicitly via their config flags.
