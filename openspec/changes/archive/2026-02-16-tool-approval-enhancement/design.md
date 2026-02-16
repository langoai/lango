## Context

The approval system currently requires explicit opt-in (`approvalRequired: true` + `sensitiveTools` list). Without this configuration, all tools — including shell execution and file deletion — run without any approval gate. The system needs a default-safe posture where dangerous tools require approval out of the box, while remaining backward-compatible with existing configurations.

## Goals / Non-Goals

**Goals:**
- Every tool carries an intrinsic risk classification (SafetyLevel)
- New installations default to requiring approval for Dangerous tools
- Existing configurations migrate seamlessly to the new policy model
- Approval messages include actionable summaries of what will execute
- Zero-value SafetyLevel is treated as Dangerous (fail-safe principle)

**Non-Goals:**
- Per-user or per-session approval policies (future work)
- Rate-limiting or time-based auto-approval
- Dynamic SafetyLevel based on parameters (e.g., `exec rm -rf` vs `exec ls`)
- UI for managing approval policies (config-file only for now)

## Decisions

### Decision 1: SafetyLevel as iota+1 enum on Tool struct
**Choice**: Add `SafetyLevel` field to `agent.Tool` starting at `iota + 1`.
**Rationale**: Starting at 1 makes the zero value (unset) distinguishable. We treat zero as Dangerous — any tool that forgets to set its level is fail-safe.
**Alternatives**: Using string-based levels was considered but rejected for type safety and zero-value semantics.

### Decision 2: Four-mode ApprovalPolicy replacing boolean
**Choice**: `dangerous` (default), `all`, `configured`, `none`.
**Rationale**: The boolean `approvalRequired` + `sensitiveTools` list conflated two concerns: whether approval is enabled and which tools it covers. A policy enum cleanly separates intent.
**Alternatives**: Keeping the boolean and adding a `policyMode` field was considered but creates confusing interaction between two fields.

### Decision 3: ExemptTools list for opt-out
**Choice**: Add `exemptTools` config field that overrides any policy.
**Rationale**: Users need escape hatches for specific tools in automated pipelines without disabling the entire policy. ExemptTools takes priority over SensitiveTools.

### Decision 4: Summary builder as switch-case in app layer
**Choice**: `buildApprovalSummary(toolName, params)` function using switch-case per tool name.
**Rationale**: Simple, no abstraction overhead. Each tool's summary is a one-liner. Adding a new tool means adding a case. Keeps tool definitions clean (no summary template in Tool struct).
**Alternatives**: Adding a `SummaryFunc` to Tool struct was considered but adds complexity for minimal benefit.

### Decision 5: Inline migration in config loader
**Choice**: `migrateApprovalPolicy(cfg)` runs after Unmarshal, before Validate.
**Rationale**: Transparent to users — old configs work without changes. Migration logic is <20 lines and straightforward.

## Risks / Trade-offs

- **[Risk] Existing users with no config get new approval gates** → New behavior is opt-out via `approvalPolicy: "none"`. Clear in docs.
- **[Risk] Summary truncation may lose important context** → 200-char limit is generous for commands; URLs rarely exceed this.
- **[Risk] Legacy approvalRequired field remains in config struct** → Marked deprecated; migration handles it. Full removal in a future version.
- **[Trade-off] SafetyLevel is static, not context-aware** → Acceptable for v1. `exec ls` and `exec rm -rf` both require approval under `dangerous` policy. Users can exempt specific tools if needed.
