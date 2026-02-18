## ADDED Requirements

### Requirement: README documents streaming in gateway feature

README.md Features list SHALL describe the Gateway as supporting real-time streaming.

#### Scenario: Gateway feature line updated
- **WHEN** a user reads the Features list in README.md
- **THEN** the Gateway bullet reads "WebSocket/HTTP server with real-time streaming"

### Requirement: README documents observational memory context limit configs

README.md Configuration Reference table SHALL include `observationalMemory.maxReflectionsInContext` and `observationalMemory.maxObservationsInContext` fields matching `ObservationalMemoryConfig` in `internal/config/types.go`.

#### Scenario: Context limit config fields present
- **WHEN** a user reads the Observational Memory config section in README.md
- **THEN** the table contains `observationalMemory.maxReflectionsInContext` (int, default `5`) and `observationalMemory.maxObservationsInContext` (int, default `20`)

### Requirement: README documents embedding cache

README.md Embedding & RAG section SHALL include an Embedding Cache subsection describing in-memory caching with 5-minute TTL and 100-entry limit.

#### Scenario: Embedding cache subsection present
- **WHEN** a user reads the Embedding & RAG section in README.md
- **THEN** there is an "Embedding Cache" heading describing automatic in-memory caching with 5-minute TTL and 100-entry limit

### Requirement: README documents observational memory context limits

README.md Observational Memory section SHALL describe context limits for reflections and observations.

#### Scenario: Context limits bullet present
- **WHEN** a user reads the Observational Memory component list in README.md
- **THEN** there is a "Context Limits" bullet describing default limits of 5 reflections and 20 observations

### Requirement: README documents WebSocket events

README.md SHALL include a WebSocket Events subsection documenting `agent.thinking`, `agent.chunk`, and `agent.done` events with their payloads.

#### Scenario: WebSocket events table present
- **WHEN** a user reads the WebSocket section in README.md
- **THEN** there is a "WebSocket Events" heading with a table listing `agent.thinking`, `agent.chunk`, and `agent.done` events

#### Scenario: Backward compatibility noted
- **WHEN** a user reads the WebSocket Events section
- **THEN** there is a note that clients not handling `agent.chunk` will still receive the full response in the RPC result
