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

### Requirement: Approval summaries for automation tools
The system SHALL provide human-readable approval summaries for dangerous/moderate automation tools in `buildApprovalSummary()`.

#### Scenario: Approval summary for cron_add
- **WHEN** the approval system generates a summary for `cron_add`
- **THEN** the summary SHALL include the job name, schedule type, and schedule value

#### Scenario: Approval summary for workflow_run
- **WHEN** the approval system generates a summary for `workflow_run` with a file_path
- **THEN** the summary SHALL include the file path
