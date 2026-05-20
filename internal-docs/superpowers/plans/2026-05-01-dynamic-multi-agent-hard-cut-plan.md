# Dynamic Multi-Agent Hard Cut Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Hard-cut built-in multi-agent execution to the `agent_spawn` runtime path, remove built-in production reliance on `transfer_to_agent`, preserve remote A2A as a separate execution model, and close the teammate runtime race/observability gaps identified in the design.

**Architecture:** Rewrite the OpenSpec contracts first, then cut over the built-in orchestration/runtime path in code. Built-in teammate execution will flow through `AgentRun -> background.Manager -> ChildSession -> CapabilityRuntime -> RunLedger`, while remote A2A remains a distinct path. The cutover also rewrites embedded `AGENT.md` prompts, tightens ADK hallucinated-agent recovery, and updates downstream CLI/docs to match the new contract.

**Tech Stack:** Go, Cobra CLI, OpenSpec, `internal/orchestration`, `internal/agentrt`, `internal/adk`, `internal/skill`, `internal/agentregistry`, `internal/background`, `internal/runledger`, `go test ./...`, `go build ./...`.

---

## Source Design

- Design spec: `internal-docs/superpowers/specs/2026-05-01-dynamic-multi-agent-hard-cut-design.md`
- Existing runtime baseline: `internal-docs/superpowers/plans/2026-05-01-production-teammate-runtime-v1-plan.md`

## Scope Check

This plan covers one coherent contract change: built-in teammate hard cut with remote A2A split semantics. It intentionally keeps remote A2A execution separate rather than trying to absorb it into the built-in control plane. That keeps the work within one change that can still be archived with matching code and specs.

## File Map

- Create `openspec/changes/dynamic-multi-agent-hard-cut/proposal.md`: change summary and blast radius.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/design.md`: working OpenSpec design copy with cutover decisions.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/tasks.md`: archive checklist.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/specs/multi-agent-orchestration/spec.md`: built-in hard-cut spec delta.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-control-plane-tools/spec.md`: `agent_spawn`-anchored built-in runtime contract.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-routing/spec.md`: remove built-in `transfer_to_agent("lango-orchestrator")` requirement from prompt files.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-registry/spec.md`: embedded `AGENT.md` contract changes.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/specs/adk-architecture/spec.md`: narrow hallucinated-agent retry to remote/legacy transfer paths.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/specs/tool-capability-layer/spec.md`: grant/recheck semantics.
- Create `openspec/changes/dynamic-multi-agent-hard-cut/specs/run-ledger/spec.md`: durable visibility expectations and audit verdict shape.
- Modify `internal/orchestration/tools.go`: remove built-in transfer wording, tighten routing/disambiguation, require spawn-only built-in execution.
- Modify `internal/orchestration/orchestrator.go`: retain root orchestration plus remote/legacy sub-agent composition only.
- Modify `internal/orchestration/orchestrator_test.go`: lock new prompt and root-only-tree behavior.
- Modify `internal/agentregistry/defaults/{operator,navigator,vault,librarian,automator,planner,chronicler,ontologist}/AGENT.md`: remove built-in transfer escalation and replace with root-runtime-compatible completion/escalation language.
- Modify `internal/adk/agent.go`: reframe hallucinated-agent retry messaging away from built-in transfer retries.
- Modify `internal/adk/agent_test.go`: cover built-in vs remote/legacy retry semantics.
- Modify `internal/skill/executor.go`: verify built-in resolution, switch built-in fork guidance to `agent_spawn`, keep transfer wording only where still valid.
- Modify `internal/skill/executor_test.go`: cover built-in default `operator` path and remote/legacy path.
- Modify `internal/agentrt/capability_runtime.go`: add latest-run optimistic recheck before blocked projection write.
- Modify `internal/agentrt/capability_runtime_test.go`: cover final-state stale-block elimination.
- Modify `internal/app/wiring.go`: keep orchestrator tree composition coherent after built-in sub-agents leave the production tree.
- Inspect and modify `internal/app/modules.go` only if the `agent_control` audit finds category or registration mismatches that must be fixed during the hard cut.
- Modify `internal/cli/agent/status.go`: keep runtime reporting honest after hard cut.
- Modify `docs/features/multi-agent.md`: describe built-in hard cut and remote A2A split honestly.
- Modify `docs/features/agent-format.md`: add upgrade note for copied custom `AGENT.md` files.

## Commit Policy

Each task ends with a suggested commit message. When executing, stage only the files listed for that task and let the user decide whether to create the commit immediately.

## Task 1: OpenSpec Change Skeleton and Inventory

**Files:**
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/proposal.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/design.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/tasks.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/specs/multi-agent-orchestration/spec.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-control-plane-tools/spec.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-routing/spec.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-registry/spec.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/specs/adk-architecture/spec.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/specs/tool-capability-layer/spec.md`
- Create: `openspec/changes/dynamic-multi-agent-hard-cut/specs/run-ledger/spec.md`

- [ ] **Step 1: Confirm the change name is free**

Run:

```bash
openspec list
```

Expected: no active change named `dynamic-multi-agent-hard-cut`.

- [ ] **Step 2: Create the change directories**

Run:

```bash
openspec new change dynamic-multi-agent-hard-cut
mkdir -p openspec/changes/dynamic-multi-agent-hard-cut/specs/multi-agent-orchestration
mkdir -p openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-control-plane-tools
mkdir -p openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-routing
mkdir -p openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-registry
mkdir -p openspec/changes/dynamic-multi-agent-hard-cut/specs/adk-architecture
mkdir -p openspec/changes/dynamic-multi-agent-hard-cut/specs/tool-capability-layer
mkdir -p openspec/changes/dynamic-multi-agent-hard-cut/specs/run-ledger
```

Expected: `openspec/changes/dynamic-multi-agent-hard-cut/` exists with all spec subdirectories.

- [ ] **Step 3: Record the grep inventory before drafting delta specs**

Run:

```bash
rg -n "transfer_to_agent|lango-orchestrator|failed to find agent|sub-agent escalation" \
  openspec/specs/multi-agent-orchestration/spec.md \
  openspec/specs/agent-control-plane-tools/spec.md \
  openspec/specs/agent-routing/spec.md \
  openspec/specs/agent-registry/spec.md \
  openspec/specs/adk-architecture/spec.md \
  openspec/specs/tool-capability-layer/spec.md \
  openspec/specs/run-ledger/spec.md
```

Expected: concrete requirement lines to carry into the change design.

- [ ] **Step 4: Write the OpenSpec `design.md` inventory section**

Add this section to `openspec/changes/dynamic-multi-agent-hard-cut/design.md`:

```markdown
## Contract Inventory

- `multi-agent-orchestration`: built-in production path still references legacy transfer compatibility.
- `agent-control-plane-tools`: built-in teammate runtime already exists but is not yet the only production path.
- `agent-routing`: embedded prompt files still require `transfer_to_agent("lango-orchestrator")`.
- `agent-registry`: embedded `AGENT.md` defaults remain part of the production prompt contract.
- `adk-architecture`: `failed to find agent` retry still assumes a useful sub-agent list exists.
- `tool-capability-layer`: grant/recheck semantics must be aligned with the hard cut.
- `run-ledger`: durable visibility expectations must be explicit for built-in teammate runs.
```

- [ ] **Step 5: Write the OpenSpec `proposal.md`**

Create `openspec/changes/dynamic-multi-agent-hard-cut/proposal.md` with:

```markdown
# Dynamic Multi-Agent Hard Cut

## Problem

Built-in multi-agent execution still mixes two incompatible models:

1. the control-plane teammate runtime based on `agent_spawn`
2. legacy ADK static delegation based on `transfer_to_agent`

This ambiguity leaks into prompts, embedded `AGENT.md` files, skill execution guidance, hallucinated-agent recovery, and operator surfaces. It also leaves capability-runtime and RunLedger observability gaps unresolved.

## Proposed Change

Make built-in teammate execution spawn-only in production. Remove built-in reliance on `transfer_to_agent`, keep remote A2A separate, narrow ADK recovery to the remaining remote/legacy transfer surface, tighten capability-runtime blocked-call behavior, and update downstream docs and CLI surfaces to match the new contract.

## User-Facing Impact

Built-in teammate work routes through `agent_spawn`-backed execution only. Remote A2A remains available as a separate compatibility path. Copied custom `AGENT.md` files that still encode built-in `transfer_to_agent("lango-orchestrator")` behavior require upgrade guidance.
```

- [ ] **Step 6: Write the OpenSpec `tasks.md`**

Create `openspec/changes/dynamic-multi-agent-hard-cut/tasks.md` with:

```markdown
# Tasks

- [ ] Rewrite the hard-cut OpenSpec contracts
- [ ] Remove built-in production transfer dependencies from orchestration and embedded prompts
- [ ] Reframe ADK hallucinated-agent recovery for remote/legacy transfer only
- [ ] Rework skill fork guidance for built-in teammates
- [ ] Narrow the capability-runtime race window and verify final-state behavior
- [ ] Verify root-only orchestration behavior and patch wiring if needed
- [ ] Audit RunLedger and `agent_control` observability
- [ ] Update CLI/docs/upgrade notes
- [ ] Run `openspec validate dynamic-multi-agent-hard-cut --strict`
- [ ] Run `go build ./...`
- [ ] Run `go test ./...`
- [ ] Archive the change with `openspec archive`
```

- [ ] **Step 7: Suggested commit**

Suggested commit message:

```bash
docs: scaffold hard cut openspec change
```

## Task 2: Rewrite the Contract Deltas

**Files:**
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/specs/multi-agent-orchestration/spec.md`
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-control-plane-tools/spec.md`
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-routing/spec.md`
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-registry/spec.md`
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/specs/adk-architecture/spec.md`
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/specs/tool-capability-layer/spec.md`
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/specs/run-ledger/spec.md`

- [ ] **Step 1: Write the failing spec validation baseline**

Run:

```bash
openspec validate dynamic-multi-agent-hard-cut --strict
```

Expected: FAIL before the spec deltas are complete, typically because the change skeleton exists but one or more required delta requirements are still missing.

- [ ] **Step 2: Add the `multi-agent-orchestration` delta**

Include this requirement block in `openspec/changes/dynamic-multi-agent-hard-cut/specs/multi-agent-orchestration/spec.md`:

```markdown
## MODIFIED Requirements

### Requirement: Hierarchical agent tree with sub-agents
Built-in production execution SHALL use the control-plane teammate runtime rather than static ADK specialist delegation. The built-in teammate set is `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`, as resolved by `BuiltinTeammateTypes()`. Remote A2A agents remain a separate execution model.

#### Scenario: Built-in work uses spawn-only production execution
- **WHEN** the runtime routes built-in specialist work under multi-agent mode
- **THEN** the production path SHALL begin with `agent_spawn`
- **AND** built-in `transfer_to_agent` delegation SHALL NOT be required

#### Scenario: Remote A2A remains separate
- **WHEN** a configured remote A2A agent is selected
- **THEN** the runtime MAY still use the remote compatibility path
- **AND** this SHALL NOT re-open built-in static delegation as the normal production path
```

- [ ] **Step 3: Add the `agent-routing` delta**

Include this requirement block in `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-routing/spec.md`:

```markdown
## MODIFIED Requirements

### Requirement: Prompt override file consistency
Embedded prompt override files for built-in teammates SHALL NOT require `transfer_to_agent("lango-orchestrator")` as the built-in escalation path after the hard cut. Remote/legacy transfer guidance may remain only where explicitly documented as compatibility behavior.

#### Scenario: Built-in AGENT.md files no longer encode built-in transfer escalation
- **WHEN** any embedded built-in `AGENT.md` file is checked
- **THEN** it SHALL NOT instruct the built-in teammate to call `transfer_to_agent("lango-orchestrator")`
```

- [ ] **Step 4: Add the `adk-architecture` delta**

Include this requirement block in `openspec/changes/dynamic-multi-agent-hard-cut/specs/adk-architecture/spec.md`:

```markdown
## MODIFIED Requirements

### Requirement: Agent hallucination retry in RunAndCollect
`RunAndCollect` SHALL continue to recover from hallucinated transfer targets only for the remaining remote/legacy transfer surface. It SHALL no longer teach built-in retry behavior through static sub-agent names.

#### Scenario: Built-in hallucinated target produces spawn-oriented recovery
- **WHEN** built-in routing fails with a hallucinated target name
- **THEN** the recovery message SHALL steer the root runtime toward `agent_spawn` or direct root answer behavior
- **AND** it SHALL NOT suggest retrying a built-in `transfer_to_agent` target
```

- [ ] **Step 5: Add the `run-ledger` audit verdict delta**

Include this requirement block in `openspec/changes/dynamic-multi-agent-hard-cut/specs/run-ledger/spec.md`:

```markdown
## MODIFIED Requirements

### Requirement: Built-in teammate durability audit
Before archive, the implementation SHALL classify built-in teammate durability for spawn submission, run status transitions, projection sync markers, approval-blocked conditions, and recovery states using one of three verdicts: recorded, not recorded but harmless, or not recorded and follow-up required.
```

- [ ] **Step 6: Add the `agent-control-plane-tools` delta**

Include this requirement block in `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-control-plane-tools/spec.md`:

```markdown
## MODIFIED Requirements

### Requirement: agent_spawn is the built-in production entrypoint
Built-in teammate execution SHALL enter production execution through `agent_spawn`. `RequestedAgent` SHALL identify the built-in teammate type, and built-in production execution SHALL NOT depend on static ADK `transfer_to_agent` routing.

#### Scenario: Built-in teammate spawn remains the production entrypoint
- **WHEN** built-in specialist work is delegated
- **THEN** `agent_spawn` SHALL create the run
- **AND** `agent_wait` / `agent_stop` SHALL operate on that run identity chain
```

- [ ] **Step 7: Add the `agent-registry` delta**

Include this requirement block in `openspec/changes/dynamic-multi-agent-hard-cut/specs/agent-registry/spec.md`:

```markdown
## MODIFIED Requirements

### Requirement: Embedded AGENT.md defaults match hard-cut built-in runtime behavior
The embedded built-in `AGENT.md` files SHALL no longer encode built-in escalation through `transfer_to_agent("lango-orchestrator")`.

#### Scenario: Embedded built-in AGENT.md files no longer require built-in transfer escalation
- **WHEN** the embedded built-in `AGENT.md` files are inspected
- **THEN** their escalation behavior SHALL match the spawn-only built-in runtime contract
```

- [ ] **Step 8: Add the `tool-capability-layer` delta**

Include this requirement block in `openspec/changes/dynamic-multi-agent-hard-cut/specs/tool-capability-layer/spec.md`:

```markdown
## MODIFIED Requirements

### Requirement: Capability runtime rechecks grant state before blocked projection
Before persisting a `blocked_waiting_approval` projection for a built-in teammate run, the runtime SHALL re-read the latest grant and allowlist state. The check narrows the stale projection window but correctness is measured on the final observed run state rather than every intermediate transition.

#### Scenario: Final observed run state is clear after grant wins the race
- **WHEN** a grant becomes effective before the final observed state is read
- **THEN** the final observed teammate run state SHALL NOT remain stuck in `blocked_waiting_approval`
- **AND** the runtime regression tests SHALL be the concrete verification gate for that condition
```

- [ ] **Step 9: Re-run spec validation**

Run:

```bash
openspec validate dynamic-multi-agent-hard-cut --strict
```

Expected: PASS for the spec-only delta set, or a narrowed set of actionable failures.

- [ ] **Step 10: Suggested commit**

Suggested commit message:

```bash
docs: rewrite hard cut spec contracts
```

## Task 3: Cut Over Built-In Orchestration and Embedded Prompts

**Files:**
- Modify: `internal/orchestration/tools.go`
- Modify: `internal/orchestration/orchestrator.go`
- Modify: `internal/orchestration/orchestrator_test.go`
- Modify: `internal/agentregistry/defaults/operator/AGENT.md`
- Modify: `internal/agentregistry/defaults/navigator/AGENT.md`
- Modify: `internal/agentregistry/defaults/vault/AGENT.md`
- Modify: `internal/agentregistry/defaults/librarian/AGENT.md`
- Modify: `internal/agentregistry/defaults/automator/AGENT.md`
- Modify: `internal/agentregistry/defaults/planner/AGENT.md`
- Modify: `internal/agentregistry/defaults/chronicler/AGENT.md`
- Modify: `internal/agentregistry/defaults/ontologist/AGENT.md`

- [ ] **Step 1: Write the failing orchestrator tests**

Add these assertions to `internal/orchestration/orchestrator_test.go`:

```go
func TestOrchestratorInstruction_BuiltinUsesSpawnOnly(t *testing.T) {
	got := buildOrchestratorInstruction("base", nil, 10, nil)
	assert.Contains(t, got, "Built-in teammate work MUST use agent_spawn")
	assert.NotContains(t, got, "Use transfer_to_agent only for legacy ADK static sub-agent fallback, specialist re-routing")
}

func TestBuildAgentTree_RootOnlyAllowed(t *testing.T) {
	root, err := BuildAgentTree(Config{
		Tools:      nil,
		Model:      nil,
		SystemPrompt: "base",
		AdaptTool:  stubAdapter,
		Specs:      []AgentSpec{},
	})
	require.NoError(t, err)
	assert.Len(t, root.SubAgents(), 0)
}
```

- [ ] **Step 2: Run the focused tests**

Run:

```bash
go test ./internal/orchestration -run 'TestOrchestratorInstruction_BuiltinUsesSpawnOnly|TestBuildAgentTree_RootOnlyAllowed' -count=1
```

Expected: FAIL before the prompt/tree changes.

- [ ] **Step 3: Inventory current embedded escalation blocks**

Run:

```bash
rg -n "## Escalation Protocol|transfer_to_agent|## Output Handling" internal/agentregistry/defaults/*/AGENT.md
```

Expected: all 8 embedded built-in `AGENT.md` files are listed, and planner is visibly different from the tool-bearing specialists.

- [ ] **Step 4: Rewrite the built-in prompt language**

Change `internal/orchestration/tools.go` so the root guidance looks like:

```go
b.WriteString("Built-in teammate work MUST use agent_spawn. Include a concise spawn_reason and the narrowest allowed_tools that can complete the task.\n")
b.WriteString("Do not use transfer_to_agent for built-in teammates.\n")
b.WriteString("Use transfer_to_agent only for remote A2A or tightly documented legacy compatibility paths.\n")
```

And rewrite built-in specialist escalation sections from:

```markdown
3. IMMEDIATELY call transfer_to_agent with agent_name "lango-orchestrator".
```

to:

```markdown
3. Return control cleanly to the root runtime by ending with a short visible escalation summary.
4. Do not call transfer_to_agent for built-in teammate escalation.
```

- [ ] **Step 5: Rewrite the embedded `AGENT.md` escalation block**

Apply this block to the 7 tool-bearing built-in `AGENT.md` files:

```markdown
## Escalation Protocol
If a task does not match your capabilities:
1. Do NOT attempt to answer or explain why you cannot help.
2. Output ONE short sentence summarizing what you tried or why you are escalating.
3. End the turn without calling transfer_to_agent.
4. Never claim that a tool or action completed unless you have direct evidence from this turn.
```

Apply this planner-specific block to `internal/agentregistry/defaults/planner/AGENT.md`:

```markdown
## Escalation Protocol
If a task does not match your capabilities:
1. Do NOT attempt to answer or explain why you cannot help.
2. Output ONE short sentence explaining why you are escalating.
3. End the turn without calling transfer_to_agent.
4. Never transfer silently.
```

- [ ] **Step 6: Run the orchestration package tests**

Run:

```bash
go test ./internal/orchestration -count=1
```

Expected: PASS.

- [ ] **Step 7: Suggested commit**

Suggested commit message:

```bash
refactor: hard cut built-in orchestration prompts
```

## Task 4: Reframe ADK Hallucinated-Agent Recovery

**Files:**
- Modify: `internal/adk/agent.go`
- Modify: `internal/adk/agent_test.go`

- [ ] **Step 1: Add a failing recovery-message test**

First extract the correction-message construction into a helper in `internal/adk/agent.go` that accepts a caller-supplied built-in predicate:

```go
func buildMissingAgentCorrection(
	badAgent string,
	subAgents []string,
	isBuiltIn func(string) bool,
) string {
	if isBuiltIn != nil && isBuiltIn(badAgent) {
		return fmt.Sprintf(
			"[System: Built-in agent %q does not exist as a transfer target. "+
				"Do not retry built-in transfer_to_agent routing. "+
				"Use agent_spawn for built-in teammate work or answer directly from gathered evidence.]",
			badAgent,
		)
	}
	return fmt.Sprintf(
		"[System: Agent %q does not exist. Valid agents: %s. Please retry using one of the valid agent names listed above.]",
		badAgent, strings.Join(subAgents, ", "),
	)
}
```

Then add this test to `internal/adk/agent_test.go`:

```go
func TestBuildMissingAgentCorrection_BuiltinDoesNotSuggestTransferRetry(t *testing.T) {
	got := buildMissingAgentCorrection(
		"web_search",
		[]string{"operator", "vault"},
		func(name string) bool { return name == "web_search" },
	)
	assert.NotContains(t, got, "Valid agents:")
	assert.Contains(t, got, "Use agent_spawn")
	assert.Contains(t, got, "answer directly from gathered evidence")
}
```

- [ ] **Step 2: Run the focused ADK test**

Run:

```bash
go test ./internal/adk -run TestBuildMissingAgentCorrection_BuiltinDoesNotSuggestTransferRetry -count=1
```

Expected: FAIL before the recovery text changes because the helper or the built-in messaging does not exist yet.

- [ ] **Step 3: Reframe the recovery branch**

Change the hallucinated-agent branch in `internal/adk/agent.go` to:

```go
names := subAgentNames(a.adkAgent)
correction := buildMissingAgentCorrection(badAgent, names, func(name string) bool {
	return containsBuiltinTargetName(name)
})
```

Then introduce a local helper in `internal/adk/agent.go` rather than importing `internal/agentrt`:

```go
func containsBuiltinTargetName(name string) bool {
	switch name {
	case "operator", "navigator", "vault", "librarian", "automator", "planner", "chronicler", "ontologist":
		return true
	default:
		return false
	}
}
```

This avoids an `internal/adk -> internal/agentrt` import cycle. Keep the remote/legacy sub-agent list branch only where a remaining compatibility path still exists.

- [ ] **Step 4: Run the ADK package tests**

Run:

```bash
go test ./internal/adk -count=1
```

Expected: PASS.

- [ ] **Step 5: Suggested commit**

Suggested commit message:

```bash
fix: reframe built-in hallucinated-agent recovery
```

## Task 5: Rework Skill Fork Guidance

**Files:**
- Modify: `internal/skill/executor.go`
- Modify: `internal/skill/executor_test.go`
- Test/Reference: `internal/agentrt/teammate_types.go`

Note: `internal/skill/executor.go` defaults `skill.Agent` to `"operator"` when unset. That means most fork-style skill executions currently follow the built-in path, so this task is high-impact rather than peripheral cleanup.

- [ ] **Step 1: Add failing built-in and remote fork tests**

Add these tests to `internal/skill/executor_test.go`:

```go
func TestExecuteFork_DefaultBuiltinUsesAgentSpawnGuidance(t *testing.T) {
	e := NewExecutor(zap.NewNop().Sugar())
	got, err := e.executeFork(SkillEntry{
		Name: "fork-default",
		Type: "fork",
		Definition: map[string]interface{}{"instruction": "inspect files"},
	}, nil)
	require.NoError(t, err)
	text := got.(string)
	assert.Contains(t, text, "Please use agent_spawn")
	assert.NotContains(t, text, "transfer_to_agent('operator')")
}

func TestExecuteFork_RemoteTargetKeepsTransferGuidance(t *testing.T) {
	e := NewExecutor(zap.NewNop().Sugar())
	got, err := e.executeFork(SkillEntry{
		Name:  "fork-remote",
		Type:  "fork",
		Agent: "remote-researcher",
		Definition: map[string]interface{}{"instruction": "ask remote"},
	}, nil)
	require.NoError(t, err)
	text := got.(string)
	assert.Contains(t, text, "transfer_to_agent('remote-researcher')")
}
```

- [ ] **Step 2: Run the focused skill tests**

Run:

```bash
go test ./internal/skill -run 'TestExecuteFork_DefaultBuiltinUsesAgentSpawnGuidance|TestExecuteFork_RemoteTargetKeepsTransferGuidance' -count=1
```

Expected: FAIL before the fork guidance split.

- [ ] **Step 3: Add built-in resolution and split the guidance**

Change `internal/skill/executor.go` to:

```go
import "github.com/langoai/lango/internal/agentrt"

builtin := false
if _, ok := agentrt.BuiltinTeammateTypes()[agentName]; ok {
	builtin = true
}

if builtin {
	result := fmt.Sprintf(`[Fork Skill Result]
This task should be delegated to the '%s' built-in teammate.

Instruction: %s

Parameters:
%s

Advisory tool restrictions: %s
(Note: tool restrictions are enforced only when using agent_spawn)

Please use agent_spawn with agent "%s".`, agentName, instruction, paramSection, advisoryTools, agentName)
	return result, nil
}
```

Keep the old transfer wording only in the non-built-in branch.

- [ ] **Step 4: Run the skill package tests**

Run:

```bash
go test ./internal/skill -count=1
```

Expected: PASS.

- [ ] **Step 5: Suggested commit**

Suggested commit message:

```bash
refactor: switch built-in skill forks to agent_spawn guidance
```

## Task 6: Narrow the Capability-Runtime Race Window

**Files:**
- Modify: `internal/agentrt/capability_runtime.go`
- Modify: `internal/agentrt/capability_runtime_test.go`

- [ ] **Step 1: Add a failing recheck-specific regression test**

Add this test to `internal/agentrt/capability_runtime_test.go`:

```go
func TestHandleBlockedToolCall_AllowedToolExpansionSkipsBlockedProjection(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-1",
		RequestedAgent: "operator",
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, nil)
	require.NoError(t, store.UpdateProjection("arun-1", RunProjectionPatch{
		AddAllowedTool: "fs_write",
	}))
	require.NoError(t, rt.HandleBlockedToolCall("arun-1", toolchain.BlockedToolCall{
		ToolName:    "fs_write",
		BlockReason: dynamicAllowedToolsBlockReason,
	}))

	run, err := store.Get("arun-1")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
}
```

- [ ] **Step 2: Run the focused runtime test**

Run:

```bash
go test ./internal/agentrt -run TestHandleBlockedToolCall_AllowedToolExpansionSkipsBlockedProjection -count=1
```

Expected: FAIL before the optimistic recheck is added.

- [ ] **Step 3: Add a concurrent final-state regression test**

Add this second test to `internal/agentrt/capability_runtime_test.go`:

```go
func TestHandleBlockedToolCall_ConcurrentGrantInterleavingLeavesNoFinalBlockedState(t *testing.T) {
	store := NewInMemoryAgentRunStore()
	require.NoError(t, store.Create(&AgentRun{
		ID:             "arun-2",
		RequestedAgent: "operator",
		Status:         AgentRunRunning,
		AllowedTools:   []string{"fs_read"},
	}))

	rt := NewCapabilityRuntime(store, &CapabilityPolicy{}, nil)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = rt.ApplyGrant("arun-2", "fs_write")
	}()

	_ = rt.HandleBlockedToolCall("arun-2", toolchain.BlockedToolCall{
		ToolName:    "fs_write",
		BlockReason: dynamicAllowedToolsBlockReason,
	})
	<-done

	run, err := store.Get("arun-2")
	require.NoError(t, err)
	assert.Equal(t, AgentRunConditionNone, run.RuntimeCondition)
}
```

This second test is not the primary recheck gate. Its purpose is to preserve the design's final-state success criterion under repeated interleavings, while Step 1 is the direct regression gate for the recheck logic itself.

- [ ] **Step 4: Implement the latest-run optimistic recheck**

Change `internal/agentrt/capability_runtime.go` inside `HandleBlockedToolCall()` to:

```go
latest, err := r.Store.Get(runID)
if err != nil {
	return err
}
if r.hasGrant(runID, call.ToolName) || containsTool(latest.AllowedTools, call.ToolName) {
	return nil
}
```

Place this immediately before the `UpdateProjection()` call for the `NeedsApproval` branch.

- [ ] **Step 5: Run the agent runtime package tests**

Run:

```bash
go test ./internal/agentrt -count=1
```

Expected: PASS.

- [ ] **Step 6: Suggested commit**

Suggested commit message:

```bash
fix: narrow teammate approval projection race
```

## Task 7: Verify Root-Only Tree and RunLedger/agent_control Audit

_Wave mapping: this task bridges late Slice 2 verification and early Slice 3 operator-surface truthfulness._

**Files:**
- Modify: `internal/app/wiring.go`
- Inspect and modify: `internal/app/modules.go` only if the `agent_control` audit reveals category or registration drift
- Modify: `internal/cli/agent/status.go`
- Create or modify: `openspec/changes/dynamic-multi-agent-hard-cut/design.md`

- [ ] **Step 1: Add a failing wiring or status test for root-only mode**

Add or update a test with this shape:

```go
func TestAgentStatus_DynamicRuntimeDoesNotRequireBuiltinSubAgents(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Background.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})
	cmd.SetArgs([]string{"--json"})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)
	assert.Contains(t, output, `"teammate_runtime": "dynamic-v1"`)
}
```

- [ ] **Step 2: Run the focused test**

Run:

```bash
go test ./internal/cli/agent -run TestAgentStatus_DynamicRuntimeDoesNotRequireBuiltinSubAgents -count=1
```

Expected: FAIL if status still implies the old built-in tree assumptions.

- [ ] **Step 3: Perform the RunLedger and `agent_control` audit**

Record this table in `openspec/changes/dynamic-multi-agent-hard-cut/design.md`:

```markdown
## RunLedger Audit Verdict

| Item | Verdict | Notes |
|------|---------|-------|
| teammate spawn submission | `recorded` / `harmless` / `follow-up` | cite the inspected file and function |
| run status transitions | `recorded` / `harmless` / `follow-up` | cite the inspected file and function |
| projection sync markers | `recorded` / `harmless` / `follow-up` | cite the inspected file and function |
| approval-blocked conditions | `recorded` / `harmless` / `follow-up` | cite the inspected file and function |
| recovery states | `recorded` / `harmless` / `follow-up` | cite the inspected file and function |
```

Use `recorded` only when there is a concrete durable or projection-sync path. Use `harmless` only when the missing durability has no operator-facing consistency risk. Use `follow-up` when the missing durability would leave operator-visible ambiguity and therefore must be fixed in the same change or explicitly tracked before archive.

- [ ] **Step 4: Verify `agent_control` propagation explicitly**

Check and record propagation across:

1. background submission origin
2. recovery surface
3. approval flow
4. CLI/TUI runtime views
5. public docs

Record the result beside the RunLedger audit in `openspec/changes/dynamic-multi-agent-hard-cut/design.md`.

- [ ] **Step 5: Patch status wording if required**

If `internal/cli/agent/status.go` still overstates built-in tree assumptions, keep the runtime calculation to the minimal truth:

```go
if cfg.Agent.MultiAgent && cfg.Background.Enabled {
	teammateRuntime = "dynamic-v1"
}
```

Do not add wording that implies remote A2A shares built-in control-plane runs unless that wiring was explicitly added. Because `config.DefaultConfig()` leaves `Background.Enabled` false, also add a short user-facing hint in docs or CLI text that `dynamic-v1` reporting requires `background.enabled: true`.

- [ ] **Step 6: Run the focused package tests**

Run:

```bash
go test ./internal/cli/agent -count=1
```

Expected: PASS.

- [ ] **Step 7: Suggested commit**

Suggested commit message:

```bash
docs: record hard cut audit verdicts
```

## Task 8: Public Docs, Full Verification, and Archive

**Files:**
- Modify: `docs/features/multi-agent.md`
- Modify: `docs/features/agent-format.md`
- Modify: `openspec/changes/dynamic-multi-agent-hard-cut/tasks.md`

- [ ] **Step 1: Update `docs/features/multi-agent.md`**

Add wording that matches the hard cut:

```markdown
Built-in teammate work now uses the `agent_spawn` runtime path in production. The older `transfer_to_agent` path remains only for remote A2A and explicitly documented legacy compatibility behavior.
```

- [ ] **Step 2: Update `docs/features/agent-format.md` upgrade notes**

Add this note:

```markdown
## Upgrade Note

If you copied or derived a custom `AGENT.md` from an older embedded default, remove built-in `transfer_to_agent("lango-orchestrator")` escalation text. Built-in teammate execution now returns control through the root runtime rather than built-in static transfer.
```

- [ ] **Step 3: Mark the OpenSpec tasks checklist complete**

Update `openspec/changes/dynamic-multi-agent-hard-cut/tasks.md` so all implemented items are checked.

- [ ] **Step 4: Run OpenSpec validation**

Run:

```bash
openspec validate dynamic-multi-agent-hard-cut --strict
```

Expected: PASS.

- [ ] **Step 5: Run the full Go verification**

Run:

```bash
go build ./...
go test ./...
```

Expected: both commands PASS.

- [ ] **Step 6: Archive the change**

Run:

```bash
openspec archive -y dynamic-multi-agent-hard-cut
```

Expected: the change moves under `openspec/changes/archive/` and main specs update.

- [ ] **Step 7: Suggested commit**

Suggested commit message:

```bash
feat: hard cut built-in multi-agent execution
```

## Self-Review

- Spec coverage:
  - hard-cut built-in path: Tasks 2, 3, 4, 5
  - remote A2A split semantics: Tasks 2, 7, 8
  - capability-runtime race: Task 6
  - RunLedger audit verdict: Task 7
  - `agent_control` observability: Task 7
  - AGENT.md upgrade path: Task 8
- Placeholder scan:
  - no `TODO`, `TBD`, or "implement later" markers remain
  - each runtime-facing task includes exact files, tests, commands, and expected results
- Type consistency:
  - built-in teammate identity uses `BuiltinTeammateTypes()`
  - runtime race logic refers to `AllowedTools`, `CapabilityRuntime`, `ApplyGrant()`, and `AgentRunConditionNone`
  - docs and OpenSpec artifact names match the change name `dynamic-multi-agent-hard-cut`
