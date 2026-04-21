# Zensical Native Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the MkDocs/Material documentation toolchain with a Zensical-native site, preserving the current public documentation experience while removing MkDocs as the canonical docs build path.

**Architecture:** Migrate in four layers: first create a Zensical-native site definition, then move hidden docs out of `docs/` so the public IA is represented structurally instead of through MkDocs-only exclusion features, then switch CI/docs references to the new toolchain, and finally close the slice out through OpenSpec. Preserve public nav, links, Mermaid, search, dark mode, and code-copy behavior as the compatibility contract.

**Tech Stack:** Zensical, TOML, GitHub Actions, Markdown, existing `docs/` tree, OpenSpec

---

## File Map

- Create: `zensical.toml`
  - New canonical docs site configuration.
- Delete: `mkdocs.yml`
  - Removed as the canonical docs toolchain config.
- Modify: `.github/workflows/docs.yml`
  - Replace MkDocs deployment steps with Zensical build/deploy steps.
- Modify: `README.md`
  - Replace MkDocs-specific docs-tooling references with Zensical-native references.
- Modify: `docs/architecture/project-structure.md`
  - Update the documented docs toolchain/layout so it no longer says `mkdocs.yml`.
- Modify: `docs/development/build-test.md`
  - Replace `mkdocs build`-style docs instructions with Zensical-native commands if present or add a short docs-build note if the page currently references the old toolchain.
- Move: hidden docs currently under `docs/` to an internal non-site path such as `internal-docs/`
  - `docs/superpowers/specs/**`
  - `docs/superpowers/plans/**`
  - `docs/architecture/adr-001-package-boundaries.md`
  - `docs/architecture/dependency-graph.md`
  - `docs/features/cockpit-approval-guide.md`
  - `docs/features/cockpit-channels-guide.md`
  - `docs/features/cockpit-tasks-guide.md`
  - `docs/features/cockpit-troubleshooting.md`
- Modify: `docs/features/cockpit.md`
  - Keep it as the single public cockpit entry after the hidden guide move.
- Create: `openspec/changes/migrate-docs-toolchain-to-zensical/**`
  - Proposal, design, tasks, and delta specs for the migration slice.
- Modify: `openspec/specs/mkdocs-documentation-site/spec.md`
  - Replace/extend with a toolchain-neutral or Zensical-native docs-site requirement set.
- Modify or create: `openspec/specs/project-docs/spec.md`
  - Capture hidden-doc relocation and public-vs-internal docs boundary.
- Modify: `openspec/specs/docs-only/spec.md`
  - Capture updated docs build references and public cockpit entry expectations.

### Task 1: Establish The Zensical-Native Site Definition

**Files:**
- Create: `zensical.toml`
- Delete: `mkdocs.yml`

- [ ] **Step 1: Capture the current public docs contract before changing tooling**

Run:

```bash
python3 -m mkdocs build --strict
python3 - <<'PY'
import pathlib
print(pathlib.Path('mkdocs.yml').read_text())
PY
```

Expected:

```text
MkDocs build succeeds and `mkdocs.yml` is available as the source of the current public nav/theme/features contract.
```

- [ ] **Step 2: Write a first Zensical-native config in `zensical.toml`**

Create `zensical.toml` with this initial structure:

```toml
[project]
site_name = "Lango"
site_url = "https://langoai.github.io/lango/"
site_description = "A high-performance AI agent built with Go"
repo_name = "langoai/lango"
repo_url = "https://github.com/langoai/lango"
edit_uri = "edit/main/docs/"
copyright = "Copyright &copy; 2026 langoai - MIT License"
docs_dir = "docs"
site_dir = "site"
extra_css = ["stylesheets/extra.css"]

nav = [
  { "Home" = "index.md" },
  { "Getting Started" = [
      "getting-started/index.md",
      { "Installation" = "getting-started/installation.md" },
      { "Quick Start" = "getting-started/quickstart.md" },
      { "Configuration Basics" = "getting-started/configuration.md" }
  ]},
  { "Architecture" = [
      "architecture/index.md",
      { "System Overview" = "architecture/overview.md" },
      { "Project Structure" = "architecture/project-structure.md" },
      { "Data Flow" = "architecture/data-flow.md" },
      { "Master Document" = "architecture/master-document.md" },
      { "External Collaboration Audit" = "architecture/external-collaboration-audit.md" },
      { "Trust, Security & Policy Audit" = "architecture/trust-security-policy-audit.md" },
      { "P2P Knowledge Exchange Track" = "architecture/p2p-knowledge-exchange-track.md" }
  ]},
  { "Features" = [
      "features/index.md",
      { "AI Providers" = "features/ai-providers.md" },
      { "Channels" = "features/channels.md" },
      { "Knowledge System" = "features/knowledge.md" },
      { "Observational Memory" = "features/observational-memory.md" },
      { "Embedding & RAG" = "features/embedding-rag.md" },
      { "Knowledge Graph" = "features/knowledge-graph.md" },
      { "Knowledge Ontology" = "features/ontology.md" },
      { "Multi-Agent Orchestration" = "features/multi-agent.md" },
      { "A2A Protocol" = "features/a2a-protocol.md" },
      { "P2P Network" = "features/p2p-network.md" },
      { "P2P Economy" = "features/economy.md" },
      { "Smart Contracts" = "features/contracts.md" },
      { "Observability" = "features/observability.md" },
      { "Skill System" = "features/skills.md" },
      { "Proactive Librarian" = "features/librarian.md" },
      { "System Prompts" = "features/system-prompts.md" },
      { "Smart Accounts" = "features/smart-accounts.md" },
      { "Config Presets" = "features/config-presets.md" },
      { "ZKP" = "features/zkp.md" },
      { "Learning" = "features/learning.md" },
      { "Agent Format" = "features/agent-format.md" },
      { "MCP Integration" = "features/mcp-integration.md" },
      { "RunLedger (Task OS)" = "features/run-ledger.md" },
      { "Session Provenance" = "features/provenance.md" },
      { "Operational Alerting" = "features/alerting.md" },
      { "Exec Safety" = "features/exec-safety.md" },
      { "Cockpit TUI" = "features/cockpit.md" }
  ]},
  { "Automation" = [
      "automation/index.md",
      { "Cron Scheduling" = "automation/cron.md" },
      { "Background Tasks" = "automation/background.md" },
      { "Workflow Engine" = "automation/workflows.md" }
  ]},
  { "Security" = [
      "security/index.md",
      { "Encryption & Secrets" = "security/encryption.md" },
      { "Envelope Migration" = "security/envelope-migration.md" },
      { "PII Redaction" = "security/pii-redaction.md" },
      { "Exportability Policy" = "security/exportability.md" },
      { "Approval Flow" = "security/approval-flow.md" },
      { "Approval CLI" = "security/approval-cli.md" },
      { "Upfront Payment Approval" = "security/upfront-payment-approval.md" },
      { "Escrow Execution" = "security/escrow-execution.md" },
      { "Actual Payment Execution Gating" = "security/actual-payment-execution-gating.md" },
      { "Dispute-Ready Receipts" = "security/dispute-ready-receipts.md" },
      { "Tool Approval" = "security/tool-approval.md" },
      { "Authentication" = "security/authentication.md" }
  ]},
  { "Research" = [
      { "Phase 7 PQ On-Chain Feasibility" = "research/phase7-pq-onchain-feasibility.md" }
  ]},
  { "Payments" = [
      "payments/index.md",
      { "USDC Payments" = "payments/usdc.md" },
      { "X402 Protocol" = "payments/x402.md" }
  ]},
  { "CLI Reference" = [
      "cli/index.md",
      { "Core Commands" = "cli/core.md" },
      { "Config Management" = "cli/config.md" },
      { "Agent & Memory" = "cli/agent-memory.md" },
      { "A2A Commands" = "cli/a2a.md" },
      { "Security Commands" = "cli/security.md" },
      { "Payment Commands" = "cli/payment.md" },
      { "P2P Commands" = "cli/p2p.md" },
      { "Economy Commands" = "cli/economy.md" },
      { "Contract Commands" = "cli/contract.md" },
      { "Metrics Commands" = "cli/metrics.md" },
      { "Automation Commands" = "cli/automation.md" },
      { "Status Dashboard" = "cli/status.md" },
      { "MCP Commands" = "cli/mcp.md" },
      { "Smart Account Commands" = "cli/smartaccount.md" },
      { "Provenance Commands" = "cli/provenance.md" },
      { "RunLedger Commands" = "cli/run.md" },
      { "Sandbox Commands" = "cli/sandbox.md" },
      { "Alerts Commands" = "cli/alerts.md" },
      { "Learning Commands" = "cli/learning.md" }
  ]},
  { "Gateway & API" = [
      "gateway/index.md",
      { "HTTP API" = "gateway/http-api.md" },
      { "WebSocket" = "gateway/websocket.md" }
  ]},
  { "Deployment" = [
      "deployment/index.md",
      { "Docker" = "deployment/docker.md" },
      { "Production Checklist" = "deployment/production.md" }
  ]},
  { "Development" = [
      "development/index.md",
      { "Build & Test" = "development/build-test.md" }
  ]},
  { "Configuration Reference" = "configuration.md" }
]

[project.theme]
logo = "assets/logo.png"
favicon = "assets/logo.png"
features = [
  "navigation.tabs",
  "navigation.tabs.sticky",
  "navigation.sections",
  "navigation.expand",
  "navigation.path",
  "navigation.top",
  "navigation.indexes",
  "navigation.footer",
  "search.highlight",
  "content.code.copy",
  "content.tabs.link",
  "toc.follow",
]

[project.theme.icon]
repo = "fontawesome/brands/github"

[[project.theme.palette]]
media = "(prefers-color-scheme: light)"
scheme = "default"
primary = "deep purple"
accent = "amber"
toggle.icon = "lucide/sun"
toggle.name = "Switch to dark mode"

[[project.theme.palette]]
media = "(prefers-color-scheme: dark)"
scheme = "slate"
primary = "deep purple"
accent = "amber"
toggle.icon = "lucide/moon"
toggle.name = "Switch to light mode"

[project.extra]
generator = false

[project.markdown_extensions.abbr]
[project.markdown_extensions.admonition]
[project.markdown_extensions.attr_list]
[project.markdown_extensions.def_list]
[project.markdown_extensions.footnotes]
[project.markdown_extensions.md_in_html]
[project.markdown_extensions.tables]

[project.markdown_extensions.toc]
permalink = true
toc_depth = 3

[project.markdown_extensions.pymdownx.details]
[project.markdown_extensions.pymdownx.inlinehilite]
[project.markdown_extensions.pymdownx.keys]
[project.markdown_extensions.pymdownx.mark]
[project.markdown_extensions.pymdownx.tabbed]
alternate_style = true

[project.markdown_extensions.pymdownx.tasklist]
custom_checkbox = true

[project.markdown_extensions.pymdownx.highlight]
line_spans = "__span"
pygments_lang_class = true

[project.markdown_extensions.pymdownx.emoji]
emoji_index = "zensical.extensions.emoji.twemoji"
emoji_generator = "zensical.extensions.emoji.to_svg"

[project.markdown_extensions.pymdownx.superfences]
custom_fences = [
  { name = "mermaid", class = "mermaid", format = "pymdownx.superfences.fence_code_format" }
]
```

- [ ] **Step 3: Write the failing migration smoke test as commands**

Run:

```bash
zensical build
```

Expected:

```text
FAIL until Zensical is installed and hidden docs are moved out of docs/.
```

- [ ] **Step 4: Remove the old canonical config**

Delete `mkdocs.yml`.

- [ ] **Step 5: Commit the native config slice**

Run:

```bash
git add zensical.toml mkdocs.yml
git -c commit.gpgsign=false commit -m "docs: add zensical site config"
```

### Task 2: Move Hidden Docs Out Of The Public Docs Tree

**Files:**
- Create: `internal-docs/`
- Move:
  - `docs/superpowers/specs/**`
  - `docs/superpowers/plans/**`
  - `docs/architecture/adr-001-package-boundaries.md`
  - `docs/architecture/dependency-graph.md`
  - `docs/features/cockpit-approval-guide.md`
  - `docs/features/cockpit-channels-guide.md`
  - `docs/features/cockpit-tasks-guide.md`
  - `docs/features/cockpit-troubleshooting.md`

- [ ] **Step 1: Create the internal-docs destination layout**

Run:

```bash
mkdir -p internal-docs/superpowers/specs
mkdir -p internal-docs/superpowers/plans
mkdir -p internal-docs/architecture
mkdir -p internal-docs/features
```

Expected:

```text
The internal-docs tree exists and mirrors the hidden-source layout.
```

- [ ] **Step 2: Move the hidden documents**

Run:

```bash
mv docs/superpowers/specs/* internal-docs/superpowers/specs/
mv docs/superpowers/plans/* internal-docs/superpowers/plans/
mv docs/architecture/adr-001-package-boundaries.md internal-docs/architecture/
mv docs/architecture/dependency-graph.md internal-docs/architecture/
mv docs/features/cockpit-approval-guide.md internal-docs/features/
mv docs/features/cockpit-channels-guide.md internal-docs/features/
mv docs/features/cockpit-tasks-guide.md internal-docs/features/
mv docs/features/cockpit-troubleshooting.md internal-docs/features/
```

Expected:

```text
Hidden documents no longer live under docs/.
```

- [ ] **Step 3: Re-run a Zensical build to verify the docs tree shape**

Run:

```bash
zensical build
```

Expected:

```text
The public docs tree no longer contains hidden pages that need exclusion-specific handling.
```

- [ ] **Step 4: Commit the hidden-doc relocation**

Run:

```bash
git add docs internal-docs
git -c commit.gpgsign=false commit -m "docs: move hidden docs out of public tree"
```

### Task 3: Restore Public Docs Feature Parity Under Zensical

**Files:**
- Modify: `zensical.toml`
- Modify: `docs/features/cockpit.md`
- Modify: `docs/architecture/project-structure.md`
- Modify: `docs/development/build-test.md`
- Modify: `README.md`

- [ ] **Step 1: Write the parity checklist as verification commands**

Run:

```bash
zensical build
```

Expected manual checks:

```text
- nav renders the current public IA
- search is available
- dark mode is available
- code copy buttons appear on code blocks
- Mermaid diagrams render
- cockpit public page still includes the consolidated operator guidance
```

- [ ] **Step 2: Update repository docs references away from MkDocs**

Update `docs/architecture/project-structure.md`:

```md
├── docs/                   # Public documentation source
├── internal-docs/          # Internal design records, plans, and hidden support docs
└── zensical.toml           # Zensical documentation site configuration
```

Update any MkDocs-specific line in `docs/development/build-test.md` to the new docs commands:

```md
## Documentation

Build the documentation site with:

```bash
zensical build
```

Preview the site locally with:

```bash
zensical serve
```
```

Update README doc-tooling wording minimally, for example:

```md
The documentation site is built with Zensical and sourced from `docs/`.
```

- [ ] **Step 3: Keep cockpit public entry intact after the hidden-guide move**

Review `docs/features/cockpit.md` and ensure the consolidated sections from the previous IA recovery remain in place:

```md
## Approval Operations
...
## Channel Operations
...
## Background Task Operations
...
## Troubleshooting
...
```

If any moved hidden guide content left broken links, fix those links to public canonical pages.

- [ ] **Step 4: Run the docs parity verification**

Run:

```bash
zensical build
go build ./...
go test ./...
```

Expected:

```text
All commands exit 0.
```

- [ ] **Step 5: Commit the parity/refs slice**

Run:

```bash
git add zensical.toml docs/features/cockpit.md docs/architecture/project-structure.md docs/development/build-test.md README.md
git -c commit.gpgsign=false commit -m "docs: migrate public docs to zensical"
```

### Task 4: Switch CI And Deployment Workflow To Zensical

**Files:**
- Modify: `.github/workflows/docs.yml`

- [ ] **Step 1: Replace the MkDocs deployment workflow**

Rewrite `.github/workflows/docs.yml` to the Zensical GitHub Pages flow:

```yaml
name: Documentation

on:
  push:
    branches:
      - main
    paths:
      - docs/**
      - internal-docs/**
      - zensical.toml
      - .github/workflows/docs.yml

permissions:
  contents: read
  pages: write
  id-token: write

jobs:
  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    steps:
      - uses: actions/configure-pages@v5
      - uses: actions/checkout@v5
      - uses: actions/setup-python@v5
        with:
          python-version: 3.x
      - run: pip install zensical
      - run: zensical build --clean
      - uses: actions/upload-pages-artifact@v4
        with:
          path: site
      - uses: actions/deploy-pages@v4
        id: deployment
```

- [ ] **Step 2: Validate the workflow file**

Run:

```bash
sed -n '1,220p' .github/workflows/docs.yml
```

Expected:

```text
No MkDocs install or `mkdocs gh-deploy` steps remain.
```

- [ ] **Step 3: Commit the CI migration**

Run:

```bash
git add .github/workflows/docs.yml
git -c commit.gpgsign=false commit -m "ci: migrate docs workflow to zensical"
```

### Task 5: OpenSpec Change, Main Spec Sync, And Archive

**Files:**
- Create: `openspec/changes/migrate-docs-toolchain-to-zensical/proposal.md`
- Create: `openspec/changes/migrate-docs-toolchain-to-zensical/design.md`
- Create: `openspec/changes/migrate-docs-toolchain-to-zensical/tasks.md`
- Create: `openspec/changes/migrate-docs-toolchain-to-zensical/specs/mkdocs-documentation-site/spec.md`
- Create: `openspec/changes/migrate-docs-toolchain-to-zensical/specs/project-docs/spec.md`
- Create: `openspec/changes/migrate-docs-toolchain-to-zensical/specs/docs-only/spec.md`
- Modify: `openspec/specs/mkdocs-documentation-site/spec.md`
- Modify: `openspec/specs/project-docs/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`

- [ ] **Step 1: Write the OpenSpec change artifacts**

Create `openspec/changes/migrate-docs-toolchain-to-zensical/proposal.md`:

```md
## Why

The current docs toolchain depends on MkDocs 1.x and Material for MkDocs, which now emits upstream compatibility warnings and no longer represents a stable long-term base. The project needs a native Zensical toolchain instead of a warning-suppression strategy.

## What Changes

- replace the canonical docs config with Zensical
- move hidden docs out of the public docs tree
- migrate local and CI docs build paths
- preserve the public documentation experience and current IA

## Impact

- docs build configuration
- GitHub Actions docs deployment
- public vs internal docs boundaries
- repository docs-tooling references
```

Create `openspec/changes/migrate-docs-toolchain-to-zensical/specs/mkdocs-documentation-site/spec.md`:

```md
## ADDED Requirements

### Requirement: Canonical docs site toolchain is Zensical
The project SHALL define its documentation site through a Zensical-native configuration instead of MkDocs as the canonical docs toolchain.

#### Scenario: Zensical build is canonical
- **WHEN** project documentation is built locally or in CI
- **THEN** the build SHALL run through Zensical instead of MkDocs

### Requirement: Public docs experience is preserved during migration
The docs toolchain migration SHALL preserve the current public documentation IA and key reader-facing features.

#### Scenario: Public site parity retained
- **WHEN** the migrated docs site is built
- **THEN** navigation, links, search, dark mode, code copy, and Mermaid support SHALL remain available
```

Create `openspec/changes/migrate-docs-toolchain-to-zensical/specs/project-docs/spec.md`:

```md
## ADDED Requirements

### Requirement: Hidden docs live outside the public docs tree
Internal design records and hidden support docs SHALL no longer live under the public `docs/` source tree.

#### Scenario: Hidden docs relocated
- **WHEN** the repository docs layout is inspected
- **THEN** internal worklogs, hidden architecture notes, and hidden cockpit guides SHALL live outside `docs/`
```

Create `openspec/changes/migrate-docs-toolchain-to-zensical/specs/docs-only/spec.md`:

```md
## ADDED Requirements

### Requirement: Repository docs references describe Zensical
Repository-facing documentation SHALL describe the docs toolchain as Zensical-based after migration.

#### Scenario: Project structure doc updated
- **WHEN** a reader inspects docs/tooling references
- **THEN** they SHALL see `zensical.toml` and the updated public-vs-internal docs layout instead of `mkdocs.yml`

### Requirement: Cockpit public page remains the single public entry
The migrated docs tree SHALL keep `docs/features/cockpit.md` as the single public cockpit entry after the hidden cockpit sub-guides are moved out of the public docs tree.

#### Scenario: Cockpit entry remains public
- **WHEN** a reader browses the Features section
- **THEN** `features/cockpit.md` SHALL remain the public cockpit page
```

- [ ] **Step 2: Sync the main specs**

Copy the new requirements into the main specs:

```bash
cp openspec/changes/migrate-docs-toolchain-to-zensical/specs/mkdocs-documentation-site/spec.md openspec/specs/mkdocs-documentation-site/spec.md
cp openspec/changes/migrate-docs-toolchain-to-zensical/specs/project-docs/spec.md openspec/specs/project-docs/spec.md
cp openspec/changes/migrate-docs-toolchain-to-zensical/specs/docs-only/spec.md openspec/specs/docs-only/spec.md
```

- [ ] **Step 3: Archive the change**

Run:

```bash
mkdir -p openspec/changes/archive
mv openspec/changes/migrate-docs-toolchain-to-zensical openspec/changes/archive/2026-04-21-migrate-docs-toolchain-to-zensical
git add openspec/specs/mkdocs-documentation-site/spec.md openspec/specs/project-docs/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-21-migrate-docs-toolchain-to-zensical
git -c commit.gpgsign=false commit -m "specs: archive zensical migration"
```

## Self-Review

- Spec coverage:
  - Zensical-native canonical config: Task 1
  - hidden docs moved out of public tree: Task 2
  - public docs parity and toolchain references: Task 3
  - CI/docs deployment switch: Task 4
  - OpenSpec closeout: Task 5
- Placeholder scan:
  - no `TODO`, `TBD`, or missing implementation steps remain
- Type/tool consistency:
  - the plan consistently uses `zensical.toml`, `zensical build`, and `zensical serve`
  - hidden docs are structurally relocated instead of relying on MkDocs-only exclusion features unsupported by Zensical
