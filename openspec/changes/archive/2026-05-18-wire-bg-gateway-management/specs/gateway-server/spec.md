## ADDED Requirements

### Requirement: Background management HTTP routes
The gateway SHALL expose authenticated REST endpoints for managing the running process's in-memory background tasks.

#### Scenario: Background task list endpoint
- **WHEN** an authenticated client calls `GET /api/bg/tasks`
- **THEN** the gateway SHALL return JSON containing the current in-memory task list
- **AND** each task SHALL expose a stable string status instead of the internal status enum integer

#### Scenario: Background task status endpoint
- **WHEN** an authenticated client calls `GET /api/bg/tasks/{id}`
- **THEN** the gateway SHALL return JSON containing the matching task details
- **AND** it SHALL return `404` when the task is not found

#### Scenario: Background task result endpoint
- **WHEN** an authenticated client calls `GET /api/bg/tasks/{id}/result`
- **THEN** the gateway SHALL return JSON containing the completed task result
- **AND** it SHALL return a non-2xx error when the task is not done or not found

#### Scenario: Background task cancel endpoint
- **WHEN** an authenticated client calls `POST /api/bg/tasks/{id}/cancel`
- **THEN** the gateway SHALL request cancellation through the running process's background manager
- **AND** it SHALL return JSON confirming the cancelled task id

#### Scenario: Background manager unavailable
- **WHEN** background automation is disabled or no manager is configured
- **THEN** all `/api/bg/*` management endpoints SHALL return `503`

#### Scenario: Background routes honor gateway auth
- **WHEN** gateway auth is configured
- **THEN** `/api/bg/*` routes SHALL require the existing gateway session authentication middleware
