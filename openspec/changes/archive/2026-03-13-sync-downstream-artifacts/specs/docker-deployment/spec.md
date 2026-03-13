## MODIFIED Requirements

### Requirement: Docker Compose includes workspace volume
The Docker Compose configuration SHALL define a `lango-workspaces` named volume for P2P workspace data persistence.

#### Scenario: Workspace volume defined
- **WHEN** a user inspects `docker-compose.yml` volumes section
- **THEN** `lango-workspaces` SHALL be listed as a named volume

### Requirement: Docker Compose references team and economy env vars
The Docker Compose configuration SHALL include commented `LANGO_TEAM` and `LANGO_ECONOMY` environment variables for optional feature activation.

#### Scenario: Team env var present
- **WHEN** a user inspects `docker-compose.yml` environment section
- **THEN** `LANGO_TEAM=true` SHALL be present as a commented variable

#### Scenario: Economy env var present
- **WHEN** a user inspects `docker-compose.yml` environment section
- **THEN** `LANGO_ECONOMY=true` SHALL be present as a commented variable
