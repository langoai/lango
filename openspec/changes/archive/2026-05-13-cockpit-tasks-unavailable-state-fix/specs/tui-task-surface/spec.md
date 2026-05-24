## ADDED Requirements

### Requirement: Tasks page distinguishes unavailable from empty task state
The cockpit Tasks page SHALL distinguish between a missing background-task manager and an empty configured task list.

#### Scenario: Nil task manager renders unavailable message
- **WHEN** the Tasks page renders with no configured background-task lister
- **THEN** the page SHALL explain that the background task manager is not configured

#### Scenario: Empty configured task list renders no-tasks message
- **WHEN** the Tasks page renders with a configured task lister that returns zero tasks
- **THEN** the page SHALL display `No active tasks`
