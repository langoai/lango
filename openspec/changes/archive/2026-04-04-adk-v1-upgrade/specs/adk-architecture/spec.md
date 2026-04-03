## MODIFIED Requirements

### Requirement: ADK Agent Abstraction
The system SHALL wrap the Google ADK Agent (v1.0.0) to integrate with the application.

#### Scenario: Agent Initialization
- **WHEN** the server starts
- **THEN** it SHALL initialize an ADK Agent instance
- **AND** configure it with the selected model and tools from the configuration

#### Scenario: ADK dependency version
- **WHEN** the project is built
- **THEN** `go.mod` SHALL declare `google.golang.org/adk v1.0.0` as a direct dependency
- **AND** `go build ./...` SHALL succeed without source code modifications to production files

#### Scenario: MCP spike test type compatibility
- **WHEN** the MCP spike test references `ConfirmationProvider`
- **THEN** it SHALL use `tool.ConfirmationProvider` (not `mcptoolset.ConfirmationProvider`)
- **AND** `go vet ./...` SHALL pass without errors
