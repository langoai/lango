## Context

The Status page has `ShortHelp() == nil` and refreshes on a periodic tick after activation. Those are meaningful parts of the operator experience, but the current public docs do not say that the page is read-only or that values refresh automatically.

## Goals / Non-Goals

**Goals:**

- Document that Status is read-only and does not rely on cockpit help-bar bindings.
- Document that the page refreshes automatically while active.

**Non-Goals:**

- Change Status runtime behavior.
- Document every individual status metric in depth.

## Decisions

- Add a short interaction note in the existing Status section instead of creating a separate keys table.
  - Rationale: the absence of actionable keys is itself the important operator takeaway.

- Extend `downstream-docs-sync` with a Status interaction scenario.
  - Rationale: this keeps the public docs contract explicit and regression-resistant.

## Risks / Trade-offs

- [Docs become slightly more explicit about a non-interactive page] → This is intentional because it helps set the right operator expectation.
