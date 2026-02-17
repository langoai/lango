## Purpose

Define the DAG-based workflow engine that enables declarative YAML workflow definitions with parallel step execution, template variable substitution, and persistent state for resume capability.

## Requirements

### Requirement: YAML workflow definition
The system SHALL parse workflow definitions from YAML files with fields: name, description, schedule (optional), deliver_to (optional), and steps array.

#### Scenario: Parse a valid workflow YAML
- **WHEN** a YAML file is parsed with name, description, and steps
- **THEN** the parser SHALL return a Workflow struct with all fields populated

#### Scenario: Parse workflow with schedule
- **WHEN** a YAML file includes a schedule field with "0 9 * * MON"
- **THEN** the workflow SHALL be eligible for cron integration

### Requirement: Step definition with dependencies
Each workflow step SHALL have: id (unique within workflow), agent, prompt, depends_on (list of step IDs), deliver_to (optional), and timeout (optional).

#### Scenario: Step with dependencies
- **WHEN** a step has depends_on ["fetch-changes"]
- **THEN** the step SHALL not execute until the "fetch-changes" step completes

#### Scenario: Step without dependencies
- **WHEN** a step has no depends_on field
- **THEN** the step SHALL be eligible to run immediately (root step)

### Requirement: DAG validation
The system SHALL validate that step dependencies form a valid DAG (no cycles, all referenced step IDs exist).

#### Scenario: Cycle detection
- **WHEN** a workflow has steps A->B->C->A forming a cycle
- **THEN** the parser SHALL return an error indicating a dependency cycle

#### Scenario: Missing dependency reference
- **WHEN** a step references a depends_on step ID that does not exist
- **THEN** the parser SHALL return an error indicating the unknown dependency

### Requirement: Topological sort with parallel layers
The DAG engine SHALL produce a topological ordering as parallel execution layers, where each layer contains independent steps that can run concurrently.

#### Scenario: Parallel execution layers
- **WHEN** a workflow has steps: A (no deps), B->A, C->A, D->[B,C]
- **THEN** the DAG SHALL produce layers: [A], [B, C], [D]

### Requirement: Template variable substitution
Step prompts SHALL support `{{step-id.result}}` syntax for referencing results of completed dependency steps using Go text/template.

#### Scenario: Variable substitution
- **WHEN** a step prompt contains "Analyze: {{fetch-changes.result}}"
- **THEN** the engine SHALL replace the template variable with the actual result of the "fetch-changes" step

### Requirement: Concurrent step execution
The engine SHALL execute steps within the same layer concurrently, limited by the configured maxConcurrentSteps semaphore.

#### Scenario: Concurrent execution within a layer
- **WHEN** a layer has 3 independent steps and maxConcurrentSteps is 4
- **THEN** all 3 steps SHALL execute concurrently

### Requirement: Ent-backed state persistence
Workflow runs and step runs SHALL be persisted in Ent ORM (WorkflowRun, WorkflowStepRun) for resume capability.

#### Scenario: State persistence on step completion
- **WHEN** a workflow step completes
- **THEN** the step result and status SHALL be saved to the WorkflowStepRun table

#### Scenario: Resume interrupted workflow
- **WHEN** Resume() is called with a run ID of a partially completed workflow
- **THEN** the engine SHALL load the saved state, skip completed steps, and continue from the last incomplete step

### Requirement: Workflow lifecycle operations
The system SHALL support Run, Resume, Cancel, Status, and ListRuns operations.

#### Scenario: Run a workflow
- **WHEN** Run() is called with a valid workflow
- **THEN** the engine SHALL create a WorkflowRun, execute steps in DAG order, and return the final result

#### Scenario: Cancel a running workflow
- **WHEN** Cancel() is called on a running workflow
- **THEN** the engine SHALL cancel the context and mark remaining steps as skipped

#### Scenario: Get workflow status
- **WHEN** Status() is called with a run ID
- **THEN** the engine SHALL return the current status of the workflow run including all step statuses

### Requirement: Step-level delivery
Individual steps MAY specify deliver_to for intermediate result delivery to channels.

#### Scenario: Step with delivery target
- **WHEN** a step completes and has deliver_to ["slack"]
- **THEN** the step result SHALL be delivered to Slack in addition to being stored

### Requirement: Workflow-level delivery
The workflow MAY specify deliver_to for final result delivery when all steps complete.

#### Scenario: Workflow completion delivery
- **WHEN** all workflow steps complete and the workflow has deliver_to ["slack"]
- **THEN** the combined final result SHALL be delivered to Slack

### Requirement: Engine shutdown
The `Engine` SHALL provide a `Shutdown()` method that cancels all running workflow executions via the cancels map.

#### Scenario: Graceful shutdown
- **WHEN** `Shutdown()` is called
- **THEN** all cancel functions in the cancels map SHALL be invoked

### Requirement: RunStatus StartedAt field
The `RunStatus` struct SHALL include a `StartedAt time.Time` field populated from the workflow run record.

#### Scenario: StartedAt in GetRunStatus
- **WHEN** `GetRunStatus()` is called
- **THEN** the returned `RunStatus` SHALL include the `StartedAt` timestamp from the database record

#### Scenario: StartedAt in ListRuns
- **WHEN** `ListRuns()` is called
- **THEN** each returned `RunStatus` SHALL include the `StartedAt` timestamp
