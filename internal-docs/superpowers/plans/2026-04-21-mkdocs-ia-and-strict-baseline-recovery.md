# MkDocs IA And Strict Baseline Recovery Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore a passing `python3 -m mkdocs build --strict` baseline while reshaping the public documentation site to expose only intentionally published user/operator docs.

**Architecture:** Use MkDocs 1.6.1's native `exclude_docs` support to remove hidden content from the built site, expand `nav` only for explicitly public documents, and consolidate cockpit sub-guides into the main cockpit entry page. Keep the slice documentation-only, then close it out through the standard OpenSpec delta-sync-and-archive flow.

**Tech Stack:** MkDocs 1.6.1, Material for MkDocs, Markdown, OpenSpec

---

## File Map

- `mkdocs.yml`
  - Add `exclude_docs` patterns for hidden content.
  - Add `Security` entries for `approval-cli` and `envelope-migration`.
  - Add a top-level `Research` section.
- `docs/getting-started/quickstart.md`
  - Fix the broken installation anchor link.
- `docs/features/cockpit.md`
  - Absorb the operator-facing material now stranded in the cockpit sub-guides.
- `docs/security/index.md`
  - No broad rewrite expected, but verify links remain coherent after nav changes.
- `openspec/changes/recover-mkdocs-ia-and-strict-baseline/**`
  - Proposal, design, tasks, and delta specs for the documentation-site recovery slice.
- `openspec/specs/mkdocs-documentation-site/spec.md`
  - Sync new requirements for exclusion-aware IA and strict baseline recovery.
- `openspec/specs/security-docs-sync/spec.md`
  - Sync requirements for the newly surfaced security deep-dive docs.
- `openspec/specs/docs-only/spec.md`
  - Sync the quickstart anchor fix and cockpit public-entry consolidation requirements.

### Task 1: Rebuild MkDocs Navigation Around Public Docs Only

**Files:**
- Modify: `mkdocs.yml`
- Modify: `docs/getting-started/quickstart.md`

- [ ] **Step 1: Reproduce the current strict failure**

Run:

```bash
python3 -m mkdocs build --strict
```

Expected:

```text
INFO    -  The following pages exist in the docs directory, but are not included in the "nav" configuration:
...
INFO    -  Doc file 'getting-started/quickstart.md' contains a link 'installation.md#platform-specific-c-compiler-setup' ...
```

- [ ] **Step 2: Add explicit exclusion rules and public nav entries**

Update `mkdocs.yml` by adding top-level exclusion rules directly under `markdown_extensions`/before `nav`:

```yaml
exclude_docs: |
  .DS_Store
  /superpowers/specs/**
  /superpowers/plans/**
  /architecture/adr-001-package-boundaries.md
  /architecture/dependency-graph.md
  /features/cockpit-approval-guide.md
  /features/cockpit-channels-guide.md
  /features/cockpit-tasks-guide.md
  /features/cockpit-troubleshooting.md
```

Extend `nav` with the intended public additions:

```yaml
  - Security:
    - security/index.md
    - Encryption & Secrets: security/encryption.md
    - Envelope Migration: security/envelope-migration.md
    - PII Redaction: security/pii-redaction.md
    - Exportability Policy: security/exportability.md
    - Approval Flow: security/approval-flow.md
    - Approval CLI: security/approval-cli.md
    - Upfront Payment Approval: security/upfront-payment-approval.md
    - Escrow Execution: security/escrow-execution.md
    - Actual Payment Execution Gating: security/actual-payment-execution-gating.md
    - Dispute-Ready Receipts: security/dispute-ready-receipts.md
    - Tool Approval: security/tool-approval.md
    - Authentication: security/authentication.md
```

Add a new top-level `Research` section:

```yaml
  - Research:
    - phase7-pq-onchain-feasibility.md: research/phase7-pq-onchain-feasibility.md
```

- [ ] **Step 3: Fix the broken quickstart anchor**

Update the installation link in `docs/getting-started/quickstart.md`:

```md
See [Installation](installation.md) for detailed instructions, including [optional C compiler setup](installation.md#optional-c-compiler-setup) required for legacy CGO integrations.
```

- [ ] **Step 4: Re-run MkDocs strict and verify nav/exclusion warnings changed**

Run:

```bash
python3 -m mkdocs build --strict
```

Expected:

```text
No warnings about superpowers/, hidden architecture docs, or cockpit sub-guides being present but missing from nav.
No quickstart installation anchor warning.
```

- [ ] **Step 5: Commit the IA/config recovery slice**

Run:

```bash
git add mkdocs.yml docs/getting-started/quickstart.md
git -c commit.gpgsign=false commit -m "docs: recover mkdocs nav baseline"
```

### Task 2: Consolidate Cockpit Sub-Guides Into The Main Cockpit Page

**Files:**
- Modify: `docs/features/cockpit.md`

- [ ] **Step 1: Review the current cockpit sub-guides before editing**

Run:

```bash
sed -n '1,220p' internal-docs/features/cockpit-approval-guide.md
sed -n '1,220p' internal-docs/features/cockpit-channels-guide.md
sed -n '1,220p' internal-docs/features/cockpit-tasks-guide.md
sed -n '1,220p' internal-docs/features/cockpit-troubleshooting.md
```

Expected:

```text
Four operator guides describing approvals, channels, task operations, and troubleshooting.
```

- [ ] **Step 2: Add operator-facing summary sections to `docs/features/cockpit.md`**

Append or merge these sections into `docs/features/cockpit.md` after the existing page-level behavior sections:

```md
## Approval Operations

Cockpit approval handling is centered in the Chat page. When a tool invocation requires approval, the cockpit automatically switches to the Chat page and renders either the inline strip or fullscreen dialog depending on tool risk. Operators respond with `a`, `s`, or `d`, and critical-risk tools require the existing double-press confirmation.

For the full security policy model, see [Tool Approval](../security/tool-approval.md) and [Approval CLI](../security/approval-cli.md).

## Channel Operations

When launched with channel support, the cockpit acts as a live operator console for Telegram, Discord, and Slack. Channel messages are routed into the chat transcript through the EventBus and remain visible even if the operator is browsing another page. Approval requests originating from channels also surface in the cockpit.

Do not run `lango cockpit --with-channels` and `lango serve` at the same time with the same channel credentials.

## Background Task Operations

The Tasks page and chat footer strip expose background task progress. Operators can inspect task state, open details, cancel running work, and retry failed or cancelled tasks directly from the cockpit. The page refreshes automatically and uses the same task lifecycle as the background automation subsystem.

For the system-level background task reference, see [Background Tasks](../automation/background.md).

## Troubleshooting

Common cockpit issues fall into four groups:

- startup and TTY issues
- context panel or runtime visibility issues
- approval or channel routing issues
- rendering/logging issues

Use `lango doctor` first for environment checks, and inspect `~/.lango/cockpit.log` when behavior is unclear.
```

- [ ] **Step 3: Re-run MkDocs strict to confirm cockpit coverage is sufficient**

Run:

```bash
python3 -m mkdocs build --strict
```

Expected:

```text
The excluded cockpit sub-guides no longer trigger not-in-nav warnings, and the public cockpit page still builds cleanly.
```

- [ ] **Step 4: Commit the cockpit consolidation**

Run:

```bash
git add docs/features/cockpit.md
git -c commit.gpgsign=false commit -m "docs: consolidate cockpit operator guidance"
```

### Task 3: Verify Security And Research IA End-To-End

**Files:**
- Modify: `docs/security/index.md` (only if link text or ordering needs truth-alignment after nav changes)
- Modify: `README.md` (only if the docs-site section or doc references now need a short truthful note)

- [ ] **Step 1: Inspect the public IA after the nav changes**

Run:

```bash
python3 -m mkdocs build --strict
```

Expected:

```text
Build succeeds with Security and Research sections present in the generated site.
```

- [ ] **Step 2: Only if necessary, truth-align the security index or README**

If the new IA leaves awkward references, make the smallest necessary edits. For example, keep the security index quick links coherent with the newly public deep-dive docs:

```md
- [Approval CLI](approval-cli.md) -- Operational approval-system behavior in the CLI/TUI surface
- [Envelope Migration](envelope-migration.md) -- Master-key envelope migration and recovery details
```

Do not add broad new documentation claims unrelated to IA recovery.

- [ ] **Step 3: Run full repository verification**

Run:

```bash
go build ./...
go test ./...
python3 -m mkdocs build --strict
```

Expected:

```text
All three commands exit 0.
```

- [ ] **Step 4: Commit any final truth-alignment edits**

Run:

```bash
git add docs/security/index.md README.md
git -c commit.gpgsign=false commit -m "docs: finish mkdocs ia truth alignment"
```

If no edits were needed in this task, skip this commit.

### Task 4: OpenSpec Change, Main Spec Sync, And Archive

**Files:**
- Create: `openspec/changes/recover-mkdocs-ia-and-strict-baseline/proposal.md`
- Create: `openspec/changes/recover-mkdocs-ia-and-strict-baseline/design.md`
- Create: `openspec/changes/recover-mkdocs-ia-and-strict-baseline/tasks.md`
- Create: `openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/mkdocs-documentation-site/spec.md`
- Create: `openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/security-docs-sync/spec.md`
- Create: `openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/docs-only/spec.md`
- Modify: `openspec/specs/mkdocs-documentation-site/spec.md`
- Modify: `openspec/specs/security-docs-sync/spec.md`
- Modify: `openspec/specs/docs-only/spec.md`

- [ ] **Step 1: Write the OpenSpec change artifacts**

Create `openspec/changes/recover-mkdocs-ia-and-strict-baseline/proposal.md`:

```md
## Why

The MkDocs site currently fails `--strict` because hidden/internal docs are still present in the build tree without being intentionally surfaced or excluded. The site navigation also omits a small set of public security and research documents that should be intentionally published.

## What Changes

- restore a passing strict MkDocs baseline
- exclude internal docs from the built site
- publish selected security deep-dive docs
- add a research section
- consolidate cockpit operator guidance into the main cockpit page
- fix the broken quickstart installation anchor

## Impact

- `mkdocs.yml`
- selected docs under `docs/`
- documentation-site specs
```

Create `openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/mkdocs-documentation-site/spec.md`:

```md
## ADDED Requirements

### Requirement: MkDocs public site excludes hidden internal documents
The MkDocs site SHALL exclude internal worklog and hidden support documents from the built site instead of leaving them as unlisted source files.

#### Scenario: Hidden docs excluded from build
- **WHEN** `python3 -m mkdocs build --strict` runs
- **THEN** hidden internal documents SHALL not be reported as present-but-not-in-nav pages

### Requirement: Public IA includes security deep-dives and research
The MkDocs site SHALL expose explicitly chosen public docs under `Security` and `Research`.

#### Scenario: Security deep-dives published
- **WHEN** a user browses the Security section
- **THEN** `approval-cli.md` and `envelope-migration.md` SHALL be publicly discoverable

#### Scenario: Research section published
- **WHEN** a user browses the site navigation
- **THEN** a `Research` section SHALL exist and include the phase 7 PQ feasibility document
```

Create `openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/docs-only/spec.md`:

```md
## ADDED Requirements

### Requirement: Quickstart installation anchor is valid
The getting-started quickstart page SHALL link to the actual installation anchor that exists on the installation page.

#### Scenario: Quickstart installation link resolves
- **WHEN** a user follows the installation deep link from `quickstart.md`
- **THEN** the target anchor SHALL exist on `installation.md`

### Requirement: Cockpit public entry consolidates operator guidance
The public cockpit feature page SHALL summarize approval, channel, task, and troubleshooting guidance without requiring separately published cockpit sub-guides.

#### Scenario: Cockpit page remains primary public entry
- **WHEN** a user reads `features/cockpit.md`
- **THEN** they SHALL find concise operator guidance for approvals, channels, tasks, and troubleshooting on that page
```

Create `openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/security-docs-sync/spec.md`:

```md
## ADDED Requirements

### Requirement: Security navigation includes operational deep-dive docs
The public security documentation set SHALL expose selected operator-facing deep-dive documents.

#### Scenario: Approval CLI doc linked from security nav
- **WHEN** a user browses the Security section
- **THEN** the Approval CLI document SHALL be publicly discoverable

#### Scenario: Envelope migration doc linked from security nav
- **WHEN** a user browses the Security section
- **THEN** the envelope migration document SHALL be publicly discoverable
```

- [ ] **Step 2: Sync the main specs**

Copy the delta requirements into the main specs:

```bash
mkdir -p openspec/specs/mkdocs-documentation-site
cp openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/mkdocs-documentation-site/spec.md openspec/specs/mkdocs-documentation-site/spec.md
cp openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/security-docs-sync/spec.md openspec/specs/security-docs-sync/spec.md
cp openspec/changes/recover-mkdocs-ia-and-strict-baseline/specs/docs-only/spec.md openspec/specs/docs-only/spec.md
```

Expected:

```text
Updated main specs reflect the MkDocs IA recovery slice.
```

- [ ] **Step 3: Archive the change and verify final status**

Run:

```bash
mkdir -p openspec/changes/archive
mv openspec/changes/recover-mkdocs-ia-and-strict-baseline openspec/changes/archive/2026-04-21-recover-mkdocs-ia-and-strict-baseline
git add openspec/specs/mkdocs-documentation-site/spec.md openspec/specs/security-docs-sync/spec.md openspec/specs/docs-only/spec.md openspec/changes/archive/2026-04-21-recover-mkdocs-ia-and-strict-baseline
git -c commit.gpgsign=false commit -m "specs: archive mkdocs ia recovery"
```

Expected:

```text
Archived change present under openspec/changes/archive/ and main specs updated.
```

## Self-Review

- Spec coverage:
  - public-vs-hidden IA enforcement: Task 1 and Task 4
  - cockpit consolidation: Task 2
  - broken quickstart anchor: Task 1
  - strict baseline validation: Task 1 and Task 3
  - OpenSpec closeout: Task 4
- Placeholder scan:
  - no `TODO`, `TBD`, or implementation gaps remain in the plan
- Type/config consistency:
  - MkDocs uses `exclude_docs` in version 1.6.1
  - the planned hidden-set matches the actual strict warnings currently emitted
