<div align="center">
  <img src="./banner.png" alt="Lango Logo">
</div>
<br>


# Lango 🐿️
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/langoai/lango)
[![CI](https://github.com/langoai/lango/actions/workflows/ci.yml/badge.svg)](https://github.com/langoai/lango/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/langoai/lango)](https://github.com/langoai/lango)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/langoai/lango)](https://goreportcard.com/report/github.com/langoai/lango)

> **Early-stage project.** Some features are experimental and may change between releases. See the [feature status table](docs/features/index.md) for stability details.

**A trustworthy multi-agent runtime in Go.** Lango is a high-performance agent runtime that lets AI agents collaborate, learn, and operate autonomously — with built-in observability, security hardening, and an optional peer-to-peer economy layer.

## Why Lango?

Most agent frameworks stop at tool-calling. Lango builds a **trustworthy operational foundation** — then optionally extends into a peer-to-peer economy:

- **Multi-Agent Orchestration** — Hierarchical sub-agent teams with role-based delegation, P2P team coordination with conflict resolution strategies, and DAG-based workflow pipelines.
- **Production Observability** — Token usage tracking, Prometheus metrics, OpenTelemetry tracing, health monitoring, alerting with webhook delivery, and audit logging.
- **Zero-Knowledge Security** — ZK proofs (Plonk/Groth16) for handshake authentication and response attestation. Agents prove identity and output integrity without revealing internals. Hardware keyring and Cloud KMS support.
- **Knowledge as Currency** — Self-learning knowledge graph, observational memory, and hybrid vector + graph RAG retrieval — agents that get smarter with every interaction can charge for their expertise.
- **Open Interoperability** — A2A protocol for remote agent discovery, MCP integration for external tool servers, and multi-provider AI support (OpenAI, Anthropic, Gemini, Ollama).
- **Peer-to-Peer Agent Economy** — Agents discover, authenticate, negotiate prices, and trade capabilities over libp2p with budget management, trust-based risk assessment, and dynamic pricing. No central hub. No vendor lock-in.
- **On-Chain Settlement** — USDC payments on Base Sepolia testnet (chainId 84532) with EIP-3009 authorization, milestone-based escrow (Hub/Vault dual-mode), Foundry smart contracts, and a Security Sentinel that detects anomalies in real time.
- **Escrow Recommendation Execution** — For knowledge-exchange transactions approved with `escrow`, Lango ships a receipt-backed execution slice that binds escrow input during approval and executes `create + fund` through `execute_escrow_recommendation`.
- **Smart Accounts** — ERC-7579 modular smart accounts (Safe-based) with ERC-4337 account abstraction, hierarchical session keys, gasless USDC transactions via paymaster, and on-chain spending limits.
- **Trust & Reputation** — Every interaction builds a verifiable reputation score. Trusted peers get post-pay terms and price discounts; new peers prepay or use escrow.

Single binary. <100ms startup. <250MB memory. Just Go.

## Features

**🤖 Agent Runtime**
- Multi-provider AI (OpenAI, Anthropic, Gemini, Ollama) with a unified interface
- Multi-channel delivery (Telegram, Discord, Slack)
- Rich built-in tools: shell, filesystem, browser automation, crypto, secrets
- Agent registry with `AGENT.md` definitions and dynamic keyword + capability routing

→ [Features overview](docs/features/index.md) · [AI providers](docs/features/ai-providers.md) · [Channels](docs/features/channels.md)

**🧠 Knowledge & Learning**
- Self-learning knowledge store with observational memory
- BoltDB triple store with hybrid vector + graph RAG retrieval
- File-based skill system with GitHub import (git clone + HTTP fallback)
- Proactive Librarian that curates the knowledge base autonomously

→ [Knowledge](docs/features/knowledge.md) · [Knowledge graph](docs/features/knowledge-graph.md) · [Embeddings & RAG](docs/features/embedding-rag.md) · [Skills](docs/features/skills.md) · [Librarian](docs/features/librarian.md)

**🔀 Multi-Agent Orchestration**
- Hierarchical sub-agents: operator, navigator, vault, librarian, automator, planner, chronicler, ontologist
- DAG-based YAML workflows with parallel step execution and state persistence
- Task-scoped P2P teams with conflict resolution (trust-weighted, majority vote, leader decides)

→ [Multi-agent](docs/features/multi-agent.md) · [Workflows & automation](docs/automation/index.md)

**🌐 P2P & Interoperability**
- libp2p network with DHT discovery, ZK-enhanced handshake, and knowledge firewall
- A2A protocol for remote agent discovery and integration
- MCP integration (stdio/HTTP/SSE) with auto-discovery, health checks, multi-scope config

→ [P2P network](docs/features/p2p-network.md) · [A2A protocol](docs/features/a2a-protocol.md) · [MCP integration](docs/features/mcp-integration.md)

**💸 On-Chain Economy**
- USDC payments on Base L2 with X402 V2 auto-pay (Coinbase SDK)
- ERC-7579 modular smart accounts with hierarchical session keys, ERC-4337 paymaster
- Milestone-based escrow with Hub/Vault dual-mode settlement (Foundry contracts)
- Security Sentinel: real-time anomaly detection on escrow flows

→ [Payments](docs/payments/index.md) · [USDC](docs/payments/usdc.md) · [X402](docs/payments/x402.md) · [Economy](docs/features/economy.md) · [Contracts](docs/features/contracts.md)

**🔒 Security & Privacy**
- Master Key envelope (MK/KEK hierarchy) with AES-256-GCM
- Brokered payload protection for sessions, learnings, inquiries, agent memory
- Hardware keyring (Touch ID / TPM) and Cloud KMS (AWS / GCP / Azure / PKCS#11)
- OS-level sandbox (macOS Seatbelt, Linux bubblewrap) with network deny and workspace-scoped writes
- Response Gatekeeper output sanitization; PII redaction; recovery mnemonic

→ [Security](docs/security/index.md) · [Encryption](docs/security/encryption.md) · [Approval flow](docs/security/approval-flow.md) · [Authentication](docs/security/authentication.md)

**🖥️ Operations & UX**
- Cockpit TUI with Mission Control, Chat, Settings, Tools, Status, Sessions, Tasks, Dead Letters, Approvals
- Persistent cron jobs and async background task manager with completion notifications
- RunLedger (Task OS) — durable execution engine with append-only journal and PEV verification
- Session Provenance — checkpoints, lineage tree, git attribution, signed bundle export/import
- Prometheus metrics, OpenTelemetry tracing, OIDC authentication, gateway WebSocket/HTTP

→ [Cockpit](docs/features/cockpit.md) · [Observability](docs/features/observability.md) · [RunLedger](docs/features/run-ledger.md) · [Provenance](docs/features/provenance.md)

**📦 Extensibility**
- Extension packs: skills, modes, prompts in inspectable bundles (`lango extension install`)
- Config presets via `lango onboard --preset`
- Per-agent persistent memory for cross-session context
- Typed event bus and tool-hook middleware (security filter, access control, knowledge save)

→ [Skills](docs/features/skills.md) · [Config presets](docs/features/config-presets.md) · [Agent memory](docs/cli/agent-memory.md)

## Quick Start

### Install

```bash
git clone https://github.com/langoai/lango.git
cd lango
make build
./bin/lango onboard
```

`onboard` is interactive — it sets up your master key passphrase, picks a provider, and writes `~/.lango/config.yaml`. For headless setups or detailed flags, see [Installation](docs/getting-started/installation.md).

### Configure (minimal)

```bash
# Set provider + API key (the --from-env form avoids saving secrets in shell history)
./bin/lango config set agent.provider openai
OPENAI_API_KEY=sk-... ./bin/lango config set providers.openai.apiKey --from-env OPENAI_API_KEY

# Or use a preset
./bin/lango onboard --preset minimal
```

Full configuration reference: [getting started](docs/getting-started/configuration.md) · [complete schema](docs/configuration.md).

### Run

```bash
# Mission Workbench TUI (default — bare invocation)
./bin/lango

# Multi-panel Cockpit dashboard (Mission Control, Sessions, Tasks, Approvals, ...)
./bin/lango cockpit

# Interactive chat session
./bin/lango chat

# Headless gateway server (port is read from `server.port` in config — see docs/configuration.md)
./bin/lango serve
```

See the [Quick start guide](docs/getting-started/quickstart.md) for first-task walkthroughs.

## Architecture

Lango is a layered Go runtime. Each layer depends only on layers below it.

```
┌─────────────────────────────────────────────────────────────┐
│  Surfaces      TUI (Workbench/Cockpit/Chat) · Gateway · CLI  │
├─────────────────────────────────────────────────────────────┤
│  Orchestration agentrt · TurnRunner · multi-agent · workflow │
├─────────────────────────────────────────────────────────────┤
│  Capabilities  tools · skills · MCP · A2A · learning · RAG   │
├─────────────────────────────────────────────────────────────┤
│  Economy       payments · contracts · escrow · smart account │
├─────────────────────────────────────────────────────────────┤
│  Knowledge     KG triple store · vector index · memory       │
├─────────────────────────────────────────────────────────────┤
│  Platform      session · storage (Ent/SQLite) · event bus    │
├─────────────────────────────────────────────────────────────┤
│  Security      MK/KEK envelope · sandbox · approval · ZK     │
└─────────────────────────────────────────────────────────────┘
                     P2P transport (libp2p) ⇄ peers
```

**Surfaces** never contain business logic — they parse input, call the orchestration layer, and format output. `agentrt` is the control plane on top of ADK that wires `TurnRunner`, multi-agent delegation, and the event bus.

**Capabilities** are pluggable: tools register via a middleware chain (security filter, access control, event publish, knowledge save); skills load from files or GitHub; MCP and A2A connect to external runtimes.

**Economy** is optional but first-class: USDC settlement on Base L2, ERC-7579 smart accounts with session keys, and Hub/Vault escrow contracts run alongside the Security Sentinel for anomaly detection.

For deeper structure, read:

- [System overview](docs/architecture/overview.md)
- [Data flow](docs/architecture/data-flow.md)
- [Project structure](docs/architecture/project-structure.md)
- [Master architecture document](docs/architecture/master-document.md)

## Documentation

Full documentation lives under `docs/`. Below is a navigation map:

### Getting Started

| Topic | Doc |
|-------|-----|
| Installation | [docs/getting-started/installation.md](docs/getting-started/installation.md) |
| Quick start | [docs/getting-started/quickstart.md](docs/getting-started/quickstart.md) |
| Configuration | [docs/getting-started/configuration.md](docs/getting-started/configuration.md) |

### CLI Reference

| Command family | Doc |
|----------------|-----|
| Overview | [docs/cli/index.md](docs/cli/index.md) |
| `agent` · `run` · `config` | [agent](docs/cli/agent.md) · [run](docs/cli/run.md) · [config](docs/cli/config.md) |
| `payment` · `contract` · `smartaccount` · `economy` | [payment](docs/cli/payment.md) · [contract](docs/cli/contract.md) · [smartaccount](docs/cli/smartaccount.md) · [economy](docs/cli/economy.md) |
| `p2p` · `a2a` · `mcp` | [p2p](docs/cli/p2p.md) · [a2a](docs/cli/a2a.md) · [mcp](docs/cli/mcp.md) |
| `learning` · `librarian` · `graph` · `agent-memory` | [learning](docs/cli/learning.md) · [librarian](docs/cli/librarian.md) · [graph](docs/cli/graph.md) · [agent-memory](docs/cli/agent-memory.md) |
| `security` · `approval` · `sandbox` | [security](docs/cli/security.md) · [approval](docs/cli/approval.md) · [sandbox](docs/cli/sandbox.md) |
| `automation` · `provenance` · `extension` · `status` | [automation](docs/cli/automation.md) · [provenance](docs/cli/provenance.md) · [extension](docs/cli/extension.md) · [status](docs/cli/status.md) |
| `alerts` · `metrics` · `core` | [alerts](docs/cli/alerts.md) · [metrics](docs/cli/metrics.md) · [core](docs/cli/core.md) |

### Architecture

| Topic | Doc |
|-------|-----|
| System overview | [docs/architecture/overview.md](docs/architecture/overview.md) |
| Data flow | [docs/architecture/data-flow.md](docs/architecture/data-flow.md) |
| Project structure | [docs/architecture/project-structure.md](docs/architecture/project-structure.md) |
| Master document | [docs/architecture/master-document.md](docs/architecture/master-document.md) |

### Features

See [docs/features/index.md](docs/features/index.md) for the complete feature catalog with stability annotations. Highlights:

- Cockpit TUI · AI providers · Channels · Multi-agent · A2A · P2P · MCP
- Knowledge graph · Embeddings & RAG · Learning · Librarian · Skills · Agent memory
- Economy · Contracts · Observability · Provenance · RunLedger · Exec safety

### Security

| Topic | Doc |
|-------|-----|
| Index | [docs/security/index.md](docs/security/index.md) |
| Encryption | [docs/security/encryption.md](docs/security/encryption.md) |
| Approval flow | [docs/security/approval-flow.md](docs/security/approval-flow.md) |
| Authentication | [docs/security/authentication.md](docs/security/authentication.md) |
| PII redaction | [docs/security/pii-redaction.md](docs/security/pii-redaction.md) |
| Envelope migration | [docs/security/envelope-migration.md](docs/security/envelope-migration.md) |
| Tool approval | [docs/security/tool-approval.md](docs/security/tool-approval.md) |
| Exportability | [docs/security/exportability.md](docs/security/exportability.md) |

### Payments & Economy

| Topic | Doc |
|-------|-----|
| Payments index | [docs/payments/index.md](docs/payments/index.md) |
| USDC | [docs/payments/usdc.md](docs/payments/usdc.md) |
| X402 | [docs/payments/x402.md](docs/payments/x402.md) |

### Automation

- [docs/automation/index.md](docs/automation/index.md) — cron jobs, workflows, background tasks

### Gateway

- [docs/gateway/index.md](docs/gateway/index.md) — HTTP and WebSocket API

### Deployment

- [docs/deployment/docker.md](docs/deployment/docker.md)
- [docs/deployment/production.md](docs/deployment/production.md)

### Development

- [docs/development/build-test.md](docs/development/build-test.md)
- [docs/development/index.md](docs/development/index.md)

## Project Status

Lango is under active development. Feature stability varies — see the [feature status table](docs/features/index.md) for current annotations (stable · beta · experimental · planned).

Active workstreams and design specs live in [`openspec/`](openspec/) (proposed and active changes) and [`internal-docs/`](internal-docs/) (planning artifacts). For Phase progress notes, see the [development index](docs/development/index.md).

## Contributing

Contributions are welcome. Before opening a PR:

1. Read [`docs/development/build-test.md`](docs/development/build-test.md) for the build, test, and lint commands.
2. Match the existing code style — see [`.claude/rules/go-style.md`](.claude/rules/go-style.md), [`go-guidelines.md`](.claude/rules/go-guidelines.md), [`go-errors.md`](.claude/rules/go-errors.md), and [`go-patterns.md`](.claude/rules/go-patterns.md).
3. For substantive changes, open an OpenSpec change first under `openspec/changes/`.

## License

[MIT](LICENSE)
