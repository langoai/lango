## Context

The product now has:

- source-primary exportability decisions
- structured artifact release approval outcomes

Both are useful, but they are still point-in-time records. The next slice needs a durable receipt model that:

- tracks the canonical current state of a submission and transaction,
- keeps append-only event history,
- and references audit, provenance, and settlement context without overbuilding a full evidence graph.

## Goals / Non-Goals

**Goals:**

- Add dedicated submission and transaction receipts.
- Store canonical approval/settlement state separately from event history.
- Add a minimal first creation surface for submission receipts.
- Add truthful operator docs for the first slice.

**Non-Goals:**

- Full dispute adjudication.
- Human dispute UI.
- Full settlement execution and reconciliation.
- Full provenance embedding.
- A generalized evidence graph.

## Decisions

### 1. Use dedicated receipt models instead of audit rows

Receipts need canonical current state plus append-only history. Audit rows alone are not a good fit for that shape.

Alternative considered:

- use audit log as the only receipt model
  - rejected because it makes current-state reads and transaction grouping awkward

### 2. Start with in-memory or lightweight store semantics in the domain package

The first slice should prove the model shape before tying it deeply into all existing economic and provenance flows.

Alternative considered:

- immediately embed all fields into existing provenance storage
  - rejected as too coupled and too large for the first slice

### 3. Keep external evidence as references

This first slice stores only lite provenance and settlement references. It does not attempt to duplicate full audit/provenance payloads.

Alternative considered:

- fully embed provenance and settlement details into receipts
  - rejected because it would overbuild the first slice

## Risks / Trade-offs

- **[Risk]** The first slice may look too light compared to the name “dispute-ready”.
  - **Mitigation:** Keep docs explicit that this is a lite dispute-ready receipt, not a full dispute engine.

- **[Risk]** The receipt creation tool may be mistaken for full lifecycle automation.
  - **Mitigation:** Keep the first tool narrowly scoped to submission/transaction record creation.

- **[Risk]** Later migration from lightweight storage to fuller persistence may be needed.
  - **Mitigation:** Keep the domain types explicit and separate from audit/provenance internals.
