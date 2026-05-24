## MODIFIED Requirements

### Requirement: Interactive command guard tests
CLI tests SHALL cover interactive-only command guards so TUI commands do not drift into non-TTY bootstrap or Bubble Tea startup.

#### Scenario: Onboard guard is covered
- **WHEN** the onboard command's interactive-terminal guard fails
- **THEN** tests SHALL assert the command returns the guard error
- **AND** tests SHALL assert the onboard run path is not invoked

#### Scenario: Settings guard is covered
- **WHEN** the settings command's interactive-terminal guard fails
- **THEN** tests SHALL assert the command returns the guard error
- **AND** tests SHALL assert the settings run path is not invoked
