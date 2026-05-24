## Why

The `agentrt` control-plane and task-management tools already validate required wrapper inputs with `toolparam.RequireString`, but the regression surface does not pin those missing-parameter messages precisely. That leaves room for future drift where the tool boundary could weaken back into generic downstream errors without immediate detection.

## What Changes

- Add exact wrapper-level regressions for missing required inputs on `agent_spawn`, `agent_wait`, `agent_stop`, `task_create`, `task_get`, and `task_update`.
- Update agent-facing prompt guidance and multi-agent docs to state those required-input contracts explicitly.
- Sync the control-plane and production-readiness specs to the same fail-closed wrapper contract.

## Impact

- `agent-control-plane-tools`: required inputs stay fail-closed at the wrapper boundary.
- `multi-agent-orchestration`: task/control-plane operator surfaces become more explicit and easier to reason about.
- `production-readiness`: exact wrapper error semantics remain covered by regression tests.
