## ADDED Requirements

### Requirement: Security passphrase prompts use command output streams
Security commands that prompt for current, new, or stored passphrases SHALL write visible passphrase prompt text through the Cobra command output stream instead of process-global stdout.

#### Scenario: Change-passphrase prompts use command output
- **WHEN** `lango security change-passphrase` prompts for the current passphrase, new passphrase, and confirmation
- **THEN** each visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Migrate-passphrase prompts use command output
- **WHEN** `lango security migrate-passphrase` prompts for a new passphrase and confirmation
- **THEN** each visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Keyring store prompt uses command output
- **WHEN** `lango security keyring store` prompts for the passphrase to store
- **THEN** the visible hidden-input prompt SHALL be written through the Cobra command output stream

#### Scenario: Warning output remains on command error stream
- **WHEN** passphrase-changing commands emit keyfile or keyring update notices or warnings
- **THEN** those messages SHALL continue to write through the Cobra command error stream
