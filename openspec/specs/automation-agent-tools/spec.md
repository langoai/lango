## Purpose

Define the agent-accessible tools for cron scheduling, background task management, and workflow orchestration that enable conversational automation through the AI agent.

## Requirements

### Requirement: Cron agent tools
The system SHALL provide 6 agent tools for cron job management: `cron_add` (Moderate), `cron_list` (Safe), `cron_pause` (Moderate), `cron_resume` (Moderate), `cron_remove` (Dangerous), `cron_history` (Safe).

#### Scenario: Create a cron job via conversation
- **WHEN** the agent calls `cron_add` with name, schedule_type, schedule, and prompt parameters
- **THEN** the system SHALL create and register a new cron job and return confirmation with the job name and schedule

#### Scenario: List cron jobs
- **WHEN** the agent calls `cron_list`
- **THEN** the system SHALL return all registered cron jobs with their IDs, names, schedules, and enabled status

#### Scenario: Pause and resume a cron job
- **WHEN** the agent calls `cron_pause` with a job ID
- **THEN** the system SHALL disable the job and unregister it from the scheduler

#### Scenario: Remove a cron job permanently
- **WHEN** the agent calls `cron_remove` with a job ID
- **THEN** the system SHALL delete the job from the store and unregister it from the scheduler

#### Scenario: View cron execution history
- **WHEN** the agent calls `cron_history` with optional job_id and limit parameters
- **THEN** the system SHALL return execution history entries filtered by job ID if provided

### Requirement: Background task agent tools
The system SHALL provide 5 agent tools for background task management: `bg_submit` (Moderate), `bg_status` (Safe), `bg_list` (Safe), `bg_result` (Safe), `bg_cancel` (Moderate).

#### Scenario: Submit a background task
- **WHEN** the agent calls `bg_submit` with a prompt and optional channel
- **THEN** the system SHALL create a background task and return the task ID immediately

#### Scenario: Check task status
- **WHEN** the agent calls `bg_status` with a task_id
- **THEN** the system SHALL return the current task snapshot including status, prompt, and timing

#### Scenario: Retrieve completed result
- **WHEN** the agent calls `bg_result` with a task_id for a completed task
- **THEN** the system SHALL return the task result text

#### Scenario: Cancel a running task
- **WHEN** the agent calls `bg_cancel` with a task_id
- **THEN** the system SHALL cancel the task and return confirmation

### Requirement: Workflow agent tools
The system SHALL provide 5 agent tools for workflow management: `workflow_run` (Moderate), `workflow_status` (Safe), `workflow_list` (Safe), `workflow_cancel` (Moderate), `workflow_save` (Moderate).

#### Scenario: Run a workflow from file path
- **WHEN** the agent calls `workflow_run` with a file_path parameter
- **THEN** the system SHALL parse, validate, and execute the workflow, returning run_id, status, and step results

#### Scenario: Run a workflow from inline YAML
- **WHEN** the agent calls `workflow_run` with yaml_content parameter
- **THEN** the system SHALL parse the inline YAML and execute it

#### Scenario: Save a workflow definition
- **WHEN** the agent calls `workflow_save` with name and yaml_content
- **THEN** the system SHALL validate the YAML and save it to the workflows directory as `<name>.flow.yaml`

#### Scenario: Cancel a running workflow
- **WHEN** the agent calls `workflow_cancel` with a run_id
- **THEN** the system SHALL cancel the running workflow and update its status

### Requirement: Shared automation helper package
The automation subsystem SHALL provide a shared `internal/automation/` helper package containing `AgentRunner`, `ChannelSender`, and `DetectChannelFromContext(ctx)` so cron, background, and workflow packages use one common contract.

#### Scenario: Automation packages share one runner interface
- **WHEN** `cron`, `background`, and `workflow` need to execute agent prompts
- **THEN** they SHALL depend on `automation.AgentRunner`
- **AND** they SHALL NOT define package-local duplicate runner interfaces

#### Scenario: Automation packages share one sender interface
- **WHEN** automation packages need channel delivery contracts
- **THEN** they SHALL depend on `automation.ChannelSender`
- **AND** they SHALL NOT define package-local duplicate sender interfaces

#### Scenario: Delivery target is derived from session context
- **WHEN** cron, background, or workflow tools omit an explicit delivery target
- **THEN** they SHALL use `automation.DetectChannelFromContext(ctx)` to derive `channel:target`
- **AND** they SHALL fall back to configured defaults only when no context-derived target exists

### Requirement: Approval summaries for automation tools
The system SHALL provide human-readable approval summaries for dangerous/moderate automation tools in `buildApprovalSummary()`.

#### Scenario: Approval summary for cron_add
- **WHEN** the approval system generates a summary for `cron_add`
- **THEN** the summary SHALL include the job name, schedule type, and schedule value

#### Scenario: Approval summary for workflow_run
- **WHEN** the approval system generates a summary for `workflow_run` with a file_path
- **THEN** the summary SHALL include the file path

### Requirement: Automation prompt sections mention auto-detection
The automation prompt section SHALL inform the agent that delivery channel parameters are optional and will be auto-detected from the current session context when omitted.

#### Scenario: Cron prompt
- **WHEN** the automation prompt section is built with cron enabled
- **THEN** the cron section SHALL include a note that deliver_to is optional and auto-detected

#### Scenario: Background prompt
- **WHEN** the automation prompt section is built with background enabled
- **THEN** the background section SHALL include a note that channel is optional and auto-detected

#### Scenario: Workflow prompt
- **WHEN** the automation prompt section is built with workflow enabled
- **THEN** the workflow section SHALL include a note that deliver_to in YAML is optional and auto-detected

### Requirement: Exec prohibition in automation prompt
The automation prompt section SHALL include an explicit instruction prohibiting the use of exec to run ANY lango CLI command, not only automation subcommands. The prohibition SHALL list all known subcommands (cron, bg, workflow, graph, memory, p2p, security, payment, config, doctor, and others) and explain that every lango CLI invocation requires passphrase authentication during bootstrap and will fail in non-interactive subprocess contexts.

#### Scenario: Prompt includes comprehensive exec prohibition
- **WHEN** any automation feature (cron, background, or workflow) is enabled
- **THEN** the automation prompt section SHALL contain text instructing the agent to NEVER use exec to run ANY "lango" CLI command, covering all subcommands including but not limited to cron, bg, workflow, graph, memory, p2p, security, payment, config, and doctor
- **AND** the prohibition SHALL explain that spawning a new lango process requires passphrase authentication and will fail in non-interactive mode
- **AND** the prohibition SHALL instruct the agent to ask the user to run commands directly in their terminal when no built-in tool equivalent exists

### Requirement: Comprehensive CLI exec guard
The `blockLangoExec()` function SHALL block ALL `lango` CLI invocations attempted through `exec` or `exec_bg` tools, using a two-phase approach: (1) specific subcommand guards with per-command tool alternative messages, and (2) a catch-all guard for any remaining `lango` prefix.

#### Scenario: Block subcommand with in-process equivalent
- **WHEN** the agent attempts to exec a `lango` subcommand that has in-process tool equivalents (graph, memory, p2p, security, payment, cron, bg, workflow)
- **THEN** the system SHALL return a blocked message listing the specific built-in tools to use instead

#### Scenario: Block subcommand without in-process equivalent
- **WHEN** the agent attempts to exec a `lango` subcommand that has no in-process equivalent (config, doctor, settings, serve, onboard, agent)
- **THEN** the system SHALL return a blocked message explaining that passphrase authentication is required and the user should run the command directly in their terminal

#### Scenario: Allow non-lango commands
- **WHEN** the agent attempts to exec a command that does not start with `lango ` or equal `lango`
- **THEN** the system SHALL allow the command to proceed (return empty string)

#### Scenario: Case-insensitive matching
- **WHEN** the agent attempts to exec a lango command in any case (e.g., `LANGO SECURITY DB-MIGRATE`)
- **THEN** the system SHALL still block and return the appropriate guidance message

### Requirement: Exec tool prompt safety rules
The `TOOL_USAGE.md` prompt SHALL include an explicit top-level rule under the Exec Tool section warning against using exec to run any `lango` CLI command. The rule SHALL list specific subcommands as examples and explain the passphrase failure mechanism. The rule SHALL also instruct the agent to inform the user and ask them to run commands directly when no built-in tool equivalent exists.

#### Scenario: TOOL_USAGE.md contains exec safety rule
- **WHEN** the agent's tool usage prompt is loaded
- **THEN** the first bullet point under "### Exec Tool" SHALL warn against running any lango CLI command via exec
