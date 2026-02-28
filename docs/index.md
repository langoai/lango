---
title: Home
---

# Lango

**A high-performance AI agent framework built with Go.**

Single binary. Multi-provider AI. Self-learning knowledge system.

!!! warning "Experimental"

    Lango is under active development. APIs and configuration formats may change between releases. Use in production at your own risk.

## Quick Install

```bash
git clone https://github.com/langoai/lango.git
cd lango
make build
./bin/lango onboard
```

See the [Installation Guide](getting-started/installation.md) for detailed instructions.

## Features

<div class="grid cards" markdown>

-   :zap: **Fast**

    ---

    Single binary, <100ms startup, <250MB memory footprint. Built with Go for maximum performance.

-   :robot: **Multi-Provider AI**

    ---

    OpenAI, Anthropic, Gemini, and Ollama with a unified interface. Switch providers without changing code.

-   :speech_balloon: **Multi-Channel**

    ---

    Connect to Telegram, Discord, and Slack. Manage conversations across channels from a single agent.

-   :wrench: **Rich Tools**

    ---

    Shell execution, file system operations, browser automation, crypto and secrets management tools.

-   :brain: **Self-Learning**

    ---

    Knowledge store, learning engine, skill system, observational memory, and a proactive librarian that grows smarter over time.

-   :globe_with_meridians: **Knowledge Graph & Graph RAG**

    ---

    BoltDB-backed triple store with hybrid vector + graph retrieval for deep contextual understanding.

-   :busts_in_silhouette: **Multi-Agent Orchestration**

    ---

    Hierarchical sub-agents (Executor, Researcher, Planner, Memory Manager) working together on complex tasks.

-   :satellite: **A2A Protocol**

    ---

    Agent-to-Agent protocol for remote agent discovery and inter-agent communication.

-   :globe_with_meridians: **P2P Network**

    ---

    Decentralized agent connectivity via libp2p with DID identity, knowledge firewall, mDNS discovery, and ZK-enhanced handshake.

-   :coin: **Blockchain Payments**

    ---

    USDC payments on Base L2 with X402 V2 auto-pay protocol support.

-   :alarm_clock: **Cron Scheduling**

    ---

    Persistent cron jobs with Ent ORM storage and multi-channel delivery.

-   :gear: **Background Execution**

    ---

    Async task manager with concurrency control for long-running operations.

-   :arrows_counterclockwise: **Workflow Engine**

    ---

    DAG-based YAML workflows with parallel step execution and dependency management.

-   :lock: **Secure**

    ---

    AES-256-GCM encryption, key registry, secret management, PII redaction, hardware keyring (Touch ID / TPM), SQLCipher database encryption, and Cloud KMS integration.

-   :floppy_disk: **Persistent**

    ---

    Ent ORM with encrypted SQLite storage for sessions, configuration, and knowledge.

-   :electric_plug: **Gateway**

    ---

    WebSocket and HTTP server with real-time streaming support.

-   :key: **Auth**

    ---

    OIDC authentication and OAuth login flow for secure access control.

</div>

## Next Steps

- [Getting Started](getting-started/index.md) -- Install, configure, and run your first agent
- [Architecture](architecture/index.md) -- Understand the system design
- [CLI Reference](cli/index.md) -- Complete command documentation
