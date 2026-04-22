# Design

## Context

The knowledge-exchange track now has landed first slices for exportability, approval, payment approval, direct payment gating, dispute-ready receipts, and escrow recommendation execution.

The missing piece is not another runtime implementation detail. It is a single documentation slice that explains how those landed pieces fit together as the first transaction-oriented runtime control plane, with clear current limits.

## Goals / Non-Goals

**Goals:**

- Document the first runtime control-plane slice for `knowledge exchange v1`.
- Center the design on `transaction receipt` and `submission receipt` as the canonical runtime records.
- Make the current limits explicit so the page does not overstate implementation status.
- Keep the change documentation-only and OpenSpec-only.

**Non-Goals:**

- No new runtime package implementation.
- No human approval UI.
- No dispute orchestration.
- No generalized team execution workflow.
- No broader settlement lifecycle beyond the already-landed first slices.

## Decisions

### 1. Treat the new page as a design slice, not a finished subsystem

The page should document the first runtime control plane without claiming that a dedicated orchestration package has already landed.

Alternative considered:

- Describe the runtime as fully implemented.
  - Rejected because it would overstate the current state of the repository.

### 2. Use transaction receipt and submission receipt as the canonical runtime objects

The design should preserve the existing receipt-backed model instead of inventing new top-level records.

Alternative considered:

- Introduce a new runtime receipt hierarchy.
  - Rejected because it would duplicate the existing receipt model and blur responsibility boundaries.

### 3. Keep the current limits visible in the public architecture page

The new page should state what is not covered yet: no human UI, no dispute orchestration, no broad settlement completion, and no generalized team runtime.

Alternative considered:

- Fold limits into later implementation docs only.
  - Rejected because the architecture page would then read like a finished subsystem rather than a bounded design slice.

## Risks / Trade-offs

- [Risk] The runtime page could be read as a claim that code already exists.
  - Mitigation: State plainly that the page is a first design slice and list the current limits.

- [Risk] The docs could drift from the landed receipt-backed tools.
  - Mitigation: Keep the page anchored to the already-landed exportability, approval, payment approval, and receipt-creation slices.

- [Risk] The navigation update could look like a new product area.
  - Mitigation: Place the page under Architecture and describe it as part of the existing P2P Knowledge Exchange Track.

## Migration Plan

1. Add the runtime architecture page.
2. Wire it into the architecture index, track page, and Zensical navigation.
3. Sync the OpenSpec delta specs into the main specs.
4. Archive the completed change.

## Open Questions

None. The scope is intentionally narrow and bounded by the already-landed first slices.

