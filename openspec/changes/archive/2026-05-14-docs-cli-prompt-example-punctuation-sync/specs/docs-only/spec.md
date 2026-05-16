## ADDED Requirements

### Requirement: Public CLI prompt examples match current punctuation
Public CLI examples SHALL mirror the current prompt punctuation emitted by the commands they document.

#### Scenario: Shared confirmation prompts include colon before input
- **WHEN** a public doc shows an example for a command using the shared confirmation helper
- **THEN** the prompt example SHALL show the `: ` separator before the entered answer
