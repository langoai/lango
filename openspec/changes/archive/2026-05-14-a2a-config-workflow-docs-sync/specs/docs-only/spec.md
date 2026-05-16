## ADDED Requirements
### Requirement: A2A config workflow examples use real profile-aware commands
Public A2A feature documentation SHALL show the real profile-aware config export/import workflow rather than argument-less placeholders.

#### Scenario: A2A remote-agent config example matches CLI contract
- **WHEN** a reader follows the config export/import example in the A2A feature docs
- **THEN** the example SHALL use `lango config export <name>`
- **AND** it SHALL use `lango config import <file> --profile <name>`
