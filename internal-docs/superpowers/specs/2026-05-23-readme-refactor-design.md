# README Refactor — Design Spec

- **Date**: 2026-05-23
- **Owner**: Tech Writer (with Architect review for diagram)
- **Status**: Draft (awaiting user review)

## Goal

Shrink `README.md` from ~1887 lines to ~450 lines by treating it as a landing page that
points readers into `docs/`, which already contains comprehensive coverage of every
topic currently inlined in the README. Deep references (CLI command listings,
configuration reference, payment internals, security details) move out of README
entirely; readers navigate to `docs/` from a single Documentation Map.

## Why now

The current README:

- Is 1887 lines and 82 sections — far past the point of useful landing-page scanning.
- Duplicates `docs/configuration.md` (1563 lines), `docs/cli/*` (25 files),
  `docs/features/*` (25 files), `docs/security/*`, `docs/payments/*`,
  `docs/architecture/*`, `docs/deployment/*` — meaning information drifts whenever
  one copy is updated without the other.
- Buries the project's strongest pitch (Why Lango / Multi-Agent / Economy) under
  configuration tables and long CLI listings.

The docs site is already production-grade. The README should hand readers off to it.

## Non-goals

- Reorganizing `docs/` itself. Docs structure is treated as the source of truth and
  left as-is.
- Rewriting feature descriptions. We are mostly cutting and linking, not editing
  prose.
- Touching `mkdocs.yml` navigation. (We will only verify it does not reference
  README content that we are removing.)
- Adding new features, new docs pages, or new CLI behavior. Pure content refactor.

## Target structure

New `README.md` — eight sections, ~450 lines total:

| # | Section | Approx lines | Notes |
|---|---------|--------------|-------|
| 1 | Hero (banner, badges, tagline, experimental warning) | ~15 | Carry over current lines 1–17 with the experimental warning collapsed to a single line. |
| 2 | Why Lango | ~25 | Keep the current 10-bullet differentiator block at lines 19–34 verbatim. |
| 3 | Features (hybrid: 8 category headings + emoji bullets + `→ docs/` link per category) | ~70 | Replaces the current ~40 flat bullets. |
| 4 | Quick Start (install, minimal config, run) | ~80 | Minimal happy-path; details live in `docs/getting-started/`. |
| 5 | Architecture overview (ASCII layer diagram + 3 paragraphs) | ~70 | Links to `docs/architecture/*` for deeper reading. |
| 6 | Documentation Map (categorized link tables) | ~120 | Single navigation surface; readers go here to find anything. |
| 7 | Project Status & Roadmap | ~20 | Link to `docs/features/index.md` feature status table; mention OpenSpec workflow. |
| 8 | Contributing & License | ~20 | Standard footer. |

### Feature categories (Section 3)

Eight categories, each rendered as a bold category title, 2–3 emoji bullets describing
the headline capabilities, and a single `→` link to the relevant `docs/` page:

1. **🤖 Agent Runtime** — multi-provider AI, multi-channel, rich tools, agent registry.
   → `docs/features/index.md`
2. **🧠 Knowledge & Learning** — knowledge store, observational memory, hybrid
   vector + graph RAG, proactive librarian, file-based skills. → `docs/features/knowledge.md`
3. **🔀 Multi-Agent Orchestration** — hierarchical sub-agents, DAG workflows, P2P teams.
   → `docs/features/multi-agent.md`
4. **🌐 P2P & Interoperability** — A2A protocol, libp2p network, MCP integration.
   → `docs/features/p2p-network.md`
5. **💸 On-Chain Economy** — USDC on Base L2, X402 V2, ERC-7579 smart accounts, escrow.
   → `docs/payments/index.md` · `docs/features/economy.md`
6. **🔒 Security & Privacy** — MK/KEK envelope, AES-256-GCM, hardware keyring, Cloud
   KMS, sandbox, Response Gatekeeper, ZK proofs. → `docs/security/index.md`
7. **🖥️ Operations & UX** — Cockpit TUI, cron + background tasks, RunLedger,
   provenance, Prometheus/OTel, OIDC auth. → `docs/features/cockpit.md` ·
   `docs/features/observability.md`
8. **📦 Extensibility** — extension packs, config presets, custom prompts, per-agent
   memory, event bus, tool hooks. → `docs/features/skills.md`

### Documentation Map (Section 6)

Ten sub-sections, each a small Markdown table (or short list) mapping topic → doc path. Categories:

1. **Getting Started** — installation, quickstart, configuration.
2. **CLI Reference** — overview + grouped command families (agent/run/config,
   payment/contract/smartaccount/economy, p2p/a2a/mcp, learning/librarian/graph/agent-memory,
   security/approval/sandbox, automation/provenance/extension/status,
   alerts/metrics/core).
3. **Architecture** — overview, data-flow, project-structure, master-document.
4. **Features** — pointer to `docs/features/index.md` + a one-line list of available
   feature pages (cockpit, AI providers, channels, multi-agent, A2A, P2P, MCP,
   knowledge graph, embeddings/RAG, learning, librarian, skills, agent memory,
   economy, contracts, observability, provenance, RunLedger, exec safety).
5. **Security** — index, encryption, approval-flow, authentication, PII redaction,
   envelope migration.
6. **Payments & Economy** — payments index, USDC, X402.
7. **Automation** — index (cron, workflows, background tasks).
8. **Deployment** — docker, production.
9. **Development** — build-test, contributor index.

### Architecture diagram (Section 5)

A simple ASCII layered diagram showing dependency direction (top → bottom):

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

Followed by 3 short paragraphs covering: layer rule (Surfaces never carry business
logic), capability pluggability (tool middleware chain, skills, MCP/A2A), and the
economy layer being optional but first-class.

## What gets removed from README

Everything below is fully duplicated in `docs/` and will be cut from README:

| Current README range | Topic | Destination already exists |
|----------------------|-------|----------------------------|
| 155–353 | CLI Commands | `docs/cli/*` (25 files) |
| 354–368 | Diagnostics | `docs/cli/status.md` |
| 369–449 | Workbench & Cockpit TUI | `docs/features/cockpit.md` |
| 450–631 | Architecture (full) | `docs/architecture/*` |
| 632–648 | AI Providers | `docs/features/ai-providers.md` |
| 649–997 | Configuration Reference (350 lines!) | `docs/configuration.md` (1563 lines) |
| 998–1101 | On-Chain Economy | `docs/payments/`, `docs/features/economy.md` |
| 1103–1174 | System Prompts | `docs/features/` (verify or stub) |
| 1175–1200 | Embedding & RAG | `docs/features/embedding-rag.md` |
| 1201–1237 | Knowledge Graph & Graph RAG | `docs/features/knowledge-graph.md` |
| 1238–1290 | Multi-Agent Orchestration | `docs/features/multi-agent.md` |
| 1291–1302 | A2A Protocol | `docs/features/a2a-protocol.md` |
| 1303–1404 | P2P Network | `docs/features/p2p-network.md` |
| 1405–1485 | Blockchain Payments | `docs/payments/usdc.md`, `docs/payments/x402.md` |
| 1486–1526 | Cron Scheduling | `docs/automation/` |
| 1527–1539 | Background Execution | `docs/automation/` |
| 1540–1602 | Workflow Engine | `docs/automation/` |
| 1603–1631 | Self-Learning System | `docs/features/learning.md` |
| 1632–1805 | Security (full) | `docs/security/*` |
| 1806–1844 | Docker | `docs/deployment/docker.md` |
| 1845+   | Examples | `docs/getting-started/` |

## Risks & mitigations

| Risk | Mitigation |
|------|------------|
| Readers used to ctrl-F'ing the README lose searchability of CLI/config details. | Documentation Map (Section 6) groups every doc page within one scrollable section, replacing ctrl-F with linkable navigation. |
| Broken `docs/…` links if a section was referenced inline but no doc exists yet. | Implementation step explicitly verifies every link with `grep -rn "(docs/" README.md` and auto-creates a one-line stub (with TODO marker) for any missing target. |
| `mkdocs.yml` references README anchors that no longer exist. | Implementation step inspects `mkdocs.yml` for `README.md` references; remove or rewrite if present. |
| Information that was in README but not yet in docs gets lost. | Spot-check the 350-line Configuration Reference (lines 649–997) against `docs/configuration.md` before deletion. Only delete if `docs/configuration.md` already contains equivalent or richer content; otherwise migrate the delta first. Apply the same check to System Prompts (1103–1174). |
| Hero/Why prose still reads like a feature dump. | Section 2 keeps the existing 10-bullet "Why Lango" block verbatim — it already focuses on differentiators (orchestration, observability, ZK, knowledge-as-currency, interoperability, P2P economy, on-chain settlement, escrow execution, smart accounts, trust & reputation) rather than feature enumeration. |

## Implementation plan (preview)

The detailed plan will be written in the next step (`writing-plans` skill). At a
high level:

1. Spot-check `docs/configuration.md` vs README config section; migrate any delta
   into the doc.
2. Spot-check System Prompts (README 1103–1174) vs `docs/features/`; create or
   extend a docs page if content is missing.
3. Rewrite `README.md` end-to-end following the eight-section structure above.
4. Run link verification: `grep -rn "(docs/" README.md` → confirm each path resolves
   to an existing file; auto-stub any missing pages with a one-line TODO placeholder.
5. Inspect `mkdocs.yml` for direct references to removed README content.
6. Commit as a single PR titled `docs: refactor README into landing page`.

## Open questions

None at design-spec time. Two were raised during brainstorming and resolved:

- **Target length and style** → ~450 lines, balanced structure.
- **Features rendering** → hybrid (category heading + emoji bullets + `→` link).
- **Missing-link handling** → auto-stub with TODO marker.

## Out of scope (do not expand)

- Restructuring `docs/` itself.
- Improving `docs/configuration.md` beyond the migration spot-check.
- Adding new features, CLI commands, or docs pages beyond stubs for broken links.
- Translating any of the new README to Korean. README stays in English per
  `.claude/CLAUDE.md`.
