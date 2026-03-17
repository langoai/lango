## Context

The Lango agent cannot discover smart account tools when the subsystem is disabled. The tool catalog only registers categories for enabled subsystems, so `builtin_list` shows no trace of disabled features. The agent has no way to diagnose why tools are missing, leading to unhelpful responses like "restart the server." Additionally, `blockLangoExec` lacks a `lango account` guard, meaning the agent gets a generic catch-all message instead of being directed to specific smart account tools.

## Goals / Non-Goals

**Goals:**
- Agent can self-diagnose missing tools via `builtin_health`
- Disabled categories visible in `builtin_list` with config hints
- `lango account` CLI attempts redirected to built-in smart account tools
- Init logs include actionable remediation hints

**Non-Goals:**
- Auto-enabling disabled features
- Changing the smart account initialization logic itself
- Adding TUI/CLI commands for diagnostics (agent-only)

## Decisions

### 1. Register disabled categories in catalog
**Decision**: Add an `else` branch after `initSmartAccount()` to register a disabled `smartaccount` category with description including enable instructions.
**Rationale**: The `Category` struct already has `Enabled` and `ConfigKey` fields. `ListCategories()` returns all registered categories regardless of enabled state. This requires zero new types — just one `RegisterCategory` call.
**Alternative**: A separate "disabled features" registry — rejected as unnecessary complexity when the catalog already supports this.

### 2. New `builtin_health` diagnostic tool
**Decision**: Add a third tool to `BuildDispatcher` that reports enabled/disabled categories with config keys.
**Rationale**: `builtin_list` already shows category enabled/disabled status, but `builtin_health` provides a clearer diagnostic-focused view with explicit hints for disabled categories. This is a single function addition to `dispatcher.go`.
**Alternative**: Enhance `builtin_list` output — rejected because `builtin_list` is for discovery, not diagnostics. Separate concerns.

### 3. Orchestrator prompt diagnostics section
**Decision**: Add a short "Diagnostics" section to `buildOrchestratorInstruction` instructing the orchestrator to use `builtin_list` or `builtin_health` when tools appear missing.
**Rationale**: The orchestrator currently has no guidance for handling missing tools. A 2-line prompt addition gives it self-diagnostic capability without changing routing logic.

### 4. `lango account` guard follows existing pattern
**Decision**: Add one entry to the `guards` slice in `blockLangoExec` following the same `{prefix, feature, tools}` pattern.
**Rationale**: 11 existing guards use this exact pattern. No new code structure needed.

## Risks / Trade-offs

- [Risk] Disabled categories pollute `builtin_list` output → Mitigation: Description clearly states "(disabled)" with enable instructions; agents can filter by `enabled` field.
- [Risk] `builtin_health` adds a third dispatcher tool → Mitigation: It's a safe, read-only tool with no parameters. Minimal overhead.
