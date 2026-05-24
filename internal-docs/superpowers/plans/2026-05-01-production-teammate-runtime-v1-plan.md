# Production Teammate Runtime V1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace Lango's static tool-less multi-agent contract with a v1 dynamic teammate runtime using existing `agentrt`, background projection, child-session, hook, approval, and operator-surface assets.

**Architecture:** Keep `agent_spawn`, `agent_wait`, and `agent_stop` as the only v1 model-facing control tools while extending their backing runtime state. Add teammate policy and projection fields inside `internal/agentrt`, preserve structured hook block metadata in `internal/toolchain`, then wire capability escalation and operator visibility without creating a parallel teammate package. OpenSpec leads the change and is archived only after implementation and verification; `openspec archive` is the step that updates the main specs.

**Tech Stack:** Go, Cobra CLI, OpenSpec, `internal/agentrt`, `internal/background`, `internal/toolchain`, `internal/approval`, `go test ./...`, `go build ./...`.

---

## Source Design

- Design spec: `internal-docs/superpowers/specs/2026-05-01-production-teammate-runtime-design.md`
- OpenSpec workflow reference: `.claude/guides/openspec/workflows.md`
- Teammate roles reference: `.claude/rules/teammate.md`
- Go style references: `.claude/rules/go-style.md`, `.claude/rules/go-errors.md`, `.claude/rules/go-guidelines.md`, `.claude/rules/go-patterns.md`, `.claude/rules/go-performance.md`

## Scope Check

This plan intentionally covers one production slice: in-process dynamic teammate runtime v1 with CLI-level operator visibility. It does not implement separate worker processes, sandboxed teammate execution, direct teammate messaging, or cockpit/TUI projection panels. Cockpit/TUI projection consumes the stable read model in a follow-up plan after this runtime slice ships.

## File Map

- Create `openspec/changes/production-teammate-runtime/proposal.md`: product and contract summary.
- Create `openspec/changes/production-teammate-runtime/design.md`: v1 architecture decisions and source-of-truth boundaries.
- Create `openspec/changes/production-teammate-runtime/tasks.md`: implementation checklist aligned with this plan.
- Create `openspec/changes/production-teammate-runtime/specs/multi-agent-orchestration/spec.md`: rewrite static tool-less orchestrator requirements into dynamic teammate requirements.
- Create `openspec/changes/production-teammate-runtime/specs/agent-control-plane-tools/spec.md`: extend `agent_spawn` and `agent_wait` contracts.
- Create `openspec/changes/production-teammate-runtime/specs/tool-execution-hooks/spec.md`: require structured block metadata preservation.
- Create `openspec/changes/production-teammate-runtime/specs/tool-capability-layer/spec.md`: define DynamicAllowedTools-to-capability-request routing.
- Create `openspec/changes/production-teammate-runtime/specs/cli-agent-inspection/spec.md`: require CLI visibility for teammate runtime mode and projected blocked state.
- Create `internal/agentrt/teammate_types.go`: built-in teammate type metadata, role maximum scope, and spawn-time tool validation.
- Create `internal/agentrt/teammate_types_test.go`: role scope and validation tests.
- Create `internal/agentrt/capability_policy.go`: pure policy for blocked tool attempts, grants, and approval decisions.
- Create `internal/agentrt/capability_policy_test.go`: capability decision tests.
- Create `internal/agentrt/capability_runtime.go`: connects blocked hook metadata to capability policy and run projection.
- Create `internal/agentrt/capability_runtime_test.go`: wiring tests for DynamicAllowedTools blocks.
- Modify `internal/agentrt/agent_run.go`: add projection fields without expanding the base status enum.
- Modify `internal/agentrt/agent_run_store.go`: add projection patch support and copy new fields.
- Modify `internal/agentrt/control_tools.go`: add `spawn_reason`, validate role scope, surface projected state in wait responses.
- Modify `internal/agentrt/control_tools_test.go`: cover spawn reason, role scope validation, and non-terminal wait timeout behavior.
- Modify `internal/toolchain/hooks.go`: add structured blocked tool call metadata and context sink.
- Modify `internal/toolchain/mw_hooks.go`: emit structured blocked call metadata before returning the existing error.
- Modify `internal/toolchain/mw_hooks_test.go`: verify blocked metadata includes tool, agent, session, params, and block reason.
- Modify `internal/orchestration/tools.go`: add temporary v1 selection rule for `agent_spawn` versus `transfer_to_agent`.
- Modify `internal/orchestration/orchestrator_test.go`: lock prompt guidance.
- Modify `internal/app/modules.go`: wire background submitter and capability runtime into the agent control plane.
- Modify `internal/cli/agent/status.go`: expose teammate runtime mode details.
- Modify `docs/features/multi-agent.md`: document current v1 dynamic teammate behavior accurately after code changes.

## Commit Policy

Each task includes suggested staging scope and a suggested commit message. Do not run `git commit` automatically while executing this plan; present the staging scope and message to the user after the task passes local verification so the user can commit manually.

## Task 1: OpenSpec Change Artifacts

**Files:**
- Create: `openspec/changes/production-teammate-runtime/proposal.md`
- Create: `openspec/changes/production-teammate-runtime/design.md`
- Create: `openspec/changes/production-teammate-runtime/tasks.md`
- Create: `openspec/changes/production-teammate-runtime/specs/multi-agent-orchestration/spec.md`
- Create: `openspec/changes/production-teammate-runtime/specs/agent-control-plane-tools/spec.md`
- Create: `openspec/changes/production-teammate-runtime/specs/tool-execution-hooks/spec.md`
- Create: `openspec/changes/production-teammate-runtime/specs/tool-capability-layer/spec.md`
- Create: `openspec/changes/production-teammate-runtime/specs/cli-agent-inspection/spec.md`

- [ ] **Step 1: Confirm active OpenSpec changes**

Run:

```bash
openspec list
```

Expected: no active change named `production-teammate-runtime`. If it exists, continue that change instead of creating a duplicate.

- [ ] **Step 2: Confirm change creation command**

Run:

```bash
openspec new change --help
```

Expected: output includes `Create a new change directory`.

- [ ] **Step 3: Create change directories**

Run:

```bash
openspec new change production-teammate-runtime
mkdir -p openspec/changes/production-teammate-runtime/specs/multi-agent-orchestration
mkdir -p openspec/changes/production-teammate-runtime/specs/agent-control-plane-tools
mkdir -p openspec/changes/production-teammate-runtime/specs/tool-execution-hooks
mkdir -p openspec/changes/production-teammate-runtime/specs/tool-capability-layer
mkdir -p openspec/changes/production-teammate-runtime/specs/cli-agent-inspection
```

Expected: `openspec/changes/production-teammate-runtime/` exists with five spec delta directories.

- [ ] **Step 4: Write `proposal.md`**

Create `openspec/changes/production-teammate-runtime/proposal.md` with:

```markdown
# Production Teammate Runtime

## Problem

`agent.multiAgent=true` currently builds a static tool-less orchestrator with specialist sub-agents. This makes dynamic teammate creation, least-privilege tool scope, background identity, capability escalation, and operator visibility inconsistent.

## Proposed Change

Reframe multi-agent mode as an in-process dynamic teammate runtime. Keep the existing model-facing `agent_spawn`, `agent_wait`, and `agent_stop` tools, extend `AgentRun` projection state, validate role maximum scope plus spawn-time `allowed_tools`, preserve structured hook block metadata, and route DynamicAllowedTools blocks into capability policy.

## User-Facing Impact

Users keep enabling multi-agent mode with `agent.multiAgent=true`. The main agent can answer directly or spawn teammates. CLI status exposes dynamic runtime availability; detailed cockpit/TUI projection is a follow-up once the shared runtime read model is stable. Remote A2A agents keep their existing v1 routing behavior.
```

- [ ] **Step 5: Write `design.md`**

Create `openspec/changes/production-teammate-runtime/design.md` with:

```markdown
# Design

## Source Of Truth

`AgentRun.ID`, background task ID, and RunLedger run ID remain the canonical identity chain. `AgentRun` stores the control-plane projection; background manager owns in-process execution; ChildSession owns context isolation; RunLedger mirrors durable state when enabled.

## V1 Tool Surface

The model-facing control surface remains `agent_spawn`, `agent_wait`, and `agent_stop`. No `teammate_*` or `agent_message` tool is exposed in v1.

## Teammate Types

The existing specialist roles become teammate types: `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. Each type has role maximum scope. Spawn-time `allowed_tools` can only narrow that scope.

## Capability Escalation

V1 emits capability requests from runtime interception of blocked DynamicAllowedTools attempts. `WithHooks` must preserve structured block metadata before converting the block into the existing user-facing error. Role maximum scope is a hard upper bound; approval cannot override it.

## Compatibility

`transfer_to_agent` remains a legacy ADK static sub-agent fallback in v1. New dynamic teammate work uses `agent_spawn`; legacy static sub-agent fallback, specialist re-routing, and existing remote A2A paths may still use `transfer_to_agent`.
```

- [ ] **Step 6: Write `tasks.md`**

Create `openspec/changes/production-teammate-runtime/tasks.md` with:

```markdown
# Tasks

- [ ] Rewrite OpenSpec contracts for dynamic teammate runtime v1
- [ ] Add AgentRun projection fields and patch support
- [ ] Add teammate type role maximum scope validation
- [ ] Extend agent_spawn with spawn_reason and role scope checks
- [ ] Preserve structured hook block metadata
- [ ] Add capability policy for blocked DynamicAllowedTools attempts
- [ ] Surface projected blocked state through agent_wait
- [ ] Add temporary prompt guidance for agent_spawn versus transfer_to_agent
- [ ] Update CLI and public docs for actual v1 behavior
- [ ] Run openspec verify/apply workflow checks
- [ ] Run go build ./...
- [ ] Run go test ./...
- [ ] Archive the OpenSpec change and update main specs through `openspec archive`
```

- [ ] **Step 7: Write `multi-agent-orchestration` delta spec**

Create `openspec/changes/production-teammate-runtime/specs/multi-agent-orchestration/spec.md` with:

```markdown
## REMOVED Requirements

### Requirement: Multi-agent orchestrator has no tools
The old guarantee that the root orchestrator has `Tools: nil` and no direct control tools is removed. The new coordinator-capable main agent may answer directly and may use `agent_spawn`, `agent_wait`, and `agent_stop` for dynamic teammate runs.

### Requirement: Orchestrator delegates all tool-requiring tasks
The old delegation-only rule is removed. Requests requiring work may be handled through dynamic teammate spawn, legacy static `transfer_to_agent` fallback, or direct answer when no tool execution is required.

### Requirement: Static BuildAgentTree as the only multi-agent shape
The static ADK sub-agent tree is no longer the only multi-agent execution model. Dynamic in-process teammate runs become the primary v1 path for new teammate work.

### Requirement: Static RoleToolSet partition as execution authority
Prefix-based `RoleToolSet` partitioning is no longer the final runtime execution authority. It is reframed as role maximum scope and default affinity. Spawn-time `allowed_tools` narrows that scope.

### Requirement: transfer_to_agent as only handoff primitive
`transfer_to_agent` is no longer the only handoff primitive. It remains as a legacy ADK static sub-agent fallback and existing remote A2A path while dynamic teammate work uses the agent control plane.

## MODIFIED Requirements

### Requirement: Hierarchical agent tree with sub-agents
When `agent.multiAgent` is true, the system SHALL support a coordinator-capable main agent that can answer directly or spawn in-process teammate runs through the agent control plane. Existing specialist identities (`operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, `ontologist`) SHALL remain available as teammate types. The static ADK sub-agent tree and `transfer_to_agent` path MAY remain available as a v1 compatibility fallback for legacy static sub-agent routing and existing remote A2A routing, but it SHALL NOT be the only handoff primitive.

#### Scenario: Dynamic teammate path is available
- **WHEN** `agent.multiAgent` is true
- **THEN** the main agent prompt SHALL describe dynamic teammate spawning through `agent_spawn`
- **AND** the prompt SHALL allow direct answers for simple conversational work
- **AND** the prompt SHALL describe `transfer_to_agent` as legacy static sub-agent fallback or re-routing only

### Requirement: Tool prefix partitioning
Tool prefix partitioning SHALL define role maximum scope and default affinity for teammate types. It SHALL NOT grant final execution authority by itself. The actual runtime tool set SHALL be the intersection of role maximum scope, spawn-time `allowed_tools`, active grants, and capability policy.

#### Scenario: Spawn-time tools narrow role scope
- **WHEN** teammate type `operator` has `fs_write` inside role maximum scope
- **AND** `agent_spawn` creates an operator teammate with `allowed_tools: ["fs_read"]`
- **THEN** `fs_write` SHALL be blocked until capability policy grants it

### Requirement: Re-routing protocol
The old static sub-agent re-routing protocol SHALL be reframed as recovery policy plus main-agent synthesis. The main agent SHALL avoid repeating a failed teammate path and MAY spawn a different teammate, summarize partial work, answer from gathered evidence, or escalate to the user.

#### Scenario: Failed teammate path is not immediately repeated
- **WHEN** a teammate run fails with a specialist identity recorded
- **THEN** recovery guidance SHALL identify the failed teammate type
- **AND** the next coordinator decision SHALL avoid immediately repeating the same teammate path for the same request

### Requirement: Event author identity
Teammate run and child-session authorship SHALL remain traceable in events, summaries, and operator projections. A teammate final result SHALL include enough identity metadata to distinguish the main agent from spawned teammates.

### Requirement: Remote agents as sub-agents
Remote A2A agents SHALL preserve their existing v1 behavior. They remain available through the existing remote A2A/static routing path. Treating remote A2A agents as dynamic teammate providers is outside v1.

#### Scenario: Remote A2A compatibility
- **WHEN** remote A2A agents are configured
- **THEN** existing remote loading and routing behavior remains available
- **AND** dynamic teammate-provider integration is not required for v1
```

- [ ] **Step 8: Write `agent-control-plane-tools` delta spec**

Create `openspec/changes/production-teammate-runtime/specs/agent-control-plane-tools/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: agent_spawn tool creates AgentRun with enriched prompt and advisory routing
The `agent_spawn` tool SHALL accept optional `spawn_reason` and `allowed_tools` parameters. When `agent` matches a built-in teammate type, `allowed_tools` SHALL be validated against that teammate type's role maximum scope before the run is created. The run projection SHALL carry spawn reason for operator visibility and audit correlation. When an in-process submitter is configured, `agent_spawn` SHALL submit the teammate prompt through the existing background manager using the pre-registered AgentRun ID.

#### Scenario: Spawn reason is stored for projection
- **WHEN** `agent_spawn` is called with `instruction: "Review patch"` and `spawn_reason: "parallel review"`
- **THEN** the created `AgentRun` SHALL expose `SpawnReason: "parallel review"`
- **AND** the response SHALL include `spawn_reason: "parallel review"`

#### Scenario: Allowed tools outside role scope are rejected
- **WHEN** `agent_spawn` is called with `agent: "planner"` and `allowed_tools: ["exec"]`
- **THEN** the tool SHALL return an error containing `tool "exec" outside role maximum scope for teammate type "planner"`
- **AND** no AgentRun SHALL be created

#### Scenario: Spawn submits through existing in-process execution path
- **WHEN** `agent_spawn` creates an operator teammate with `allowed_tools: ["fs_read"]`
- **THEN** the background submitter SHALL receive a context carrying agent name `operator`
- **AND** the same context SHALL carry DynamicAllowedTools `["fs_read"]`
- **AND** the background origin session SHALL preserve the parent session key rather than replacing it with the child AgentRun ID
- **AND** blocked DynamicAllowedTools attempts in that context SHALL update the matching AgentRun projection

### Requirement: agent_wait polls AgentRunStore until terminal status
`agent_wait` SHALL include projected condition fields in non-terminal timeout responses. During `blocked_waiting_approval`, `agent_wait` timeout SHALL return `timeout: true` and SHALL NOT cancel the run.

#### Scenario: Wait timeout during blocked approval is non-terminal
- **GIVEN** an `AgentRun` with status `running`, `RuntimeCondition: "blocked_waiting_approval"`, `BlockedReason: "capability request pending"`, and `GrantRequestID: "grant-1"`
- **WHEN** `agent_wait` times out
- **THEN** the response SHALL include `timeout: true`, `status: "running"`, `condition: "blocked_waiting_approval"`, `blocked_reason: "capability request pending"`, and `grant_request_id: "grant-1"`
- **AND** the run status SHALL remain `running`
```

- [ ] **Step 9: Write `tool-execution-hooks` delta spec**

Create `openspec/changes/production-teammate-runtime/specs/tool-execution-hooks/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: WithHooks middleware bridge
When a PreToolHook returns Action=Block, `WithHooks` SHALL preserve structured blocked-call metadata before returning the existing blocked-tool error. The metadata SHALL include tool name, agent name, session key, block reason, original params, and context.

#### Scenario: Blocked call metadata emitted before error
- **WHEN** `AgentAccessControlHook` blocks tool `exec` for agent `operator`
- **THEN** a blocked-call sink installed on the context SHALL receive tool name `exec`, agent name `operator`, and the hook block reason
- **AND** the tool handler SHALL still return the existing error format `tool 'exec' blocked by hook: <reason>`
```

- [ ] **Step 10: Write `tool-capability-layer` delta spec**

Create `openspec/changes/production-teammate-runtime/specs/tool-capability-layer/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: DynamicAllowedTools with runtime essentials
When `AgentAccessControlHook` blocks a teammate tool call with reason `tool restricted by DynamicAllowedTools`, the teammate runtime SHALL classify the blocked tool against role maximum scope. If the tool is inside role maximum scope, the runtime SHALL emit a capability request. If the tool is outside role maximum scope, the runtime SHALL emit a structured denial that user approval cannot override.

#### Scenario: Blocked tool inside role maximum scope requests capability
- **GIVEN** teammate type `operator` has `fs_write` inside role maximum scope
- **AND** the current run's DynamicAllowedTools contains only `fs_read`
- **WHEN** the teammate attempts `fs_write`
- **THEN** the runtime SHALL produce a capability decision requiring approval
- **AND** the run projection SHALL become `blocked_waiting_approval`

#### Scenario: Blocked tool outside role maximum scope is denied
- **GIVEN** teammate type `planner` does not have `exec` inside role maximum scope
- **WHEN** the teammate attempts `exec`
- **THEN** the runtime SHALL deny the request
- **AND** no approval request SHALL be surfaced

#### Scenario: Approved capability extends dynamic allowlist
- **GIVEN** an operator teammate has DynamicAllowedTools `["fs_read"]`
- **AND** the teammate requests dangerous tool `exec`
- **WHEN** the user or policy approves the capability request
- **THEN** the runtime SHALL append `exec` to the run's `AllowedTools` once
- **AND** the next teammate context built for that run SHALL include `exec` in DynamicAllowedTools
```

- [ ] **Step 11: Write `cli-agent-inspection` delta spec**

Create `openspec/changes/production-teammate-runtime/specs/cli-agent-inspection/spec.md` with:

```markdown
## MODIFIED Requirements

### Requirement: Agent status inspection
The `lango agent status` command SHALL identify dynamic teammate runtime availability when multi-agent mode is enabled. JSON output SHALL include `teammate_runtime: "dynamic-v1"` when `agent.multiAgent` is true. Table output SHALL include a `Teammate Runtime` line.

#### Scenario: Multi-agent status shows teammate runtime
- **GIVEN** `agent.multiAgent` is true
- **WHEN** the user runs `lango agent status --json`
- **THEN** the output SHALL include `"teammate_runtime": "dynamic-v1"`
```

- [ ] **Step 12: Verify OpenSpec status**

Run:

```bash
openspec status --change production-teammate-runtime
```

Expected: proposal, design, tasks, and specs are present. The change is ready for implementation.

- [ ] **Step 13: Record suggested commit OpenSpec artifacts**

Suggested user-run commands after review:

```bash
git add openspec/changes/production-teammate-runtime
git commit -m "spec: define production teammate runtime"
```

Expected: user can run the suggested commit after review with only OpenSpec files staged.

## Task 2: AgentRun Projection Fields

**Files:**
- Modify: `internal/agentrt/agent_run.go`
- Modify: `internal/agentrt/agent_run_store.go`
- Test: `internal/agentrt/agent_run_store_test.go`

- [ ] **Step 1: Write failing projection-field tests**

Create `internal/agentrt/agent_run_store_test.go` with:

```go
package agentrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentRunStore_CopyIncludesProjectionFields(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "arun-projection",
		Status:           AgentRunRunning,
		RequestedAgent:   "operator",
		SpawnReason:      "parallel file inspection",
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "capability request pending",
		GrantRequestID:   "grant-123",
		WaitingOnRunID:   "arun-child",
		RecoveryState:    "retry_with_hint",
		AllowedTools:     []string{"fs_read"},
	}))

	got, err := store.Get("arun-projection")
	require.NoError(t, err)

	assert.Equal(t, "parallel file inspection", got.SpawnReason)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, got.RuntimeCondition)
	assert.Equal(t, "capability request pending", got.BlockedReason)
	assert.Equal(t, "grant-123", got.GrantRequestID)
	assert.Equal(t, "arun-child", got.WaitingOnRunID)
	assert.Equal(t, "retry_with_hint", got.RecoveryState)

	got.AllowedTools[0] = "exec"
	again, err := store.Get("arun-projection")
	require.NoError(t, err)
	assert.Equal(t, []string{"fs_read"}, again.AllowedTools)
}

func TestAgentRunStore_UpdateProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "arun-blocked",
		Status: AgentRunRunning,
	}))

	err := store.UpdateProjection("arun-blocked", RunProjectionPatch{
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "capability request pending",
		GrantRequestID:   "grant-456",
	})
	require.NoError(t, err)

	got, err := store.Get("arun-blocked")
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, got.Status)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, got.RuntimeCondition)
	assert.Equal(t, "capability request pending", got.BlockedReason)
	assert.Equal(t, "grant-456", got.GrantRequestID)
}

func TestAgentRunStore_UpdateProjectionAddsAllowedToolOnce(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:           "arun-grant",
		Status:       AgentRunRunning,
		AllowedTools: []string{"fs_read"},
	}))

	require.NoError(t, store.UpdateProjection("arun-grant", RunProjectionPatch{AddAllowedTool: "exec"}))
	require.NoError(t, store.UpdateProjection("arun-grant", RunProjectionPatch{AddAllowedTool: "exec"}))

	got, err := store.Get("arun-grant")
	require.NoError(t, err)
	assert.Equal(t, []string{"fs_read", "exec"}, got.AllowedTools)
}

func TestAgentRunStore_UpdateProjectionRejectsTerminalRun(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:     "arun-done",
		Status: AgentRunCompleted,
	}))

	err := store.UpdateProjection("arun-done", RunProjectionPatch{
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already completed")
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/agentrt -run 'TestAgentRunStore_(CopyIncludesProjectionFields|UpdateProjection)' -count=1
```

Expected: FAIL with undefined `AgentRunConditionBlockedWaitingApproval`, undefined fields such as `SpawnReason`, and missing `UpdateProjection`.

- [ ] **Step 3: Add projection fields to `AgentRun`**

Modify `internal/agentrt/agent_run.go`:

```go
// AgentRunCondition represents projected runtime conditions layered on top of
// the base AgentRunStatus. Conditions are not terminal states.
type AgentRunCondition string

const (
	AgentRunConditionNone                   AgentRunCondition = ""
	AgentRunConditionBlockedWaitingApproval AgentRunCondition = "blocked_waiting_approval"
	AgentRunConditionBlockedWaitingMessage  AgentRunCondition = "blocked_waiting_message"
	AgentRunConditionWaitingOnTeammate      AgentRunCondition = "waiting_on_teammate"
	AgentRunConditionResuming               AgentRunCondition = "resuming"
	AgentRunConditionOrphaned               AgentRunCondition = "orphaned"
	AgentRunConditionRecovering             AgentRunCondition = "recovering"
)
```

Add these fields to `AgentRun` after `AllowedTools`:

```go
	SpawnReason      string
	RuntimeCondition AgentRunCondition
	BlockedReason    string
	GrantRequestID   string
	WaitingOnRunID   string
	RecoveryState    string
```

- [ ] **Step 4: Extend store interface and in-memory store**

Modify `internal/agentrt/agent_run_store.go`:

```go
type RunProjectionPatch struct {
	RuntimeCondition AgentRunCondition
	BlockedReason    string
	GrantRequestID   string
	WaitingOnRunID   string
	RecoveryState    string
	AddAllowedTool   string
}
```

Add `UpdateProjection` to `AgentRunStore`:

```go
	UpdateProjection(id string, patch RunProjectionPatch) error
```

Add method on `InMemoryAgentRunStore`:

```go
func (s *InMemoryAgentRunStore) UpdateProjection(id string, patch RunProjectionPatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	if !ok {
		return fmt.Errorf("update agent run projection: ID %q not found", id)
	}
	if run.Status.isTerminal() {
		return fmt.Errorf("update agent run projection: ID %q is already %s", id, run.Status)
	}

	run.RuntimeCondition = patch.RuntimeCondition
	run.BlockedReason = patch.BlockedReason
	run.GrantRequestID = patch.GrantRequestID
	run.WaitingOnRunID = patch.WaitingOnRunID
	run.RecoveryState = patch.RecoveryState
	if patch.AddAllowedTool != "" && !stringSliceContains(run.AllowedTools, patch.AddAllowedTool) {
		run.AllowedTools = append(run.AllowedTools, patch.AddAllowedTool)
	}
	return nil
}

func stringSliceContains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
```

- [ ] **Step 5: Run projection tests**

Run:

```bash
go test ./internal/agentrt -run 'TestAgentRunStore_(CopyIncludesProjectionFields|UpdateProjection)' -count=1
```

Expected: PASS.

- [ ] **Step 6: Record suggested commit projection fields**

Suggested user-run commands after review:

```bash
git add internal/agentrt/agent_run.go internal/agentrt/agent_run_store.go internal/agentrt/agent_run_store_test.go
git commit -m "feat: add teammate run projection fields"
```

Expected: user can run the suggested commit after review with only agent runtime projection files staged.

## Task 3: Teammate Type Role Scope

**Files:**
- Create: `internal/agentrt/teammate_types.go`
- Create: `internal/agentrt/teammate_types_test.go`
- Modify: `internal/agentrt/control_tools.go`
- Modify: `internal/agentrt/control_tools_test.go`

- [ ] **Step 1: Lock actual registered tool inventory**

Run:

```bash
rg -n 'Name:\s*"|NewTool\("' internal/tools internal/agentrt internal/background internal/cron internal/workflow internal/memory internal/agentmemory internal/ontology internal/app/tools_meta.go internal/toolcatalog/dispatcher.go internal/tooloutput/tools.go -g '*.go'
```

Expected output includes the concrete tool names used in the role scope below, including `exec`, `fs_read`, `fs_write`, `fs_list`, `fs_delete`, `browser_search`, `web_search`, `workflow_run`, `memory_agent_save`, `ontology_assert_fact`, `ontology_list_conflicts`, `tool_output_get`, `builtin_list`, `builtin_search`, and `builtin_health`. If any name in the code block below is not present in this inventory, replace it with the registered name before writing tests.

- [ ] **Step 2: Write failing teammate type tests**

Create `internal/agentrt/teammate_types_test.go` with:

```go
package agentrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinTeammateTypes(t *testing.T) {
	types := BuiltinTeammateTypes()
	require.Contains(t, types, "operator")
	require.Contains(t, types, "navigator")
	require.Contains(t, types, "vault")
	require.Contains(t, types, "librarian")
	require.Contains(t, types, "automator")
	require.Contains(t, types, "planner")
	require.Contains(t, types, "chronicler")
	require.Contains(t, types, "ontologist")

	assert.True(t, types["operator"].AllowsTool("fs_read"))
	assert.True(t, types["operator"].AllowsTool("exec"))
	assert.False(t, types["planner"].AllowsTool("exec"))
	assert.True(t, types["planner"].AllowsTool("agent_wait"))
}

func TestValidateAllowedToolsForTeammate(t *testing.T) {
	err := ValidateAllowedToolsForTeammate("operator", []string{"fs_read", "exec"})
	require.NoError(t, err)

	err = ValidateAllowedToolsForTeammate("planner", []string{"exec"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `tool "exec" outside role maximum scope for teammate type "planner"`)
}

func TestValidateAllowedToolsForTeammatePreservesCompatibilityForUnknownAgent(t *testing.T) {
	err := ValidateAllowedToolsForTeammate("custom-agent", []string{"custom_tool"})
	require.NoError(t, err)
}
```

- [ ] **Step 3: Run tests and verify failure**

Run:

```bash
go test ./internal/agentrt -run 'Test(BuiltinTeammateTypes|ValidateAllowedToolsForTeammate)' -count=1
```

Expected: FAIL with undefined `BuiltinTeammateTypes` and `ValidateAllowedToolsForTeammate`.

- [ ] **Step 4: Add teammate type metadata**

Create `internal/agentrt/teammate_types.go` with:

```go
package agentrt

import "fmt"

type TeammateType struct {
	Name        string
	Description string
	MaxTools    map[string]bool
}

func (t TeammateType) AllowsTool(toolName string) bool {
	if len(t.MaxTools) == 0 {
		return false
	}
	return t.MaxTools[toolName]
}

func BuiltinTeammateTypes() map[string]TeammateType {
	return map[string]TeammateType{
		"operator": {
			Name:        "operator",
			Description: "Local execution, file operations, and skill execution.",
			MaxTools: allowTools(
				"exec", "exec_bg", "exec_status", "exec_stop",
				"fs_read", "fs_write", "fs_list", "fs_delete", "fs_edit", "fs_mkdir", "fs_stat",
				"tool_output_get", "builtin_list", "builtin_search", "builtin_health",
				"agent_spawn", "agent_wait", "agent_stop", "task_create", "task_get", "task_list", "task_update",
			),
		},
		"navigator": {
			Name:        "navigator",
			Description: "Browser and web navigation work.",
			MaxTools: allowTools(
				"browser_navigate", "browser_search", "browser_observe", "browser_action", "browser_screenshot", "browser_extract",
				"web_search", "tool_output_get", "builtin_list", "builtin_search", "builtin_health",
			),
		},
		"vault": {
			Name:        "vault",
			Description: "Cryptography, secrets, payment, wallet, and signing work.",
			MaxTools: allowTools(
				"crypto_encrypt", "crypto_decrypt", "crypto_sign", "crypto_hash", "crypto_keys",
				"secrets_store", "secrets_get", "secrets_list", "secrets_delete",
				"payment_send", "payment_balance", "payment_history", "payment_limits", "payment_wallet_info", "payment_create_wallet", "payment_x402_fetch",
				"tool_output_get",
				"builtin_list", "builtin_search", "builtin_health",
			),
		},
		"librarian": {
			Name:        "librarian",
			Description: "Knowledge search, RAG, graph traversal, learning, and skill management.",
			MaxTools: allowTools(
				"search_knowledge", "get_knowledge_history", "search_learnings", "graph_query", "graph_traverse",
				"save_knowledge", "save_learning", "create_skill", "list_skills", "view_skill", "import_skill",
				"tool_output_get", "builtin_list", "builtin_search", "builtin_health",
			),
		},
		"automator": {
			Name:        "automator",
			Description: "Background, cron, workflow, and scheduled work.",
			MaxTools: allowTools(
				"cron_add", "cron_list", "cron_pause", "cron_resume", "cron_remove", "cron_history",
				"bg_submit", "bg_status", "bg_list", "bg_result", "bg_cancel",
				"workflow_run", "workflow_status", "workflow_list", "workflow_cancel", "workflow_save",
				"agent_spawn", "agent_wait", "agent_stop",
				"task_create", "task_get", "task_list", "task_update", "tool_output_get",
				"builtin_list", "builtin_search", "builtin_health",
			),
		},
		"planner": {
			Name:        "planner",
			Description: "Planning, decomposition, and strategy.",
			MaxTools: allowTools(
				"agent_spawn", "agent_wait", "agent_stop", "task_create", "task_get",
				"task_list", "task_update", "tool_output_get", "builtin_list", "builtin_search", "builtin_health",
			),
		},
		"chronicler": {
			Name:        "chronicler",
			Description: "Memory, observations, reflections, and recall.",
			MaxTools: allowTools(
				"memory_list_observations", "memory_list_reflections",
				"memory_agent_save", "memory_agent_recall", "memory_agent_forget",
				"tool_output_get", "builtin_list", "builtin_search", "builtin_health",
			),
		},
		"ontologist": {
			Name:        "ontologist",
			Description: "Ontology types, entities, facts, conflicts, and ingestion.",
			MaxTools: allowTools(
				"ontology_list_actions", "ontology_promote_type", "ontology_promote_predicate",
				"ontology_list_types", "ontology_describe_type", "ontology_query_entities", "ontology_get_entity",
				"ontology_assert_fact", "ontology_retract_fact", "ontology_list_conflicts", "ontology_resolve_conflict",
				"ontology_merge_entities", "ontology_facts_at", "ontology_import_json", "ontology_import_csv",
				"ontology_from_mcp", "ontology_schema_health", "ontology_type_usage",
				"tool_output_get", "builtin_list", "builtin_search", "builtin_health",
			),
		},
	}
}

func ValidateAllowedToolsForTeammate(teammateType string, allowedTools []string) error {
	if teammateType == "" || len(allowedTools) == 0 {
		return nil
	}
	types := BuiltinTeammateTypes()
	tt, ok := types[teammateType]
	if !ok {
		return nil
	}
	for _, toolName := range allowedTools {
		if !tt.AllowsTool(toolName) {
			return fmt.Errorf("tool %q outside role maximum scope for teammate type %q", toolName, teammateType)
		}
	}
	return nil
}

func allowTools(names ...string) map[string]bool {
	result := make(map[string]bool, len(names))
	for _, name := range names {
		result[name] = true
	}
	return result
}
```

- [ ] **Step 5: Run teammate type tests**

Run:

```bash
go test ./internal/agentrt -run 'Test(BuiltinTeammateTypes|ValidateAllowedToolsForTeammate)' -count=1
```

Expected: PASS.

- [ ] **Step 6: Write failing spawn validation tests**

Append to `internal/agentrt/control_tools_test.go`:

```go
func TestAgentSpawn_WithSpawnReason(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	result, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction":  "inspect files",
		"agent":        "operator",
		"spawn_reason": "parallel file inspection",
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, "parallel file inspection", m["spawn_reason"])

	run, err := store.Get(m["agent_id"].(string))
	require.NoError(t, err)
	assert.Equal(t, "parallel file inspection", run.SpawnReason)
}

func TestAgentSpawn_RejectsAllowedToolsOutsideRoleScope(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	cp := &AgentControlPlane{
		RunStore:   store,
		Projection: NewAgentRunProjection(store),
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	_, err := spawnTool.call(context.Background(), map[string]interface{}{
		"instruction":   "plan and execute",
		"agent":         "planner",
		"allowed_tools": []interface{}{"exec"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `tool "exec" outside role maximum scope for teammate type "planner"`)
	assert.Empty(t, store.List())
}
```

- [ ] **Step 7: Run spawn tests and verify failure**

Run:

```bash
go test ./internal/agentrt -run 'TestAgentSpawn_(WithSpawnReason|RejectsAllowedToolsOutsideRoleScope)' -count=1
```

Expected: FAIL because `spawn_reason` is not parsed, returned, or validated.

- [ ] **Step 8: Extend `agent_spawn`**

Modify `internal/agentrt/control_tools.go`:

```go
			Str("spawn_reason", "Short reason why this teammate is being spawned").
```

Add parsing after `allowedTools := toolparam.StringSlice(params, "allowed_tools")`:

```go
			spawnReason := toolparam.OptionalString(params, "spawn_reason", "")
			if err := ValidateAllowedToolsForTeammate(requestedAgent, allowedTools); err != nil {
				return nil, fmt.Errorf("agent spawn: %w", err)
			}
```

Add field to `AgentRun` literal:

```go
				SpawnReason:    spawnReason,
```

Add response field:

```go
				"spawn_reason":    spawnReason,
```

- [ ] **Step 9: Run agentrt tests**

Run:

```bash
go test ./internal/agentrt -count=1
```

Expected: PASS.

- [ ] **Step 10: Record suggested commit role scope and spawn reason**

Suggested user-run commands after review:

```bash
git add internal/agentrt/teammate_types.go internal/agentrt/teammate_types_test.go internal/agentrt/control_tools.go internal/agentrt/control_tools_test.go
git commit -m "feat: validate teammate role tool scope"
```

Expected: user can run the suggested commit after review with teammate type and control tool changes.

## Task 4: Structured Hook Block Metadata

**Files:**
- Modify: `internal/toolchain/hooks.go`
- Modify: `internal/toolchain/mw_hooks.go`
- Modify: `internal/toolchain/mw_hooks_test.go`

- [ ] **Step 1: Write failing metadata sink test**

Append to `internal/toolchain/mw_hooks_test.go`:

```go
func TestWithHooks_EmitsBlockedToolCallMetadata(t *testing.T) {
	reg := NewHookRegistry()
	reg.RegisterPre(&stubPreHook{
		name:     "blocker",
		priority: 1,
		result:   PreHookResult{Action: Block, BlockReason: "tool restricted by DynamicAllowedTools"},
	})

	tool := &agent.Tool{
		Name: "exec",
		Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			t.Fatal("handler should not execute")
			return nil, nil
		},
	}

	var captured BlockedToolCall
	ctx := WithAgentName(context.Background(), "operator")
	ctx = WithBlockedToolCallSink(ctx, func(call BlockedToolCall) {
		captured = call
	})

	handler := WithHooks(reg)(tool, tool.Handler)
	_, err := handler(ctx, map[string]interface{}{"command": "rm -rf /tmp/nope"})
	require.Error(t, err)

	assert.Equal(t, "exec", captured.ToolName)
	assert.Equal(t, "operator", captured.AgentName)
	assert.Equal(t, "tool restricted by DynamicAllowedTools", captured.BlockReason)
	assert.Equal(t, "rm -rf /tmp/nope", captured.Params["command"])
	assert.NotNil(t, captured.Ctx)
}
```

- [ ] **Step 2: Run hook test and verify failure**

Run:

```bash
go test ./internal/toolchain -run TestWithHooks_EmitsBlockedToolCallMetadata -count=1
```

Expected: FAIL with undefined `BlockedToolCall` and `WithBlockedToolCallSink`.

- [ ] **Step 3: Add blocked metadata context API**

Modify `internal/toolchain/hooks.go`:

```go
type BlockedToolCall struct {
	ToolName    string
	AgentName   string
	SessionKey  string
	BlockReason string
	Params      map[string]interface{}
	Ctx         context.Context
}

type blockedToolCallSinkCtxKey struct{}

type BlockedToolCallSink func(BlockedToolCall)

func WithBlockedToolCallSink(ctx context.Context, sink BlockedToolCallSink) context.Context {
	return context.WithValue(ctx, blockedToolCallSinkCtxKey{}, sink)
}

func emitBlockedToolCall(ctx context.Context, call BlockedToolCall) {
	sink, ok := ctx.Value(blockedToolCallSinkCtxKey{}).(BlockedToolCallSink)
	if !ok || sink == nil {
		return
	}
	sink(call)
}
```

- [ ] **Step 4: Emit metadata before blocked error**

Modify the `case Block:` branch in `internal/toolchain/mw_hooks.go`:

```go
			case Block:
				emitBlockedToolCall(ctx, BlockedToolCall{
					ToolName:    tool.Name,
					AgentName:   hctx.AgentName,
					SessionKey:  hctx.SessionKey,
					BlockReason: preResult.BlockReason,
					Params:      params,
					Ctx:         ctx,
				})
				return nil, fmt.Errorf("tool '%s' blocked by hook: %s", tool.Name, preResult.BlockReason)
```

- [ ] **Step 5: Run toolchain tests**

Run:

```bash
go test ./internal/toolchain -count=1
```

Expected: PASS.

- [ ] **Step 6: Record suggested commit hook metadata**

Suggested user-run commands after review:

```bash
git add internal/toolchain/hooks.go internal/toolchain/mw_hooks.go internal/toolchain/mw_hooks_test.go
git commit -m "feat: preserve blocked tool metadata"
```

Expected: user can run the suggested commit after review with only toolchain files staged.

## Task 5: Capability Policy

**Files:**
- Create: `internal/agentrt/capability_policy.go`
- Create: `internal/agentrt/capability_policy_test.go`

- [ ] **Step 1: Write failing capability policy tests**

Create `internal/agentrt/capability_policy_test.go` with:

```go
package agentrt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
)

func TestCapabilityPolicy_DeniesOutsideRoleScope(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:          "arun-1",
		TeammateType:  "planner",
		ToolName:      "exec",
		CurrentAllowed: []string{"agent_wait"},
		ToolSafety:    agent.SafetyLevelDangerous,
	})

	assert.Equal(t, CapabilityDecisionDeny, decision.Kind)
	assert.Contains(t, decision.Reason, "outside role maximum scope")
	assert.Empty(t, decision.GrantRequestID)
}

func TestCapabilityPolicy_AllowsAlreadyGrantedTool(t *testing.T) {
	policy := CapabilityPolicy{
		ActiveGrants: map[string]map[string]bool{
			"arun-2": {"fs_write": true},
		},
	}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:          "arun-2",
		TeammateType:  "operator",
		ToolName:      "fs_write",
		CurrentAllowed: []string{"fs_read"},
		ToolSafety:    agent.SafetyLevelModerate,
	})

	assert.Equal(t, CapabilityDecisionAllow, decision.Kind)
	assert.Equal(t, "existing grant", decision.Reason)
}

func TestCapabilityPolicy_RequiresApprovalForDangerousToolInsideRoleScope(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:          "arun-3",
		TeammateType:  "operator",
		ToolName:      "exec",
		CurrentAllowed: []string{"fs_read"},
		ToolSafety:    agent.SafetyLevelDangerous,
	})

	require.Equal(t, CapabilityDecisionNeedsApproval, decision.Kind)
	assert.Equal(t, "dangerous tool requires approval", decision.Reason)
	assert.Equal(t, "grant-arun-3-exec", decision.GrantRequestID)
}

func TestCapabilityPolicy_AllowsSafeToolInsideRoleScope(t *testing.T) {
	policy := CapabilityPolicy{}

	decision := policy.Evaluate(CapabilityRequest{
		RunID:          "arun-4",
		TeammateType:  "operator",
		ToolName:      "fs_read",
		CurrentAllowed: []string{},
		ToolSafety:    agent.SafetyLevelSafe,
	})

	assert.Equal(t, CapabilityDecisionAllow, decision.Kind)
	assert.Equal(t, "safe tool inside role maximum scope", decision.Reason)
}

func TestCapabilityPolicy_GrantInitializesActiveGrantMap(t *testing.T) {
	policy := CapabilityPolicy{}
	policy.Grant("arun-5", "exec")

	assert.True(t, policy.ActiveGrants["arun-5"]["exec"])
}
```

- [ ] **Step 2: Run tests and verify failure**

Run:

```bash
go test ./internal/agentrt -run TestCapabilityPolicy -count=1
```

Expected: FAIL with undefined `CapabilityPolicy`, `CapabilityRequest`, and decision constants.

- [ ] **Step 3: Implement pure capability policy**

Create `internal/agentrt/capability_policy.go` with:

```go
package agentrt

import (
	"fmt"

	"github.com/langoai/lango/internal/agent"
)

type CapabilityDecisionKind string

const (
	CapabilityDecisionAllow         CapabilityDecisionKind = "allow"
	CapabilityDecisionNeedsApproval CapabilityDecisionKind = "needs_approval"
	CapabilityDecisionDeny          CapabilityDecisionKind = "deny"
)

type CapabilityRequest struct {
	RunID          string
	TeammateType  string
	ToolName      string
	CurrentAllowed []string
	ToolSafety    agent.SafetyLevel
}

type CapabilityDecision struct {
	Kind           CapabilityDecisionKind
	Reason         string
	GrantRequestID string
}

type CapabilityPolicy struct {
	ActiveGrants map[string]map[string]bool
}

func (p CapabilityPolicy) Evaluate(req CapabilityRequest) CapabilityDecision {
	if !teammateAllowsTool(req.TeammateType, req.ToolName) {
		return CapabilityDecision{
			Kind:   CapabilityDecisionDeny,
			Reason: fmt.Sprintf("tool %q outside role maximum scope for teammate type %q", req.ToolName, req.TeammateType),
		}
	}
	if p.hasGrant(req.RunID, req.ToolName) {
		return CapabilityDecision{Kind: CapabilityDecisionAllow, Reason: "existing grant"}
	}
	if req.ToolSafety.IsDangerous() {
		return CapabilityDecision{
			Kind:           CapabilityDecisionNeedsApproval,
			Reason:         "dangerous tool requires approval",
			GrantRequestID: fmt.Sprintf("grant-%s-%s", req.RunID, req.ToolName),
		}
	}
	return CapabilityDecision{Kind: CapabilityDecisionAllow, Reason: "safe tool inside role maximum scope"}
}

func (p CapabilityPolicy) hasGrant(runID, toolName string) bool {
	if p.ActiveGrants == nil {
		return false
	}
	grants := p.ActiveGrants[runID]
	return grants[toolName]
}

func (p *CapabilityPolicy) Grant(runID, toolName string) {
	if p.ActiveGrants == nil {
		p.ActiveGrants = make(map[string]map[string]bool)
	}
	if p.ActiveGrants[runID] == nil {
		p.ActiveGrants[runID] = make(map[string]bool)
	}
	p.ActiveGrants[runID][toolName] = true
}

func teammateAllowsTool(teammateType, toolName string) bool {
	types := BuiltinTeammateTypes()
	tt, ok := types[teammateType]
	if !ok {
		return false
	}
	return tt.AllowsTool(toolName)
}
```

- [ ] **Step 4: Run capability policy tests**

Run:

```bash
go test ./internal/agentrt -run TestCapabilityPolicy -count=1
```

Expected: PASS.

- [ ] **Step 5: Record suggested commit capability policy**

Suggested user-run commands after review:

```bash
git add internal/agentrt/capability_policy.go internal/agentrt/capability_policy_test.go
git commit -m "feat: add teammate capability policy"
```

Expected: user can run the suggested commit after review with capability policy files.

## Task 6: Capability Runtime Wiring

**Files:**
- Create: `internal/agentrt/capability_runtime.go`
- Create: `internal/agentrt/capability_runtime_test.go`
- Modify: `internal/agentrt/control_tools.go`
- Modify: `internal/agentrt/control_tools_test.go`
- Modify: `internal/app/modules.go`

- [ ] **Step 1: Write failing wiring tests**

Create `internal/agentrt/capability_runtime_test.go` with:

```go
package agentrt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolchain"
)

func TestCapabilityRuntime_BlockedDynamicAllowedToolsUpdatesProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-cap",
		Status:         AgentRunRunning,
		RequestedAgent: "operator",
		AllowedTools:   []string{"fs_read"},
	}))

	runtime := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(toolName string) agent.SafetyLevel {
		assert.Equal(t, "exec", toolName)
		return agent.SafetyLevelDangerous
	})

	sink := runtime.BlockedToolSinkForRun("arun-cap")
	sink(toolchain.BlockedToolCall{
		ToolName:    "exec",
		AgentName:   "operator",
		BlockReason: "tool restricted by DynamicAllowedTools",
		Params:      map[string]interface{}{"command": "ls"},
		Ctx:         context.Background(),
	})

	run, err := store.Get("arun-cap")
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, run.Status)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, run.RuntimeCondition)
	assert.Equal(t, "dangerous tool requires approval", run.BlockedReason)
	assert.Equal(t, "grant-arun-cap-exec", run.GrantRequestID)
}

func TestCapabilityRuntime_OutsideScopeDenialDoesNotRequestApproval(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-deny",
		Status:         AgentRunRunning,
		RequestedAgent: "planner",
		AllowedTools:   []string{"agent_wait"},
	}))

	runtime := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(string) agent.SafetyLevel {
		return agent.SafetyLevelDangerous
	})

	runtime.BlockedToolSinkForRun("arun-deny")(toolchain.BlockedToolCall{
		ToolName:    "exec",
		AgentName:   "planner",
		BlockReason: "tool restricted by DynamicAllowedTools",
		Ctx:         context.Background(),
	})

	run, err := store.Get("arun-deny")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
	assert.Contains(t, run.BlockedReason, "outside role maximum scope")
	assert.Empty(t, run.GrantRequestID)
}

func TestCapabilityRuntime_ApplyGrantClearsBlockedProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "arun-grant",
		Status:           AgentRunRunning,
		RequestedAgent:   "operator",
		AllowedTools:     []string{"fs_read"},
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "dangerous tool requires approval",
		GrantRequestID:   "grant-arun-grant-exec",
	}))

	runtime := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(string) agent.SafetyLevel {
		return agent.SafetyLevelDangerous
	})

	require.NoError(t, runtime.ApplyGrant("arun-grant", "exec"))
	assert.True(t, runtime.Policy.ActiveGrants["arun-grant"]["exec"])

	run, err := store.Get("arun-grant")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
	assert.Empty(t, run.BlockedReason)
	assert.Empty(t, run.GrantRequestID)
	assert.Equal(t, []string{"fs_read", "exec"}, run.AllowedTools)

	ctx := runtime.ContextForRun(context.Background(), run)
	assert.Equal(t, []string{"fs_read", "exec"}, ctxkeys.DynamicAllowedToolsFromContext(ctx))
}

func TestCapabilityRuntime_ContextForRunWiresBlockedHookToProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-context",
		Status:         AgentRunRunning,
		RequestedAgent: "operator",
		AllowedTools:   []string{"fs_read"},
	}))
	run, err := store.Get("arun-context")
	require.NoError(t, err)

	runtime := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(toolName string) agent.SafetyLevel {
		assert.Equal(t, "exec", toolName)
		return agent.SafetyLevelDangerous
	})

	reg := toolchain.NewHookRegistry()
	reg.RegisterPre(toolchain.NewAgentAccessControlHook(nil))
	tool := &agent.Tool{
		Name: "exec",
		Handler: func(context.Context, map[string]interface{}) (interface{}, error) {
			t.Fatal("handler should not execute")
			return nil, nil
		},
	}

	ctx := runtime.ContextForRun(context.Background(), run)
	handler := toolchain.WithHooks(reg)(tool, tool.Handler)
	_, err = handler(ctx, map[string]interface{}{"command": "ls"})
	require.Error(t, err)

	got, err := store.Get("arun-context")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionBlockedWaitingApproval, got.RuntimeCondition)
	assert.Equal(t, "grant-arun-context-exec", got.GrantRequestID)
}
```

- [ ] **Step 2: Run wiring tests and verify failure**

Run:

```bash
go test ./internal/agentrt -run TestCapabilityRuntime -count=1
```

Expected: FAIL with undefined `NewCapabilityRuntime`.

- [ ] **Step 3: Implement capability runtime wiring**

Create `internal/agentrt/capability_runtime.go` with:

```go
package agentrt

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/toolchain"
)

const dynamicAllowedToolsBlockReason = "tool restricted by DynamicAllowedTools"

type ToolSafetyLookup func(toolName string) agent.SafetyLevel

type CapabilityRuntime struct {
	Store      AgentRunStore
	Policy     *CapabilityPolicy
	ToolSafety ToolSafetyLookup
}

func NewCapabilityRuntime(store AgentRunStore, policy *CapabilityPolicy, lookup ToolSafetyLookup) *CapabilityRuntime {
	if policy == nil {
		policy = &CapabilityPolicy{}
	}
	if policy.ActiveGrants == nil {
		policy.ActiveGrants = make(map[string]map[string]bool)
	}
	return &CapabilityRuntime{
		Store:      store,
		Policy:     policy,
		ToolSafety: lookup,
	}
}

func (r *CapabilityRuntime) BlockedToolSinkForRun(runID string) toolchain.BlockedToolCallSink {
	return func(call toolchain.BlockedToolCall) {
		_ = r.HandleBlockedToolCall(runID, call)
	}
}

func (r *CapabilityRuntime) ContextForRun(ctx context.Context, run *AgentRun) context.Context {
	ctx = ctxkeys.WithAgentName(ctx, run.RequestedAgent)
	ctx = ctxkeys.WithDynamicAllowedTools(ctx, run.AllowedTools)
	return toolchain.WithBlockedToolCallSink(ctx, r.BlockedToolSinkForRun(run.ID))
}

func (r *CapabilityRuntime) HandleBlockedToolCall(runID string, call toolchain.BlockedToolCall) error {
	if call.BlockReason != dynamicAllowedToolsBlockReason {
		return nil
	}
	run, err := r.Store.Get(runID)
	if err != nil {
		return err
	}
	safety := agent.SafetyLevelDangerous
	if r.ToolSafety != nil {
		safety = r.ToolSafety(call.ToolName)
	}
	decision := r.Policy.Evaluate(CapabilityRequest{
		RunID:           run.ID,
		TeammateType:   run.RequestedAgent,
		ToolName:       call.ToolName,
		CurrentAllowed: run.AllowedTools,
		ToolSafety:     safety,
	})
	switch decision.Kind {
	case CapabilityDecisionNeedsApproval:
		return r.Store.UpdateProjection(run.ID, RunProjectionPatch{
			RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
			BlockedReason:    decision.Reason,
			GrantRequestID:   decision.GrantRequestID,
		})
	case CapabilityDecisionDeny:
		return r.Store.UpdateProjection(run.ID, RunProjectionPatch{
			RuntimeCondition: AgentRunConditionNone,
			BlockedReason:    decision.Reason,
		})
	case CapabilityDecisionAllow:
		return nil
	default:
		return fmt.Errorf("capability decision: unknown kind %q", decision.Kind)
	}
}

func (r *CapabilityRuntime) ApplyGrant(runID, toolName string) error {
	r.Policy.Grant(runID, toolName)
	return r.Store.UpdateProjection(runID, RunProjectionPatch{AddAllowedTool: toolName})
}
```

- [ ] **Step 4: Wire `agent_spawn` to the submitter context**

Modify `internal/agentrt/control_tools.go`:

```go
type AgentRunSubmitter interface {
	Submit(context.Context, string, background.Origin) (string, error)
}

type AgentControlPlane struct {
	RunStore          AgentRunStore
	Projection        *AgentRunProjection
	Submitter         AgentRunSubmitter
	CapabilityRuntime *CapabilityRuntime
}
```

After `cp.Projection.RegisterPending(agentID)`, submit the child run when a submitter is configured:

```go
			if cp.Submitter != nil {
				childCtx := ctx
				if cp.CapabilityRuntime != nil {
					childCtx = cp.CapabilityRuntime.ContextForRun(ctx, run)
				}
				parentSession := session.SessionKeyFromContext(ctx)
				if _, err := cp.Submitter.Submit(childCtx, enrichedInstruction, background.Origin{
					Channel: "agent_control",
					Session: parentSession,
				}); err != nil {
					_ = cp.RunStore.UpdateStatus(agentID, AgentRunFailed, "", err.Error())
					return nil, fmt.Errorf("agent spawn submit: %w", err)
				}
			}
```

Append this submitter test to `internal/agentrt/control_tools_test.go`:

```go
type recordingAgentRunSubmitter struct {
	ctx    context.Context
	prompt string
	origin background.Origin
}

func (s *recordingAgentRunSubmitter) Submit(ctx context.Context, prompt string, origin background.Origin) (string, error) {
	s.ctx = ctx
	s.prompt = prompt
	s.origin = origin
	return origin.Session, nil
}

func TestAgentSpawn_SubmitsWithTeammateRuntimeContext(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	submitter := &recordingAgentRunSubmitter{}
	runtime := NewCapabilityRuntime(store, &CapabilityPolicy{}, func(string) agent.SafetyLevel {
		return agent.SafetyLevelDangerous
	})
	cp := &AgentControlPlane{
		RunStore:          store,
		Projection:        NewAgentRunProjection(store),
		Submitter:         submitter,
		CapabilityRuntime: runtime,
	}
	tools := BuildControlTools(cp)
	spawnTool := findControlTool(t, tools, "agent_spawn")

	ctx := session.WithSessionKey(context.Background(), "sess-parent")
	result, err := spawnTool.call(ctx, map[string]interface{}{
		"instruction":   "inspect files",
		"agent":         "operator",
		"allowed_tools": []interface{}{"fs_read"},
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.NotEmpty(t, m["agent_id"])
	assert.Equal(t, "sess-parent", submitter.origin.Session)
	assert.Equal(t, "agent_control", submitter.origin.Channel)
	assert.Contains(t, submitter.prompt, "inspect files")
	assert.Equal(t, "operator", ctxkeys.AgentNameFromContext(submitter.ctx))
	assert.Equal(t, []string{"fs_read"}, ctxkeys.DynamicAllowedToolsFromContext(submitter.ctx))
}
```

Update imports explicitly:

- `internal/agentrt/control_tools.go`: add `github.com/langoai/lango/internal/background` and `github.com/langoai/lango/internal/session`.
- `internal/agentrt/control_tools_test.go`: add `github.com/langoai/lango/internal/background` and `github.com/langoai/lango/internal/session`; `ctxkeys` is already imported in the existing file.

- [ ] **Step 5: Locate existing app wiring symbols**

Run:

```bash
rg -n 'agentRunStore|agentRunProjection|controlPlane|bg :=|catalog\.|GetToolSafetyLevel|background\.NewManager' internal/app/modules.go internal/toolcatalog/catalog.go
```

Expected: `internal/app/modules.go` shows `bg`, `agentRunStore`, `agentRunProjection`, and `controlPlane`; `internal/toolcatalog/catalog.go` shows existing `GetToolSafetyLevel(name) (agent.SafetyLevel, bool)`. If any symbol differs, update the next code block to use the actual names before editing.

- [ ] **Step 6: Wire app module dependencies**

Modify `internal/app/modules.go` where the control plane is built:

```go
capabilityRuntime := agentrt.NewCapabilityRuntime(agentRunStore, &agentrt.CapabilityPolicy{}, func(toolName string) agent.SafetyLevel {
	level, _ := catalog.GetToolSafetyLevel(toolName)
	return level
})
controlPlane := &agentrt.AgentControlPlane{
	RunStore:          agentRunStore,
	Projection:        agentRunProjection,
	Submitter:         bg,
	CapabilityRuntime: capabilityRuntime,
}
```

- [ ] **Step 7: Verify import boundary before running tests**

Run:

```bash
go list -f '{{.Imports}}' ./internal/agentrt
go list -f '{{.Imports}}' ./internal/toolchain
```

Expected: `internal/agentrt` may import `internal/toolchain` after this task, but `internal/toolchain` must not import `internal/agentrt`. If both directions appear, stop and split the shared `BlockedToolCall` contract into a lower-level package before continuing.

- [ ] **Step 8: Run wiring tests**

Run:

```bash
go test ./internal/agentrt -run 'Test(CapabilityRuntime|AgentSpawn_SubmitsWithTeammateRuntimeContext)' -count=1
```

Expected: PASS.

- [ ] **Step 9: Record suggested commit capability wiring**

Suggested user-run commands after review:

```bash
git add internal/agentrt/capability_runtime.go internal/agentrt/capability_runtime_test.go internal/agentrt/control_tools.go internal/agentrt/control_tools_test.go internal/app/modules.go
git commit -m "feat: wire teammate capability projection"
```

Expected: user can run the suggested commit after review with capability runtime wiring files.

## Task 7: agent_wait Projected State Response

**Files:**
- Modify: `internal/agentrt/control_tools.go`
- Modify: `internal/agentrt/control_tools_test.go`

- [ ] **Step 1: Write failing wait projection test**

Append to `internal/agentrt/control_tools_test.go`:

```go
func TestAgentWait_TimeoutIncludesBlockedProjectionWithoutCancelling(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:               "wait-blocked",
		Status:           AgentRunRunning,
		RuntimeCondition: AgentRunConditionBlockedWaitingApproval,
		BlockedReason:    "capability request pending",
		GrantRequestID:   "grant-wait-blocked-fs_write",
	}))

	cp := &AgentControlPlane{RunStore: store}
	tools := BuildControlTools(cp)
	waitTool := findControlTool(t, tools, "agent_wait")

	result, err := waitTool.call(context.Background(), map[string]interface{}{
		"agent_id": "wait-blocked",
		"timeout":  float64(1),
	})
	require.NoError(t, err)

	m := result.(map[string]interface{})
	assert.Equal(t, true, m["timeout"])
	assert.Equal(t, "running", m["status"])
	assert.Equal(t, "blocked_waiting_approval", m["condition"])
	assert.Equal(t, "capability request pending", m["blocked_reason"])
	assert.Equal(t, "grant-wait-blocked-fs_write", m["grant_request_id"])

	run, err := store.Get("wait-blocked")
	require.NoError(t, err)
	assert.Equal(t, AgentRunRunning, run.Status)
}
```

- [ ] **Step 2: Run test and verify failure**

Run:

```bash
go test ./internal/agentrt -run TestAgentWait_TimeoutIncludesBlockedProjectionWithoutCancelling -count=1
```

Expected: FAIL because timeout responses do not include condition fields.

- [ ] **Step 3: Add response helper**

Modify `internal/agentrt/control_tools.go` by adding:

```go
func agentRunResponse(run *AgentRun) map[string]interface{} {
	resp := map[string]interface{}{
		"agent_id": run.ID,
		"status":   string(run.Status),
	}
	if run.Result != "" {
		resp["result"] = run.Result
	}
	if run.Error != "" {
		resp["error"] = run.Error
	}
	if run.RuntimeCondition != AgentRunConditionNone {
		resp["condition"] = string(run.RuntimeCondition)
	}
	if run.BlockedReason != "" {
		resp["blocked_reason"] = run.BlockedReason
	}
	if run.GrantRequestID != "" {
		resp["grant_request_id"] = run.GrantRequestID
	}
	if run.WaitingOnRunID != "" {
		resp["waiting_on_run_id"] = run.WaitingOnRunID
	}
	if run.RecoveryState != "" {
		resp["recovery_state"] = run.RecoveryState
	}
	return resp
}
```

Replace the terminal response in `buildAgentWait` with:

```go
					return agentRunResponse(run), nil
```

Replace the timeout response with:

```go
					resp := agentRunResponse(run)
					resp["timeout"] = true
					return resp, nil
```

- [ ] **Step 4: Run agentrt tests**

Run:

```bash
go test ./internal/agentrt -count=1
```

Expected: PASS.

- [ ] **Step 5: Record suggested commit wait projection response**

Suggested user-run commands after review:

```bash
git add internal/agentrt/control_tools.go internal/agentrt/control_tools_test.go
git commit -m "feat: surface teammate run projection in wait"
```

Expected: user can run the suggested commit after review with control tool files.

## Task 8: Prompt Compatibility Rule

**Files:**
- Modify: `internal/orchestration/tools.go`
- Modify: `internal/orchestration/orchestrator_test.go`

- [ ] **Step 1: Write failing prompt test**

Append to `internal/orchestration/orchestrator_test.go`:

```go
func TestOrchestratorInstruction_IncludesDynamicTeammateSelectionRule(t *testing.T) {
	got := buildOrchestratorInstruction(
		"base",
		[]routingEntry{{Name: "operator", Description: "execute", Accepts: "tasks", Returns: "results"}},
		10,
		nil,
	)

	assert.Contains(t, got, "Use agent_spawn for new dynamic teammate work.")
	assert.Contains(t, got, "Use transfer_to_agent only for legacy ADK static sub-agent fallback")
	assert.Contains(t, got, "existing remote A2A paths")
}
```

- [ ] **Step 2: Run prompt test and verify failure**

Run:

```bash
go test ./internal/orchestration -run TestOrchestratorInstruction_IncludesDynamicTeammateSelectionRule -count=1
```

Expected: FAIL because the prompt still only describes tool-less `transfer_to_agent` delegation.

- [ ] **Step 3: Add v1 selection rule**

Modify `internal/orchestration/tools.go` inside `buildOrchestratorInstruction` after the `## Your Role` lines:

```go
	b.WriteString("\n## Dynamic Teammate Runtime V1\n")
	b.WriteString("Use agent_spawn for new dynamic teammate work. Include a concise spawn_reason and the narrowest allowed_tools that can complete the task.\n")
	b.WriteString("Use transfer_to_agent only for legacy ADK static sub-agent fallback, specialist re-routing, or existing remote A2A paths.\n")
	b.WriteString("When both paths could work, prefer agent_spawn for new work because it creates an inspectable AgentRun with projection, cancellation, and wait semantics.\n")
```

Also replace the old hard statement:

```go
	b.WriteString("You do NOT have tools. You MUST delegate all tool-requiring tasks to the appropriate sub-agent using transfer_to_agent.\n")
```

with:

```go
	b.WriteString("You coordinate work, answer directly when no teammate is needed, and use runtime control tools for new dynamic teammate work.\n")
```

- [ ] **Step 4: Run orchestration tests**

Run:

```bash
go test ./internal/orchestration -count=1
```

Expected: `TestBuildOrchestratorInstruction_DelegateOnly`, `TestBuildOrchestratorInstruction_HasAssessStep`, `TestBuildOrchestratorInstruction_HasReRoutingProtocol`, or `TestBuildOrchestratorInstruction_DelegationRulesOrder` may fail if their old assertions assume a tool-less, transfer-only orchestrator.

Apply these assertion updates in `internal/orchestration/orchestrator_test.go`:

```go
func TestBuildOrchestratorInstruction_DelegateOnly(t *testing.T) {
	got := buildOrchestratorInstruction("base", nil, 5, nil)

	assert.Contains(t, got, "You coordinate work, answer directly when no teammate is needed")
	assert.Contains(t, got, "Use agent_spawn for new dynamic teammate work.")
	assert.NotContains(t, got, "You do NOT have tools")
	// Output Awareness section replaces the old Diagnostics section.
	assert.Contains(t, got, "Output Awareness")
	assert.NotContains(t, got, "builtin_health")
}
```

In `TestBuildOrchestratorInstruction_HasAssessStep`, keep the direct-answer and no-function-call guard assertions unchanged. Add:

```go
assert.Contains(t, got, "Use agent_spawn for new dynamic teammate work.")
```

In `TestBuildOrchestratorInstruction_HasReRoutingProtocol`, keep the re-routing assertions unchanged and add:

```go
assert.Contains(t, got, "Use transfer_to_agent only for legacy ADK static sub-agent fallback")
```

In `TestBuildOrchestratorInstruction_DelegationRulesOrder`, replace the old delegation search:

```go
delegateIdx := strings.Index(got, "delegate to the sub-agent")
```

with:

```go
delegateIdx := strings.Index(got, "Use agent_spawn for new dynamic teammate work.")
```

- [ ] **Step 5: Run orchestration tests again**

Run:

```bash
go test ./internal/orchestration -count=1
```

Expected: PASS.

- [ ] **Step 6: Record suggested commit prompt guidance**

Suggested user-run commands after review:

```bash
git add internal/orchestration/tools.go internal/orchestration/orchestrator_test.go
git commit -m "feat: guide dynamic teammate routing"
```

Expected: user can run the suggested commit after review with orchestration prompt files.

## Task 9: CLI Status And Public Docs

**Files:**
- Modify: `internal/cli/agent/status.go`
- Test: existing CLI tests if present under `internal/cli/agent`
- Modify: `docs/features/multi-agent.md`
- Modify: `openspec/changes/production-teammate-runtime/tasks.md`

- [ ] **Step 1: Write failing status JSON test**

Create `internal/cli/agent/status_test.go` with:

```go
package agent

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestStatusCmd_JSONIncludesTeammateRuntime(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Agent.Provider = "test-provider"
		cfg.Agent.Model = "test-model"
		return cfg, nil
	})
	cmd.SetArgs([]string{"--json"})

	output := captureStdout(t, func() {
		require.NoError(t, cmd.Execute())
	})

	assert.Contains(t, output, `"teammate_runtime": "dynamic-v1"`)
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	fn()

	require.NoError(t, w.Close())
	os.Stdout = old
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	return buf.String()
}
```

- [ ] **Step 2: Run status test and verify failure**

Run:

```bash
go test ./internal/cli/agent -run TestStatusCmd_JSONIncludesTeammateRuntime -count=1
```

Expected: FAIL because JSON output lacks `teammate_runtime`.

- [ ] **Step 3: Add teammate runtime to status output**

Modify `internal/cli/agent/status.go` by adding field to `statusOutput`:

```go
				TeammateRuntime       string       `json:"teammate_runtime,omitempty"`
```

Set it after `mode` is computed:

```go
			teammateRuntime := ""
			if cfg.Agent.MultiAgent {
				teammateRuntime = "dynamic-v1"
			}
```

Add it to `s := statusOutput{...}`:

```go
				TeammateRuntime:       teammateRuntime,
```

Add table output after `Multi-Agent`:

```go
			if s.TeammateRuntime != "" {
				fmt.Printf("  Teammate Runtime:  %s\n", s.TeammateRuntime)
			}
```

- [ ] **Step 4: Run CLI status tests**

Run:

```bash
go test ./internal/cli/agent -count=1
```

Expected: PASS.

- [ ] **Step 5: Update public multi-agent docs**

Modify `docs/features/multi-agent.md` so it says:

```markdown
## Dynamic Teammate Runtime V1

When `agent.multiAgent` is enabled, Lango uses the dynamic teammate runtime v1 for new teammate work. The model-facing control tools remain `agent_spawn`, `agent_wait`, and `agent_stop`.

- `agent_spawn` creates an inspectable `AgentRun` with teammate type, spawn reason, child-session key when isolation is active, and spawn-time `allowed_tools`.
- `agent_wait` returns terminal results when a teammate completes. If a teammate is blocked waiting for approval and the wait call times out, the run is not cancelled; the response remains non-terminal and includes projected blocked state.
- `agent_stop` cancels the run through the existing `AgentRunStore` cancellation path.

The existing `transfer_to_agent` path remains available for legacy ADK static sub-agent fallback, specialist re-routing, and existing remote A2A routing in v1. New dynamic teammate work should prefer `agent_spawn`.

Role maximum scope is enforced before a spawned teammate receives `allowed_tools`. Runtime capability escalation can request additional tools only inside the teammate type's maximum scope, and dangerous tools require policy or user approval.
```

- [ ] **Step 6: Mark OpenSpec tasks complete as implemented**

Edit `openspec/changes/production-teammate-runtime/tasks.md` by changing completed items from `- [ ]` to `- [x]` for the runtime tasks implemented so far.

- [ ] **Step 7: Run docs and CLI checks**

Run:

```bash
go test ./internal/cli/agent ./internal/agentrt ./internal/toolchain ./internal/orchestration -count=1
```

Expected: PASS.

- [ ] **Step 8: Record suggested commit CLI and docs**

Suggested user-run commands after review:

```bash
git add internal/cli/agent/status.go internal/cli/agent/status_test.go docs/features/multi-agent.md openspec/changes/production-teammate-runtime/tasks.md
git commit -m "docs: describe dynamic teammate runtime v1"
```

Expected: user can run the suggested commit after review with CLI status, docs, and task checklist changes.

## Task 10: Full Verification And OpenSpec Archive

**Files:**
- Modify: `openspec/specs/*` through OpenSpec archive
- Move: `openspec/changes/production-teammate-runtime` to archive through OpenSpec archive

- [ ] **Step 1: Run focused package tests**

Run:

```bash
go test ./internal/agentrt ./internal/toolchain ./internal/orchestration ./internal/cli/agent -count=1
```

Expected: PASS.

- [ ] **Step 2: Run full test suite**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 3: Run full build**

Run:

```bash
go build ./...
```

Expected: PASS.

- [ ] **Step 4: Verify OpenSpec change**

Run:

```bash
openspec status --change production-teammate-runtime
openspec validate production-teammate-runtime --strict
```

Expected: status shows all artifacts present, and strict validation passes.

- [ ] **Step 5: Use OpenSpec verify skill**

Invoke `superpowers:verification-before-completion` and `.codex/skills/openspec-verify-change/SKILL.md` workflow for `production-teammate-runtime`.

Expected: verification reports no critical issues. Any warning with a missing test or spec mismatch is fixed before continuing.

- [ ] **Step 6: Archive change and update main specs**

Run:

```bash
openspec archive -y production-teammate-runtime
```

Expected: affected main specs under `openspec/specs/` update with the approved delta requirements, and the change moves to `openspec/changes/archive/<date>-production-teammate-runtime/`.

- [ ] **Step 7: Verify archived OpenSpec state**

Run:

```bash
openspec list
openspec validate --all
```

Expected: the active change is gone, archived specs validate, and no `production-teammate-runtime` active change remains.

- [ ] **Step 8: Record suggested commit archive**

Suggested user-run commands after review:

```bash
git add openspec
git commit -m "spec: archive production teammate runtime"
```

Expected: user can run the suggested commit after review with only OpenSpec archive and main-spec changes.

## Self-Review

### Spec Coverage

- Dynamic teammate run creation under `agent.multiAgent=true`: Task 1, Task 3, Task 6, Task 8.
- Main-agent direct answer and spawn decision protocol: Task 1, Task 8.
- Spawn reason audit/projection: Task 2, Task 3, Task 9.
- Role maximum scope plus spawn-time `AllowedTools`: Task 3.
- Capability request and policy-first approval: Task 4, Task 5, Task 6.
- ChildSession isolation: preserved by design and OpenSpec scope; no new code in this plan changes the existing ChildSession primitive.
- RunLedger/background projection: preserved by Task 2; this v1 plan does not add durable RunLedger columns.
- CLI inspection: Task 9 covers CLI status. Cockpit/TUI projection is explicitly out of this v1 plan and requires a follow-up plan once the runtime exposes a stable shared read model.
- Recovery behavior: Task 2 adds projection fields; existing `RecoveryPolicy` remains the recovery authority.
- Worker process/sandbox: excluded by scope.

### Marker Scan

The plan was scanned for banned marker patterns and vague test instructions; none remain.

### Type Consistency

- `AgentRunConditionBlockedWaitingApproval` is defined in Task 2 and reused in Tasks 2 and 6.
- `RunProjectionPatch` is defined in Task 2 and used through `AgentRunStore.UpdateProjection` in Tasks 2 and 6.
- `SpawnReason` is added to `AgentRun` in Task 2 and used by `agent_spawn` in Task 3.
- `BlockedToolCall` and `WithBlockedToolCallSink` are defined in Task 4, emitted by `WithHooks`, and consumed by `CapabilityRuntime` in Task 6.
- `CapabilityPolicy`, `CapabilityRequest`, and `CapabilityDecisionKind` are defined in Task 5 and first consumed by `CapabilityRuntime` in Task 6.
- Existing `toolcatalog.Catalog.GetToolSafetyLevel` is verified in Task 6 before `internal/app/modules.go` wires the lookup into `CapabilityRuntime`.
