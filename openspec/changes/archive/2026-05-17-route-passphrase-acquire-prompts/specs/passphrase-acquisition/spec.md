## ADDED Requirements

### Requirement: Interactive passphrase prompts use acquisition writer
The interactive branch of passphrase acquisition SHALL write visible passphrase prompt text through the writer supplied to the acquisition stream seam.

#### Scenario: Existing passphrase prompt uses injected writer
- **WHEN** `acquireWithIO(...)` falls through to interactive acquisition with creation disabled
- **THEN** the visible `Enter passphrase: ` prompt SHALL be written through the injected writer
- **AND** no prompt text SHALL require process-global stdout capture

#### Scenario: First-run passphrase confirmation uses injected writer
- **WHEN** `acquireWithIO(...)` falls through to interactive acquisition with creation enabled
- **THEN** both visible first-run passphrase prompt strings SHALL be written through the injected writer
- **AND** no prompt text SHALL require process-global stdout capture

#### Scenario: Hidden input behavior is preserved
- **WHEN** interactive acquisition reads the passphrase
- **THEN** hidden input SHALL continue to use the shared terminal password reader path
