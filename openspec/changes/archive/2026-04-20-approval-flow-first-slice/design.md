## Context

The approval infrastructure in Lango is already real for tool execution, but `knowledge exchange v1` needs a product-level approval model layered above it. The recently landed exportability slice now produces structured exportability receipts, which gives approval flow a concrete upstream input.

The next slice should stay narrow:

- use `artifact release approval` as the primary approval object,
- defer full human UI and dispute orchestration,
- keep approval receipts audit-backed and operator-visible.

## Goals / Non-Goals

**Goals:**

- Add a focused approval-flow domain package.
- Introduce structured artifact release decisions (`approve`, `reject`, `request-revision`, `escalate`).
- Emit approval-flow receipts into audit storage.
- Add a minimal agent-facing approval tool and truthful docs.

**Non-Goals:**

- Full human escalation UI.
- Upfront payment approval runtime.
- Partial-settlement execution.
- Full dispute engine.

## Decisions

### 1. Start with artifact release approval only

The design includes two approval objects, but the first slice implements only `artifact release approval`. This is the real gate for `knowledge exchange v1` and keeps the slice focused.

Alternative considered:

- implement both upfront payment and artifact release approval together
  - rejected as too broad for the next slice

### 2. Use a dedicated `internal/approvalflow` package

Approval states and release outcome records are a product concept, not just tool middleware detail. A focused package avoids overloading `approval`, `toolchain`, or `app`.

Alternative considered:

- embed release approval logic directly in `tools_meta.go`
  - rejected because it would mix product policy and tool plumbing too tightly

### 3. Store first-slice approval receipts in audit log

The audit log is already the lightest durable append-only store for first-slice approval receipts. It matches the exportability slice pattern and avoids pulling provenance/dispute machinery forward too early.

Alternative considered:

- put approval receipts directly into provenance
  - rejected because it would over-couple this slice to later evidence unification work

## Risks / Trade-offs

- **[Risk]** Approval may be mistaken for a full dispute engine.
  - **Mitigation:** Keep the docs explicit that this slice only adds structured release decisions and outcome records.

- **[Risk]** Tool-level approval and product-level approval may be conflated.
  - **Mitigation:** Use a separate `approval-flow` domain package and explicit `approve_artifact_release` tool.

- **[Risk]** Audit-backed receipts may later need reshaping for broader evidence models.
  - **Mitigation:** Keep the first-slice receipt structure narrow and additive.
