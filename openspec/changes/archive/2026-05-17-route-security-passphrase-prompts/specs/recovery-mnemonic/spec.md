## ADDED Requirements

### Requirement: Recovery passphrase prompts use command output streams
Recovery commands that prompt for passphrases or recovery mnemonic input SHALL write visible hidden-input prompt text through the Cobra command output stream instead of process-global stdout.

#### Scenario: Recovery setup authorization prompt uses command output
- **WHEN** `lango security recovery setup` prompts for the current passphrase to authorize setup
- **THEN** the visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Recovery restore prompts use command output
- **WHEN** `lango security recovery restore` prompts for the recovery mnemonic, new passphrase, and confirmation
- **THEN** each visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Recovery warning output remains on command error stream
- **WHEN** recovery restore emits keyfile or keyring update notices or warnings
- **THEN** those messages SHALL continue to write through the Cobra command error stream
