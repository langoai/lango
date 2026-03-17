### Requirement: Automation prompt prefix
All prompts originating from cron executor, background task manager, or workflow engine SHALL be prepended with an `[Automated Task]` prefix before being passed to the agent runner.

#### Scenario: Cron job prompt includes automation prefix
- **WHEN** a cron job executes via `buildPromptWithHistory()`
- **THEN** the prompt sent to the agent runner starts with `[Automated Task — Execute the following task using tools. Do NOT answer from general knowledge alone.]`
- **AND** the original user prompt is preceded by `Task: `

#### Scenario: Cron job with history includes automation prefix
- **WHEN** a cron job executes and previous execution history exists
- **THEN** the prompt contains the automation prefix followed by the history section followed by `Task: <original prompt>`

#### Scenario: Background task prompt includes automation prefix
- **WHEN** a background task executes via the manager's `execute()` method
- **THEN** the prompt sent to the agent runner starts with the automation prefix followed by `Task: <original prompt>`

#### Scenario: Workflow step prompt includes automation prefix
- **WHEN** a workflow step executes via the engine's `executeStep()` method
- **THEN** the rendered prompt is wrapped with the automation prefix followed by `Task: <rendered prompt>`

### Requirement: Orchestrator automated task routing rule
The orchestrator instruction SHALL include an "Automated Task Handling" section that overrides the Decision Protocol's ASSESS step for prompts starting with `[Automated Task`.

#### Scenario: Orchestrator delegates automated task
- **WHEN** the orchestrator receives a prompt starting with `[Automated Task`
- **THEN** the orchestrator MUST delegate to a sub-agent based on the task content
- **AND** the orchestrator MUST NOT respond directly

#### Scenario: Orchestrator routes based on task content not scheduling keywords
- **WHEN** an automated task prompt contains "search for latest news"
- **THEN** the orchestrator delegates to the librarian (search capability) not the automator (scheduling keywords)

#### Scenario: Routing rule precedes Decision Protocol
- **WHEN** the orchestrator instruction is assembled
- **THEN** the "Automated Task Handling" section appears before the "Decision Protocol" section
