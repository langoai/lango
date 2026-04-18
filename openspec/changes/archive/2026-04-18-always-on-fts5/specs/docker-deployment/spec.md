## ADDED Requirements

### Requirement: Default Docker build does not require an fts5 tag
The default Docker builder stage MUST compile the standard runtime without passing a dedicated `fts5` build tag.

#### Scenario: Docker default build uses normal build command
- **WHEN** the Docker image is built with the default builder stage
- **THEN** the `go build` command does not pass `-tags "fts5"`
- **AND** the resulting image still uses the normal FTS5-capable runtime
