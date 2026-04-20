## Context

The master document, P2P knowledge-exchange track, and trust/security audit now all require explicit exportability boundaries for `knowledge exchange v1`. The codebase already has several reusable building blocks:

- knowledge assets with `key`, `tags`, and `source`,
- audit storage for append-only decision records,
- provenance and receipt-oriented documentation,
- security configuration and CLI status surfaces.

What is missing is the policy layer that connects these pieces into one artifact-level decision model.

## Goals / Non-Goals

**Goals:**

- Add a source-primary exportability evaluator for artifact decisions.
- Persist minimum source metadata needed for lineage-aware evaluation.
- Emit durable decision receipts in a form that later approval/dispute flows can reuse.
- Expose the first-slice policy state through minimal operator-facing surfaces.

**Non-Goals:**

- Full user-authored policy-rule DSL.
- Human override UI/workflow.
- Provenance-bundle embedding of exportability receipts.
- Full dispute-ready receipt unification across provenance, audit, and settlement.
- Asset tagging for non-knowledge stores in this slice.

## Decisions

### 1. Use a new `internal/exportability` package

The evaluator logic is new cross-cutting policy behavior and should not be embedded inside `knowledge`, `approval`, or `provenance`. A focused package keeps the first slice narrow and reusable by later approval or receipt work.

Alternative considered:

- put the evaluator inside `internal/knowledge`
  - rejected because exportability is a policy concern, not a knowledge-storage concern

### 2. Persist source class on `knowledge` first

The codebase already has a real persisted knowledge store plus an agent-facing `save_knowledge` surface. Using that store as the first asset registry keeps the slice implementable without inventing a separate asset system.

Alternative considered:

- create a new generic asset registry
  - rejected as too broad for the first slice

### 3. Store receipts in audit log first

The audit log is already append-only and suited for durable decision records. It is the lightest-weight place to persist first-slice exportability receipts before broader provenance unification.

Alternative considered:

- store receipts only in provenance
  - rejected because provenance export/import is broader, more specialized, and not yet the canonical operator-facing receipt surface

### 4. Keep evaluation source-primary and intentionally strict

This slice blocks artifacts if any source is `private-confidential`, and escalates incomplete metadata to `needs-human-review`. This matches the design's root principle and avoids pretending that sanitization alone is enough.

Alternative considered:

- content-based override in the first slice
  - rejected because it weakens the policy boundary too early

## Risks / Trade-offs

- **[Risk]** Knowledge-source tagging may look more authoritative than it really is if users assume it covers every asset type.
  - **Mitigation:** Document clearly that this slice covers knowledge assets first, not all artifacts in the system.

- **[Risk]** Versioning knowledge entries on source-class changes may create more versions than before.
  - **Mitigation:** Keep the new version rule narrow: only classification metadata changes that affect exportability semantics create a new version.

- **[Risk]** Audit-log receipts alone are not yet full dispute-ready receipts.
  - **Mitigation:** Keep receipt wording explicit: this is a first-slice exportability receipt, not the final unified evidence model.

- **[Risk]** CLI status could imply a full policy system exists.
  - **Mitigation:** Surface only enabled state and first-slice wording; keep docs truthful about what is not implemented yet.
