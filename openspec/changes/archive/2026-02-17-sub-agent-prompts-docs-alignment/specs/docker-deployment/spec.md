## MODIFIED Requirements

### Requirement: Optional prompts volume mount in docker-compose.yml
The docker-compose.yml SHALL include a commented-out volume mount for `./prompts:/usr/share/lango/prompts` to allow runtime prompt customization without rebuilding the Docker image.

#### Scenario: Commented volume mount present
- **WHEN** docker-compose.yml is inspected
- **THEN** it SHALL contain a commented volume mount line for prompts directory customization

#### Scenario: Default behavior unchanged
- **WHEN** docker-compose.yml is used without uncommenting the prompts volume
- **THEN** the container SHALL use the embedded prompts from the Docker image as before
