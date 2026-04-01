---
title: Development
---

# Development

This section covers the development workflow for contributing to Lango.

<div class="grid cards" markdown>

-   :hammer_and_wrench: **[Build & Test](build-test.md)**

    ---

    Build requirements, Makefile targets, test commands, and the local development workflow.

    [:octicons-arrow-right-24: Learn more](build-test.md)

-   :building_construction: **[Architecture](../architecture/index.md)**

    ---

    System overview, project structure, data flow, and layer boundaries.

    [:octicons-arrow-right-24: Learn more](../architecture/index.md)

</div>

## Module Overview

Lango is organized into layered internal packages under `internal/`. Key modules:

| Layer | Packages | Description |
|-------|----------|-------------|
| **Agent Core** | `agent`, `agentregistry`, `agentrt`, `orchestration` | Agent lifecycle, registry, runtime control plane, multi-agent delegation |
| **Knowledge** | `knowledge`, `learning`, `ontology`, `embedding`, `search`, `retrieval` | Knowledge store, learning entries, typed ontology, vector embeddings, FTS5 search, retrieval coordinator |
| **Memory** | `memory`, `agentmemory` | Observational memory (observations/reflections), per-agent persistent memory |
| **Tools** | `tools/`, `toolchain`, `toolcatalog` | Tool implementations (exec, fs, browser, crypto, secrets, payment), tool chaining, catalog registry |
| **Execution Safety** | `tools/exec`, `gatekeeper`, `sandbox` | Policy-based command safety, response sanitization, OS-level process isolation |
| **Automation** | `cron`, `background`, `workflow`, `automation` | Cron scheduling, async tasks, DAG workflow engine |
| **Networking** | `p2p/`, `gateway`, `a2a`, `mcp/` | P2P libp2p networking, HTTP/WS gateway, A2A protocol, MCP integration |
| **Economy** | `economy/`, `payment`, `wallet`, `contract`, `smartaccount/` | Budget, escrow, pricing, USDC payments, ERC-7579 smart accounts |
| **Security** | `security/`, `keyring`, `approval` | Encryption, PII redaction, hardware keyring, tool approval system |
| **Durability** | `runledger`, `provenance`, `session` | Append-only run journal, session provenance/checkpoints, session management |
| **Infrastructure** | `config`, `app`, `bootstrap`, `ent/`, `eventbus`, `logging` | Configuration, app wiring, bootstrapping, Ent ORM, event bus, structured logging |
| **CLI/TUI** | `cli/`, `cli/cockpit/` | CLI command groups, multi-panel cockpit TUI |

## Code Conventions

- **Go version**: 1.25+ with CGO enabled (required for SQLCipher)
- **Build tags**: `fts5,vec` for full-text search and vector embeddings
- **ORM**: [Ent](https://entgo.io/) for database schema and queries
- **CLI framework**: [Cobra](https://github.com/spf13/cobra) for command structure
- **TUI framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) for terminal UIs
- **Style guide**: See `.claude/rules/go-*.md` for detailed Go conventions

## Related

- [Architecture](../architecture/index.md) -- System design and project structure
- [Installation](../getting-started/installation.md) -- Build from source
