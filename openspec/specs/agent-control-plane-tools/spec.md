## Purpose

Agent Control Plane Tools provide the tool-level interface for agent lifecycle management and structured task tracking within the agent runtime. These tools allow agents to spawn child agents, wait for their completion, stop them, and manage structured tasks — forming the operational surface of the agent control plane.

## Requirements

### Requirement: agent_spawn tool creates AgentRun with enriched prompt and advisory routing
The `agent_spawn` tool SHALL create an `AgentRun` entry in the `AgentRunStore` with a randomly generated ID (prefixed `arun-`), the caller's instruction, and optional advisory routing. When a `RequestedAgent` is specified, the instruction SHALL be enriched with a system prefix indicating the preferred specialist. The tool SHALL accept an optional `allowed_tools` parameter that populates `AgentRun.AllowedTools` for DynamicAllowedTools enforcement. The tool's SafetyLevel SHALL be `SafetyLevelModerate`.

#### Scenario: Spawn with plain instruction
- **WHEN** `agent_spawn` is called with `instruction: "Summarize the document"`
- **AND** no `agent` parameter is provided
- **THEN** an `AgentRun` SHALL be created with status `spawned`
- **AND** `Instruction` SHALL equal the original instruction without enrichment
- **AND** `RequestedAgent` SHALL be empty
- **AND** the response SHALL include `agent_id`, `status: "spawned"`, and `requested_agent: ""`

#### Scenario: Spawn with advisory agent routing
- **WHEN** `agent_spawn` is called with `instruction: "Analyze security"` and `agent: "security-specialist"`
- **THEN** `Instruction` SHALL be enriched to `"[System: This task is best handled by the 'security-specialist' specialist.]\n\nAnalyze security"`
- **AND** `RequestedAgent` SHALL be `"security-specialist"`

#### Scenario: Spawn with allowed_tools restriction
- **WHEN** `agent_spawn` is called with `allowed_tools: ["fs_read", "search_knowledge"]`
- **THEN** `AgentRun.AllowedTools` SHALL be `["fs_read", "search_knowledge"]`

#### Scenario: Spawn registers pending ID with projection
- **WHEN** `agent_spawn` succeeds and a `Projection` is configured on the control plane
- **THEN** `Projection.RegisterPending(agentID)` SHALL be called
- **AND** the registered ID SHALL be returned by the next `PrepareTask` call from the background manager

#### Scenario: Spawn increments depth from parent context
- **WHEN** the parent context has `SpawnDepth` of 1
- **THEN** the created `AgentRun.SpawnDepth` SHALL be 2

### Requirement: agent_wait polls AgentRunStore until terminal status
The `agent_wait` tool SHALL poll the `AgentRunStore` at 500ms intervals until the target `AgentRun` reaches a terminal status (`completed`, `failed`, or `cancelled`) or a timeout is reached. The default timeout SHALL be 300 seconds. The tool's SafetyLevel SHALL be `SafetyLevelSafe`.

#### Scenario: Wait for completed agent
- **GIVEN** an `AgentRun` with ID `arun-abc123` in status `running`
- **WHEN** `agent_wait` is called with `agent_id: "arun-abc123"`
- **AND** the run transitions to `completed` with `result: "Analysis complete"`
- **THEN** the response SHALL include `agent_id: "arun-abc123"`, `status: "completed"`, and `result: "Analysis complete"`

#### Scenario: Wait times out
- **GIVEN** an `AgentRun` that remains in status `running`
- **WHEN** `agent_wait` is called with `agent_id: "arun-abc123"` and `timeout: 5`
- **AND** the run does not reach a terminal status within 5 seconds
- **THEN** the response SHALL include `timeout: true` and the current non-terminal status

#### Scenario: Wait respects context cancellation
- **WHEN** `agent_wait` is polling and the parent context is cancelled
- **THEN** the tool SHALL return a context error immediately

#### Scenario: Wait for non-existent agent
- **WHEN** `agent_wait` is called with an `agent_id` that does not exist in the store
- **THEN** the tool SHALL return an error

### Requirement: agent_stop cancels via AgentRunStore.Cancel
The `agent_stop` tool SHALL cancel a spawned agent by invoking `AgentRunStore.Cancel(agentID)`, which sets the status to `cancelled`, records `CompletedAt`, and calls the run's `CancelFn` if set. The tool's SafetyLevel SHALL be `SafetyLevelSafe`.

#### Scenario: Stop a running agent
- **GIVEN** an `AgentRun` with ID `arun-abc123` in status `running`
- **WHEN** `agent_stop` is called with `agent_id: "arun-abc123"`
- **THEN** the run's status SHALL become `cancelled`
- **AND** the response SHALL include `agent_id: "arun-abc123"` and `status: "cancelled"`

#### Scenario: Stop an already terminal agent
- **GIVEN** an `AgentRun` with ID `arun-abc123` in status `completed`
- **WHEN** `agent_stop` is called with `agent_id: "arun-abc123"`
- **THEN** the tool SHALL return an error indicating the run is already terminal

### Requirement: agent_message excluded from initial tool set
The `agent_message` tool SHALL NOT be included in the initial `BuildControlTools` output. It is deferred to a future change for inter-agent messaging.

#### Scenario: BuildControlTools returns exactly three tools
- **WHEN** `BuildControlTools` is called
- **THEN** it SHALL return exactly `[agent_spawn, agent_wait, agent_stop]`
- **AND** no `agent_message` tool SHALL be present

### Requirement: Task management tools provide CRUD on TaskEntry
The system SHALL provide four task management tools — `task_create`, `task_get`, `task_list`, `task_update` — backed by a `TaskStore` interface. All four tools SHALL have `SafetyLevel` set to `SafetyLevelSafe`.

#### Scenario: Create a task
- **WHEN** `task_create` is called with `title: "Implement feature X"`
- **THEN** a `TaskEntry` SHALL be created with a random ID (prefixed `task-`), status `"todo"`, and the given title
- **AND** the response SHALL include the generated `task_id` and `status: "todo"`

#### Scenario: Create a hierarchical task
- **WHEN** `task_create` is called with `title: "Sub-task A"` and `parent_id: "task-parent123"`
- **THEN** the created `TaskEntry.ParentID` SHALL be `"task-parent123"`

#### Scenario: Get a task by ID
- **GIVEN** a `TaskEntry` with ID `task-abc123`
- **WHEN** `task_get` is called with `task_id: "task-abc123"`
- **THEN** the response SHALL include all fields: `id`, `title`, `status`, `agent_id`, `parent_id`, `description`, `created_at`, `updated_at`

#### Scenario: Get a non-existent task
- **WHEN** `task_get` is called with a `task_id` that does not exist
- **THEN** the tool SHALL return an error

#### Scenario: List tasks with status filter
- **GIVEN** tasks with statuses `"todo"`, `"in_progress"`, and `"done"`
- **WHEN** `task_list` is called with `status: "todo"`
- **THEN** the response SHALL include only tasks with status `"todo"`

#### Scenario: List tasks with parent filter
- **WHEN** `task_list` is called with `parent_id: "task-parent123"`
- **THEN** the response SHALL include only tasks whose `ParentID` matches

#### Scenario: List all tasks without filters
- **WHEN** `task_list` is called with no filters
- **THEN** the response SHALL include all tasks with a `count` field

#### Scenario: Update task status
- **GIVEN** a `TaskEntry` with status `"todo"`
- **WHEN** `task_update` is called with `task_id` and `status: "in_progress"`
- **THEN** the task's status SHALL become `"in_progress"` and `UpdatedAt` SHALL be refreshed

#### Scenario: Update task description only
- **WHEN** `task_update` is called with `task_id` and `description: "Updated details"` but no `status`
- **THEN** the task's description SHALL change and its status SHALL remain unchanged

### Requirement: AgentRunProjection implements background.Projection for ID unification
`AgentRunProjection` SHALL implement the `background.Projection` interface to synchronize the `AgentRun` lifecycle with the background manager's task lifecycle. `RegisterPending` pre-assigns an `AgentRun.ID` so that `PrepareTask` returns it to the background manager, ensuring both layers share the same canonical ID. `SyncTask` SHALL map background task status transitions to `AgentRun` status via a defined mapping.

#### Scenario: PrepareTask returns pre-registered ID
- **GIVEN** `RegisterPending("arun-abc123")` was called
- **WHEN** `PrepareTask` is invoked by the background manager
- **THEN** it SHALL return `"arun-abc123"`
- **AND** the pending registration SHALL be consumed (one-time use)

#### Scenario: PrepareTask with no pending ID
- **WHEN** `PrepareTask` is invoked and no ID was registered via `RegisterPending`
- **THEN** it SHALL return an error

#### Scenario: SyncTask maps background status to AgentRun status
- **WHEN** `SyncTask` receives a `TaskSnapshot` with status `Running`
- **THEN** it SHALL call `AgentRunStore.UpdateStatus` with `AgentRunRunning`

#### Scenario: Status mapping — Pending to Spawned
- **WHEN** background status is `Pending`
- **THEN** AgentRun status SHALL be `spawned`

#### Scenario: Status mapping — Done to Completed
- **WHEN** background status is `Done`
- **THEN** AgentRun status SHALL be `completed` with the snapshot result

#### Scenario: Status mapping — Failed to Failed
- **WHEN** background status is `Failed`
- **THEN** AgentRun status SHALL be `failed` with the snapshot error

#### Scenario: Status mapping — Cancelled to Cancelled
- **WHEN** background status is `Cancelled`
- **THEN** AgentRun status SHALL be `cancelled`

### Requirement: DynamicAllowedTools enforcement via context key and access control hook
The system SHALL enforce per-agent tool restrictions at runtime using a `DynamicAllowedTools` context key. The `AgentAccessControlHook.Pre()` method SHALL check for `DynamicAllowedToolsFromContext(ctx)` and, when a non-empty allowlist is present, block any tool not in the allowlist — except for runtime essentials. Runtime essentials (`tool_output_get`, `builtin_list`, `builtin_search`, `builtin_health`) SHALL always be allowed. `builtin_invoke` SHALL be excluded from runtime essentials because it can proxy-execute other tools, bypassing the allowlist.

#### Scenario: Tool in DynamicAllowedTools passes
- **GIVEN** DynamicAllowedTools context contains `["fs_read", "search_knowledge"]`
- **WHEN** `AgentAccessControlHook.Pre()` is called for tool `fs_read`
- **THEN** the result action SHALL be `Continue`

#### Scenario: Tool not in DynamicAllowedTools is blocked
- **GIVEN** DynamicAllowedTools context contains `["fs_read", "search_knowledge"]`
- **WHEN** `AgentAccessControlHook.Pre()` is called for tool `exec`
- **THEN** the result action SHALL be `Block` with reason `"tool restricted by DynamicAllowedTools"`

#### Scenario: Runtime essential passes regardless of allowlist
- **GIVEN** DynamicAllowedTools context contains `["fs_read"]`
- **WHEN** `AgentAccessControlHook.Pre()` is called for tool `builtin_list`
- **THEN** the result action SHALL be `Continue`

#### Scenario: builtin_invoke is NOT a runtime essential
- **GIVEN** DynamicAllowedTools context contains `["fs_read"]`
- **WHEN** `AgentAccessControlHook.Pre()` is called for tool `builtin_invoke`
- **THEN** the result action SHALL be `Block`

#### Scenario: No DynamicAllowedTools — all tools allowed
- **GIVEN** no DynamicAllowedTools context key is set
- **WHEN** `AgentAccessControlHook.Pre()` is called for any tool
- **THEN** the DynamicAllowedTools check SHALL not block (other ACL checks still apply)

#### Scenario: Deny list takes precedence over DynamicAllowedTools
- **GIVEN** DynamicAllowedTools context contains `["exec"]`
- **AND** the agent's DeniedTools includes `"exec"`
- **WHEN** `AgentAccessControlHook.Pre()` is called for tool `exec`
- **THEN** the result action SHALL be `Block` (deny list checked first)

### Requirement: RecursionGuard enforces spawn depth, self-spawn, and cycle detection
`RecursionGuard` SHALL prevent runaway agent spawn recursion by checking three conditions before any spawn: (1) depth limit — `SpawnDepth` from context must be less than `MaxDepth` (default 3); (2) self-spawn prevention — spawner must not equal target; (3) cycle detection — target must not already appear in the `SpawnChain` from context.

#### Scenario: Spawn within depth limit
- **GIVEN** `MaxDepth` is 3 and `SpawnDepth` from context is 1
- **WHEN** `RecursionGuard.Check(ctx, "agent-a", "agent-b")` is called
- **THEN** it SHALL return nil (allowed)

#### Scenario: Spawn exceeds depth limit
- **GIVEN** `MaxDepth` is 3 and `SpawnDepth` from context is 3
- **WHEN** `RecursionGuard.Check(ctx, "agent-a", "agent-b")` is called
- **THEN** it SHALL return an error containing `"spawn depth 3 exceeds max 3"`

#### Scenario: Self-spawn blocked
- **WHEN** `RecursionGuard.Check(ctx, "agent-a", "agent-a")` is called
- **THEN** it SHALL return an error containing `"self-spawn blocked"`

#### Scenario: Cycle detected in spawn chain
- **GIVEN** `SpawnChain` from context is `["agent-a", "agent-b"]`
- **WHEN** `RecursionGuard.Check(ctx, "agent-c", "agent-a")` is called
- **THEN** it SHALL return an error containing `"cycle detected"` and the chain

#### Scenario: Default MaxDepth is 3
- **WHEN** `NewRecursionGuard(0)` is called
- **THEN** `MaxDepth` SHALL be set to 3

#### Scenario: Empty spawner bypasses self-spawn check
- **WHEN** `RecursionGuard.Check(ctx, "", "agent-a")` is called with depth within limit and no cycle
- **THEN** it SHALL return nil (self-spawn check is skipped for empty spawner)

### Requirement: RequestedAgent routing is advisory via enriched prompt
The `RequestedAgent` field on `AgentRun` SHALL be advisory only — it influences routing through an enriched system prompt prefix, not through code-level enforcement. The supervisor or orchestrator is free to route to any available agent regardless of the `RequestedAgent` value.

#### Scenario: Enriched prompt contains advisory routing hint
- **WHEN** `agent_spawn` is called with `agent: "researcher"`
- **THEN** the stored `Instruction` SHALL contain the prefix `"[System: This task is best handled by the 'researcher' specialist.]"`
- **AND** no code-level enforcement SHALL guarantee routing to `"researcher"`

#### Scenario: No advisory routing without agent parameter
- **WHEN** `agent_spawn` is called without an `agent` parameter
- **THEN** the stored `Instruction` SHALL equal the raw instruction without any system prefix

### Requirement: AgentRunStore provides lifecycle management with terminal status guards
`AgentRunStore` SHALL provide `Create`, `Get`, `List`, `UpdateStatus`, and `Cancel` operations. `UpdateStatus` and `Cancel` SHALL reject updates to runs that are already in a terminal status (`completed`, `failed`, `cancelled`). `Cancel` SHALL invoke the run's `CancelFn` if set. `Get` SHALL return a copy of the run with `CancelFn` deliberately set to nil to prevent external cancellation through snapshots.

#### Scenario: Create and retrieve an agent run
- **WHEN** `Create` is called with a valid `AgentRun`
- **THEN** `Get` SHALL return a copy with matching fields

#### Scenario: Create duplicate ID rejected
- **GIVEN** an `AgentRun` with ID `"arun-abc"` already exists
- **WHEN** `Create` is called with the same ID
- **THEN** it SHALL return an error containing `"already exists"`

#### Scenario: UpdateStatus on terminal run rejected
- **GIVEN** an `AgentRun` in status `completed`
- **WHEN** `UpdateStatus` is called
- **THEN** it SHALL return an error indicating the run is already terminal

#### Scenario: Cancel sets CompletedAt and invokes CancelFn
- **GIVEN** an `AgentRun` in status `running` with a `CancelFn` set
- **WHEN** `Cancel` is called
- **THEN** `Status` SHALL be `cancelled`, `CompletedAt` SHALL be set, and `CancelFn` SHALL be invoked

#### Scenario: Get returns copy without CancelFn
- **GIVEN** an `AgentRun` with a `CancelFn` set
- **WHEN** `Get` is called
- **THEN** the returned copy SHALL have `CancelFn` set to nil
