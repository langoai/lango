## ADDED Requirements

### Requirement: Completed-turn empty workbench body names the next-step state

The standalone workbench SHALL describe the completed-turn empty body as a next-step loop instead of the generic no-missions empty state.

#### Scenario: Completed-turn body names the next step
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **THEN** the empty-state body SHALL indicate that the last turn completed and the next step is ready
