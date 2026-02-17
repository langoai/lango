## Context

Cron scheduling, background task execution, and workflow engine are fully implemented at the infrastructure level (`internal/cron/`, `internal/background/`, `internal/workflow/`) and exposed via CLI commands. However, the AI agent has no tools to interact with these systems during conversation. Users must switch to CLI to schedule jobs or submit background tasks, breaking the conversational flow.

Additionally, the `BackgroundManager` and `WorkflowEngine` lack `Shutdown()` methods, meaning `App.Stop()` does not gracefully terminate running background tasks or workflows. The initialization order in `App.New()` places cron/bg/workflow after agent creation, which means their tools cannot be approval-wrapped.

## Goals / Non-Goals

**Goals:**
- Expose all cron/bg/workflow operations as agent tools so users can manage automation via conversation
- Route automation tools to a dedicated "automator" sub-agent in multi-agent mode
- Add proper shutdown lifecycle for BackgroundManager and WorkflowEngine
- Add start notifications so users know when async jobs begin execution
- Guide the agent via system prompt to connect natural language requests to appropriate tools

**Non-Goals:**
- Building a workflow visual editor or GUI
- Adding new cron/bg/workflow features beyond tool exposure
- Implementing tool-level rate limiting (relies on existing approval policy)
- Adding persistent storage for background tasks (remains in-memory)

## Decisions

### 1. Tool registration before approval wrapping
**Decision**: Move cron/bg/workflow initialization from after agent creation to before approval wrapping (step 5j-l in App.New).
**Rationale**: Tools must be in the `tools` slice before the approval wrapping loop (step 8) to get safety-level-based approval enforcement. The `agentRunnerAdapter` is lazy (calls `app.runAgent()` at execution time), so initializing it before the agent exists is safe — actual job execution only happens after `App.Start()`.

### 2. Dedicated "automator" sub-agent
**Decision**: Add a new "automator" AgentSpec rather than assigning automation tools to the existing "operator" agent.
**Rationale**: The operator handles shell/file/skill operations. Mixing scheduling and file I/O in one agent creates confusion in the routing table. A dedicated automator provides clearer keyword matching for scheduling-related requests and better rejection boundaries.

### 3. Safety levels for automation tools
**Decision**: `cron_remove` = Dangerous, `cron_add/pause/resume` = Moderate, `cron_list/history` = Safe. Same pattern for bg/workflow tools.
**Rationale**: Permanent deletion requires explicit approval. Creation and state changes need moderate approval. Read-only operations are always safe.

### 4. Workflow save location
**Decision**: Workflow YAML files saved to `cfg.Workflow.StateDir` (default `~/.lango/workflows/`).
**Rationale**: Reuses the existing config field. Keeps workflow definitions alongside other user data under `~/.lango/`.

## Risks / Trade-offs

- **[Risk] Background tasks are in-memory only** → If the process crashes, running tasks are lost. Mitigation: This is the existing design; no regression. Background tasks are for short-lived async operations.
- **[Risk] Workflow tools block during execution** → `workflow_run` blocks until the workflow completes. Mitigation: Users can combine with `bg_submit` for truly async workflow execution, or use workflow tools that have per-step timeouts.
- **[Trade-off] 16 tools added to context** → More tools increase prompt size. Mitigation: Tools are only added when their respective config flags are enabled (`cron.enabled`, `background.enabled`, `workflow.enabled`).
