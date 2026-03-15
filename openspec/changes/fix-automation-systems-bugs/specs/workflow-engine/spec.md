## ADDED Requirements

### Requirement: Run-scoped session keys
The workflow engine SHALL include the `runID` in step session keys to isolate sessions across re-runs of the same workflow. The session key format SHALL be `workflow:{workflowName}:{runID}:{stepID}`.

#### Scenario: Different runs produce different session keys
- **WHEN** the same workflow is run twice producing runID "run-1" and "run-2"
- **THEN** step "step-a" SHALL use session keys `workflow:wf:run-1:step-a` and `workflow:wf:run-2:step-a` respectively

#### Scenario: Session isolation prevents result contamination
- **WHEN** a workflow is re-run after a previous completion
- **THEN** each step SHALL execute in a fresh session without access to previous run's conversation history

### Requirement: Step-level cancellation checks
The workflow engine SHALL check for context cancellation at two points during DAG execution: (1) at the beginning of `executeStep()` before any work, and (2) in the `runDAG()` goroutine after acquiring the concurrency semaphore but before calling `executeStep()`.

#### Scenario: Cancelled before step starts
- **WHEN** the workflow context is cancelled before `executeStep()` is called
- **THEN** `executeStep()` SHALL return `ctx.Err()` immediately without running the agent

#### Scenario: Cancelled after semaphore acquisition
- **WHEN** the workflow context is cancelled while a step goroutine is waiting for the semaphore, and then acquires the semaphore
- **THEN** the goroutine SHALL check `ctx.Err()`, mark the step as cancelled, and return without calling `executeStep()`

#### Scenario: Cancellation prevents pending steps
- **WHEN** a workflow with 3 layers is cancelled during layer 2 execution
- **THEN** layer 3 steps SHALL NOT start execution
