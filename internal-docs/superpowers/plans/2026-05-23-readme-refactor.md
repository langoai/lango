# README Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite `README.md` from 1887 lines into a ~450-line landing page that hands deep references off to `docs/`, eliminating duplicated content.

**Important worker note — fence handling:** In Tasks 4–8 below, each task contains a long block of new README content quoted **inside** a code fence (so the plan can be read as Markdown). When inserting this content into `README.md`, write only the **inner** content. Do NOT include the outer ```` ```markdown ```` opener or its closing ```` ``` ````. The inner content itself contains nested ```` ```bash ```` and ```` ``` ```` (ASCII diagram) blocks — those ARE part of the README content and must be preserved literally.

**Important worker note — code-change rule:** This refactor touches only `README.md` and `docs/`. Per `.claude/CLAUDE.md`, the "OpenSpec workflow after code changes" rule applies to code changes; docs-only changes do not require an OpenSpec change. Skip the OpenSpec workflow. The `go build ./...` check in Task 10 is a sanity check only, not required by the rule.

**Architecture:** Single-file refactor. `README.md` becomes an 8-section landing page (Hero, Why, Features, Quick Start, Architecture, Documentation Map, Status, Contributing/License). All deep references (CLI listings, configuration tables, payment internals, security details) are removed from README and replaced with links into existing `docs/` pages. Pre-flight spot-checks confirm `docs/` already contains equivalent content before deletion. A post-write link verification script catches any broken `docs/…` reference and auto-stubs missing pages.

**Tech Stack:** Markdown, plain shell (grep/find/awk), git. No build-system changes.

**Reference:** Design spec at `internal-docs/superpowers/specs/2026-05-23-readme-refactor-design.md`.

---

## File Structure

| Path | Action | Responsibility |
|------|--------|----------------|
| `README.md` | Rewrite (1887 → ~450 lines) | Landing page + Documentation Map |
| `docs/configuration.md` | Possibly extend (only if delta found vs README 649–997) | Reference for full configuration schema |
| `docs/features/system-prompts.md` | Create only if missing AND README 1103–1174 contains unique content | System prompt customization reference |
| `docs/**` (any link target) | Auto-stub one-liner if missing | Prevent broken links from new README |

No source code, no tests, no CI changes. Verification commands are inline shell snippets.

---

### Task 1: Pre-flight verification of working tree

**Files:**
- Read-only: `README.md`, `docs/configuration.md`, `Makefile`

- [ ] **Step 1: Confirm clean working tree**

Run:
```bash
cd /Users/juwonkim/GolandProjects/lango
git status --short
```
Expected: empty output (no uncommitted changes). If output is non-empty, STOP and ask the user how to proceed — do not stash or discard their work.

- [ ] **Step 2: Confirm current README size as baseline**

Run:
```bash
wc -l README.md
```
Expected: `1887 README.md` (or close — accept ±5 lines as drift). Note the number; the new README must be ≤500 lines.

- [ ] **Step 3: Confirm docs/ targets exist**

Run:
```bash
for f in \
  docs/index.md \
  docs/configuration.md \
  docs/getting-started/installation.md \
  docs/getting-started/quickstart.md \
  docs/getting-started/configuration.md \
  docs/cli/index.md \
  docs/architecture/overview.md \
  docs/architecture/data-flow.md \
  docs/architecture/project-structure.md \
  docs/architecture/master-document.md \
  docs/features/index.md \
  docs/features/cockpit.md \
  docs/features/multi-agent.md \
  docs/features/p2p-network.md \
  docs/features/knowledge-graph.md \
  docs/features/economy.md \
  docs/features/skills.md \
  docs/features/observability.md \
  docs/features/learning.md \
  docs/features/embedding-rag.md \
  docs/features/a2a-protocol.md \
  docs/features/mcp-integration.md \
  docs/security/index.md \
  docs/security/encryption.md \
  docs/security/approval-flow.md \
  docs/security/authentication.md \
  docs/security/pii-redaction.md \
  docs/security/envelope-migration.md \
  docs/payments/index.md \
  docs/payments/usdc.md \
  docs/payments/x402.md \
  docs/automation/index.md \
  docs/deployment/docker.md \
  docs/deployment/production.md \
  docs/development/build-test.md \
  docs/development/index.md; do
    test -f "$f" && echo "OK  $f" || echo "MISSING  $f"
done
```
Expected: every line starts with `OK`. Any `MISSING` line gets tracked — Task 9 will auto-stub them, but a large gap (>3 missing) means the docs landscape changed since the spec was written; pause and reconsider links before proceeding.

- [ ] **Step 4: Verify Quick Start install commands match reality**

Run:
```bash
grep -nE '^(build|install)[[:space:]]*:' Makefile | head -5
grep -nE 'bin/lango|lango[[:space:]]' docs/getting-started/installation.md | head -10
```
Expected: `make build` produces an executable at `./bin/lango` (or whatever the Makefile defines). If the path differs (e.g. `bin/lango` without leading `./`, or `cmd/lango`), update Task 5's Quick Start snippet to match. Do not invent commands — verify against the Makefile and the existing installation doc first.

- [ ] **Step 5: Verify docs/features/system-prompts.md status**

Run:
```bash
ls -l docs/features/system-prompts.md 2>&1 && wc -l docs/features/system-prompts.md
```
Expected at planning time: file exists (~113 lines). If it exists, Task 3 will be a no-op spot-check. If it does NOT exist (drift since planning), Task 3 Step 3 will create it.

- [ ] **Step 6: Commit baseline check note**

No commit yet — Task 1 is pure verification. Proceed to Task 2.

---

### Task 2: Spot-check Configuration Reference for migration delta

**Files:**
- Read-only: `README.md` lines 649–997, `docs/configuration.md`

- [ ] **Step 1: Extract subsection headings from README config range**

Run:
```bash
awk 'NR>=649 && NR<=997 && /^### / {print NR": "$0}' README.md
```
Save the output mentally (or to scratch). These are the topics that must exist in `docs/configuration.md`.

- [ ] **Step 2: Check headings in docs/configuration.md**

Run:
```bash
grep -n "^## \|^### " docs/configuration.md
```
Compare against Step 1 output. For each README subsection, confirm a corresponding heading or content paragraph exists in `docs/configuration.md`.

- [ ] **Step 3: Decision point**

- If every README subsection is covered in `docs/configuration.md` → no migration needed, proceed to Task 3.
- If a README subsection is missing from docs → copy the relevant lines from README into `docs/configuration.md` at the most appropriate location. Maintain heading hierarchy (README's `###` becomes `##` or `###` in docs depending on existing structure).
- Do **not** modify README in this task — that happens in Tasks 4–9.

- [ ] **Step 4: Verify configuration.md still builds as markdown**

Run:
```bash
wc -l docs/configuration.md
head -20 docs/configuration.md
```
Expected: line count ≥ original (1563); top-of-file frontmatter and `# Configuration Reference` heading intact.

- [ ] **Step 5: Stage if changes were made (do not commit)**

Per the project rule `feedback_user_commits_manually`: the worker stages files but does NOT run `git commit`. Only if Step 3 modified `docs/configuration.md`:
```bash
git add docs/configuration.md
```
Then surface this suggested commit message to the user — let them run the commit themselves:
```
docs(configuration): migrate residual content from README
```
If Step 3 made no changes, skip staging entirely.

---

### Task 3: Spot-check System Prompts for migration delta

**Expected outcome:** No-op spot-check. At planning time `docs/features/system-prompts.md` already exists (~113 lines). Task 1 Step 5 confirms this. Step 3 below should not fire unless the file was deleted between planning and execution.

**Files:**
- Read-only: `README.md` lines 1103–1174
- Possibly create (only if file no longer exists): `docs/features/system-prompts.md`

- [ ] **Step 1: Inspect README System Prompts section**

Run:
```bash
sed -n '1103,1174p' README.md
```
Read the content. Note its subsections (e.g., Prompt Sections, Customizing Prompts, Per-Agent Prompt Customization).

- [ ] **Step 2: Check if a docs target exists**

Run:
```bash
ls docs/features/ | grep -i prompt
grep -rn "system prompt\|prompt section\|customizing prompts" docs/features/ | head -20
```
Possible outcomes:
- A page like `system-prompts.md` exists → confirm content equivalence and skip Step 3.
- Content is partially absorbed into another feature page (e.g., `multi-agent.md`) → confirm coverage.
- No coverage anywhere → proceed to Step 3.

- [ ] **Step 3: Create docs/features/system-prompts.md (only if Step 2 found no coverage)**

Create file with this exact content (replace `<README BODY>` with the actual lines 1103–1174 content, reformatted with `##` headings instead of `###`):

```markdown
---
title: System Prompts
---

# System Prompts

Lango composes a system prompt from named sections. Operators can override or
extend any section without recompiling — both globally and per agent.

<README BODY>
```

- [ ] **Step 4: Verify new file**

Run:
```bash
test -f docs/features/system-prompts.md && wc -l docs/features/system-prompts.md
```
Expected: file exists, ≥30 lines.

- [ ] **Step 5: Stage if file was created (do not commit)**

Per `feedback_user_commits_manually`: the worker stages but does NOT commit. Only if Step 3 created a file:
```bash
git add docs/features/system-prompts.md
```
Surface this suggested commit message to the user for them to run:
```
docs(features): add system-prompts reference page
```

---

### Task 4: Write new README — Hero, Why, Features

**Files:**
- Modify: `README.md` (full rewrite begins; intermediate state is acceptable since later tasks append further sections)

- [ ] **Step 1: Begin atomic rewrite — write Hero + Why + Features**

Overwrite `README.md` with the following exact content (this is sections 1–3 of the new README; Task 5–7 will append sections 4–8 by editing the file rather than overwriting it):

```markdown
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
```

- [ ] **Step 2: Verify line count is on track**

Run:
```bash
wc -l README.md
```
Expected: between 90 and 130 lines after Task 4 (sections 1–3 only).

- [ ] **Step 3: Do not commit yet**

The README is mid-rewrite. Wait until Task 9 completes before committing.

---

### Task 5: Append Quick Start section

**Files:**
- Modify: `README.md` (append section 4 at the bottom)

- [ ] **Step 1: Append Quick Start to README.md**

Open the file and append the following block at the end:

```markdown

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
# Set provider + API key
./bin/lango config set ai.provider openai
./bin/lango config set ai.openai.api_key sk-...

# Or use a preset
./bin/lango onboard --preset minimal
```

Full configuration reference: [getting started](docs/getting-started/configuration.md) · [complete schema](docs/configuration.md).

### Run

```bash
# Interactive Cockpit TUI (default)
./bin/lango

# Single-turn chat
./bin/lango run "What is the capital of France?"

# Headless gateway server
./bin/lango gateway --port 8080
```

See the [Quick start guide](docs/getting-started/quickstart.md) for first-task walkthroughs.
```

- [ ] **Step 2: Verify section count**

Run:
```bash
grep -c "^## " README.md
```
Expected: exactly 3 — `## Why Lango?`, `## Features`, `## Quick Start`.

---

### Task 6: Append Architecture overview section

**Files:**
- Modify: `README.md` (append section 5)

- [ ] **Step 1: Append Architecture section**

```markdown

## Architecture

Lango is a layered Go runtime. Each layer depends only on layers below it.

```
┌─────────────────────────────────────────────────────────────┐
│  Surfaces      Cockpit TUI · Gateway (HTTP/WS) · CLI         │
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
```

- [ ] **Step 2: Spot-check the diagram renders**

Run:
```bash
sed -n '/^## Architecture/,/^## /p' README.md | head -40
```
Expected: the ASCII box renders cleanly in monospace. If the editor or pasting collapsed vertical bars, fix in place before continuing.

---

### Task 7: Append Documentation Map section

**Files:**
- Modify: `README.md` (append section 6)

- [ ] **Step 1: Append Documentation Map**

```markdown

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
```

- [ ] **Step 2: Verify table count**

Run:
```bash
grep -c "^|" README.md
```
Expected: at least 30 table rows (tables across Documentation Map).

---

### Task 8: Append Project Status, Contributing, License

**Files:**
- Modify: `README.md` (append sections 7–8)

- [ ] **Step 1: Append final sections**

```markdown

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
```

- [ ] **Step 2: Final line-count check**

Run:
```bash
wc -l README.md
```
Expected: between 380 and 500 lines.

- [ ] **Step 3: Final section structure check**

Run:
```bash
grep -n "^## \|^### " README.md
```
Expected: a clean top-level outline — exactly these `##` headings in order:
`Why Lango?`, `Features`, `Quick Start`, `Architecture`, `Documentation`, `Project Status`, `Contributing`, `License`. Each may have `###` subheadings under Quick Start and Documentation.

---

### Task 9: Verify all `docs/` links and auto-stub missing pages

**Files:**
- Read-only: `README.md`
- Possibly create: any `docs/**.md` referenced but missing

- [ ] **Step 1: Extract every docs/ link from README**

Run:
```bash
grep -oE '\(docs/[^)]+\)' README.md | tr -d '()' | sort -u
```
This prints the deduplicated list of doc paths the new README references.

- [ ] **Step 2: Check each path exists**

Run:
```bash
MISSING=""
for f in $(grep -oE '\(docs/[^)]+\)' README.md | tr -d '()' | sort -u); do
  test -f "$f" || MISSING="$MISSING $f"
done
echo "MISSING:$MISSING"
```
Expected: `MISSING:` (empty after the colon).

- [ ] **Step 3: Auto-stub any missing pages**

For each path in `MISSING`, create a one-line placeholder. The stub format:

```markdown
---
title: TODO
---

# TODO: <topic name derived from filename>

<!-- TODO(readme-refactor): This page was stubbed during the 2026-05-23 README refactor. Replace with real content. -->

This page is a placeholder. Related references:

- [Features index](../features/index.md)
- [Documentation map (README)](../../README.md)
```

For each missing path, the title and topic should be a Title-Case version of the filename (e.g., `system-prompts.md` → `System Prompts`).

Example shell loop to generate stubs. **Note:** macOS BSD `sed` does not support `\u` (uppercase) or `\b` (word boundary) — use `awk` for Title Case so the loop works on both macOS and Linux:

```bash
for f in $MISSING; do
  mkdir -p "$(dirname "$f")"
  topic=$(basename "$f" .md | tr '-' ' ' | awk '{for(i=1;i<=NF;i++)$i=toupper(substr($i,1,1))substr($i,2)}1')
  cat > "$f" <<EOF
---
title: ${topic} (TODO)
---

# ${topic}

<!-- TODO(readme-refactor): This page was stubbed during the 2026-05-23 README refactor. Replace with real content. -->

This page is a placeholder. Related references:

- [Features index](../features/index.md)
- [Documentation map (README)](../../README.md)
EOF
done
```

Verify the Title Case worked on the first stub before generating the rest:
```bash
# Sanity check the awk pipeline on a sample
echo "system-prompts" | tr '-' ' ' | awk '{for(i=1;i<=NF;i++)$i=toupper(substr($i,1,1))substr($i,2)}1'
# Expected: System Prompts
```

If any generated stub lives at depth other than 2 (e.g. `docs/foo.md` instead of `docs/category/foo.md`), the relative links in the template (`../features/index.md`, `../../README.md`) will be wrong. With the current README content all stubbed targets live at depth ≥2, so this is not expected to trigger — but verify with `find docs -maxdepth 1 -name '*.md' -newer /tmp/readme-refactor-marker 2>/dev/null` after stubs are written (touch the marker before Step 3).

- [ ] **Step 4: Re-run verification to confirm zero missing**

Run:
```bash
MISSING=""
for f in $(grep -oE '\(docs/[^)]+\)' README.md | tr -d '()' | sort -u); do
  test -f "$f" || MISSING="$MISSING $f"
done
echo "MISSING:$MISSING"
```
Expected: `MISSING:` (empty).

---

### Task 10: Optional sanity check — Go build

**Files:**
- Read-only. This refactor touches no Go source. CLAUDE.md's `go build ./...` rule applies to "any code changes" — this is docs-only. Treat this task as an optional sanity check that the working tree is not coincidentally broken from upstream merges; it does NOT validate the refactor.

- [ ] **Step 1: Optional build check**

Run:
```bash
go build ./...
```
Expected: zero output, exit 0. **If this fails**, do NOT assume the docs refactor caused it — verify the same failure exists on `git stash && git checkout -- README.md` if needed. A pre-existing build failure is not in scope for this refactor.

- [ ] **Step 2: Skip tests**

No Go source changed; do not run `go test ./...`. The OpenSpec workflow is also NOT required for this docs-only change per the same scoping.

---

### Task 11: Final review and commit

**Files:**
- Modify: commit `README.md` plus any stubs and migrated docs

- [ ] **Step 1: Review the full diff**

Run:
```bash
git diff --stat
git diff README.md | head -120
```
Expected: `README.md` shows large deletions and the new structure. Any new stub or migrated-content files appear as additions.

- [ ] **Step 2: Confirm line count reduction**

Run:
```bash
wc -l README.md
git show HEAD:README.md | wc -l
```
Expected: new count between 380 and 500; previous count ~1887. Ratio should be ≥3x reduction.

- [ ] **Step 3: Stage all changes**

Run:
```bash
git add README.md
# Add docs files only if they were modified or created in earlier tasks:
git add docs/configuration.md 2>/dev/null
git add docs/features/system-prompts.md 2>/dev/null
# Add any auto-stubs created in Task 9 (untracked files only):
git ls-files --others --exclude-standard docs/ | xargs -r git add
```

- [ ] **Step 4: Commit**

Per `.claude/CLAUDE.md` rule **user-commits-manually**: do **not** run `git commit` yourself. Instead, present the suggested commit message to the user and let them commit:

```
docs(readme): refactor into landing page with docs/ navigation

Reduce README from 1887 to ~450 lines. Move CLI/configuration/payment/
security details into the existing docs/ tree. Add Documentation Map as
the primary navigation surface. Auto-stub any link targets that were not
yet present in docs/.

See internal-docs/superpowers/specs/2026-05-23-readme-refactor-design.md
for the design and internal-docs/superpowers/plans/2026-05-23-readme-refactor.md
for the implementation plan.
```

Stop here. The user runs `git commit -m "..."` themselves.

---

## Self-Review

- **Spec coverage:**
  - Hero/Why → Task 4 ✓
  - Features (hybrid, 8 categories) → Task 4 ✓
  - Quick Start → Task 5 ✓
  - Architecture diagram + 3 paragraphs → Task 6 ✓
  - Documentation Map (9 categories) → Task 7 ✓
  - Project Status, Contributing, License → Task 8 ✓
  - Link verification + auto-stub → Task 9 ✓
  - Configuration spot-check → Task 2 ✓
  - System Prompts spot-check → Task 3 ✓
  - mkdocs.yml verification → resolved during planning (no mkdocs.yml present); explicit task omitted.
- **Placeholder scan:** No `TBD`/`fill in later`/`add appropriate X` strings. The auto-stub template intentionally writes `TODO` markers — that is the contract for missing-page placeholders, not laziness in the plan.
- **Type consistency:** N/A (no type system). Path consistency is enforced by Task 9 link verification.
- **User-commits-manually:** Task 11 honors the project memory rule — Claude does not commit, only suggests the message.
