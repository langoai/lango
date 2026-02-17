## MODIFIED Requirements

### Requirement: README documents all 7 sub-agents
The README.md SHALL list all 7 sub-agents (operator, navigator, vault, librarian, automator, planner, chronicler) in every location that references the sub-agent roster.

#### Scenario: Feature list includes automator
- **WHEN** the README Features section is read
- **THEN** the Multi-Agent Orchestration bullet SHALL list automator alongside the other 6 sub-agents

#### Scenario: Directory structure includes automator
- **WHEN** the README Architecture directory tree is read
- **THEN** the orchestration directory comment SHALL list automator alongside the other 6 sub-agents

#### Scenario: Per-agent prompt docs include automator
- **WHEN** the README Per-Agent Prompt Customization section is read
- **THEN** automator SHALL be listed in the sub-agent enumeration

#### Scenario: Orchestration table includes automator row
- **WHEN** the README Multi-Agent Orchestration table is read
- **THEN** it SHALL contain a row for automator with role "Automation: cron scheduling, background tasks, workflow pipelines" and tools "cron_*, bg_*, workflow_*"

#### Scenario: Workflow supported agents include automator
- **WHEN** the README Workflow Engine supported agents text is read
- **THEN** automator SHALL be listed among the supported agent names
