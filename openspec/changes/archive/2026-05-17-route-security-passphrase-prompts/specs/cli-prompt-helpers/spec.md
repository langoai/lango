## ADDED Requirements

### Requirement: Shared hidden passphrase confirmation supports explicit output streams
The shared CLI prompt package SHALL provide a passphrase confirmation helper that writes all visible hidden-input prompt text through a supplied output stream while preserving terminal-hidden password reading.

#### Scenario: Passphrase confirmation uses explicit output
- **WHEN** the explicit-output passphrase confirmation helper prompts for a passphrase and its confirmation
- **THEN** both visible prompt strings SHALL be written through the supplied output stream
- **AND** the helper SHALL return the confirmed passphrase when both hidden inputs match

#### Scenario: Passphrase confirmation mismatch still fails
- **WHEN** the explicit-output passphrase confirmation helper receives different hidden input values
- **THEN** it SHALL return a mismatch error
- **AND** it SHALL NOT return either entered value
