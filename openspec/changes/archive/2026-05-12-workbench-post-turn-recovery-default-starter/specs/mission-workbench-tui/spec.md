## ADDED Requirements

### Requirement: Failed completed-turn Enter default uses a recovery-oriented starter

The standalone workbench SHALL seed a recovery-oriented default starter when the latest completed turn failed and the operator presses `Enter` from the empty composer.

#### Scenario: Failed turn Enter default uses the recovery starter
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** the latest assistant activity summary represents a failed turn
- **AND** the composer is empty
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL seed the recovery-oriented starter instead of reusing the generic completed-turn default
