## ADDED Requirements

### Requirement: Running-state workbench copy advertises next-prompt staging

The standalone workbench SHALL describe the next-prompt staging path while a submitted starter prompt is still in flight.

#### Scenario: Running-state copy mentions interrupt-and-run follow-up
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** a submitted starter prompt is still streaming
- **THEN** the running-state guidance SHALL mention typing the next prompt
- **AND** SHALL mention pressing `Enter` to interrupt and run it
