---
title: Configuration Presets
---

# Configuration Presets

Presets are purpose-built configuration templates that enable a curated set of features with a single flag. Instead of manually toggling individual settings, choose a preset that matches your use case.

## Usage

Apply a preset during onboarding:

```bash
lango onboard --preset <name>
```

Or when creating a new profile:

```bash
lango config create my-profile --preset researcher
```

## Available Presets

### `minimal`

**Basic AI agent (quick start)**

The default configuration. Enables only the core agent with no optional features. Ideal for getting started quickly or running a simple chatbot.

### `researcher`

**Knowledge, RAG, Graph (research/analysis)**

Designed for knowledge-intensive workflows -- document ingestion, semantic search, and graph-based reasoning.

Enables:

- **Knowledge** -- Document ingestion and retrieval
- **Observational Memory** -- Automatic context tracking across conversations
- **Graph** -- Knowledge graph for entity and relationship queries
- **Embedding & RAG** -- Vector embeddings (OpenAI `text-embedding-3-small`) for semantic search
- **Librarian** -- Knowledge inquiry management

### `collaborator`

**P2P team, payment, workspace (collaboration)**

Designed for multi-agent collaboration over a peer-to-peer network with economic coordination.

Enables:

- **P2P** -- Peer-to-peer networking and discovery
- **Payment** -- USDC wallet and transactions
- **Economy** -- Budget, pricing, negotiation, and escrow

### `full`

**All features enabled (power user)**

Enables everything in `researcher` plus automation, multi-agent orchestration, and MCP integration.

Enables:

- **Knowledge** -- Document ingestion and retrieval
- **Observational Memory** -- Automatic context tracking
- **Graph** -- Knowledge graph
- **Embedding & RAG** -- Vector embeddings (OpenAI `text-embedding-3-small`)
- **Librarian** -- Knowledge inquiry management
- **Cron** -- Scheduled job execution
- **Background** -- Async background task runner
- **Workflow** -- DAG-based YAML workflow engine
- **MCP** -- Model Context Protocol server integration
- **Agent Memory** -- Persistent per-agent memory
- **Multi-Agent** -- Sub-agent orchestration (executor, researcher, planner)

!!! note
    The `full` preset does not enable P2P, Payment, or Economy. Add those manually via `lango settings` if needed.

## Feature Matrix

| Feature | `minimal` | `researcher` | `collaborator` | `full` |
|---------|:---------:|:------------:|:--------------:|:------:|
| Knowledge | | x | | x |
| Observational Memory | | x | | x |
| Graph | | x | | x |
| Embedding & RAG | | x | | x |
| Librarian | | x | | x |
| P2P | | | x | |
| Payment | | | x | |
| Economy | | | x | |
| Cron | | | | x |
| Background | | | | x |
| Workflow | | | | x |
| MCP | | | | x |
| Agent Memory | | | | x |
| Multi-Agent | | | | x |

## Customizing After Preset

Presets set initial values only. After onboarding with a preset, you can enable or disable any feature individually:

```bash
lango settings
```

This opens the interactive TUI editor where you can toggle features, change models, and adjust all configuration options.

## Checking Current Status

See which features are currently enabled:

```bash
lango status
```

See [Status Command](../cli/status.md) for details on the unified status dashboard.
