## ADDED Requirements

### Requirement: Post-turn empty workbench defaults Enter to the next-step starter

The standalone workbench SHALL shift its default empty-state `Enter` starter after at least one turn completes so the next loop starts from a next-step prompt instead of repeating the initial repository summary prompt.

#### Scenario: Completed turn changes the empty-state default starter
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** there is no staged follow-up draft and no armed starter prompt
- **THEN** the empty-state copy SHALL advertise the next-step starter as the default `Enter` prompt

#### Scenario: Enter seeds the next-step starter after a completed turn
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** there is no staged follow-up draft and no armed starter prompt
- **AND** the operator presses `Enter`
- **THEN** the workbench SHALL arm the next-step starter prompt instead of the original repository-summary starter
