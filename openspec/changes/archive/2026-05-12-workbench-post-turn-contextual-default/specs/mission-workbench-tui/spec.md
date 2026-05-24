## ADDED Requirements

### Requirement: Post-turn empty workbench default stays context-appropriate

The standalone workbench SHALL refine its post-turn default `Enter` starter so it still feels like the next step in the current workspace instead of falling back to a mismatched generic review prompt.

#### Scenario: Generic workspace uses structure-oriented post-turn default
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** no detected repo or workspace context is available
- **THEN** the empty-state default `Enter` starter SHALL pivot to the structure-oriented starter instead of the generic recent-changes starter

#### Scenario: Detected workspace keeps next-change post-turn default
- **WHEN** bare `lango` renders an empty Mission Control workbench state with a ready profile
- **AND** at least one prior turn has completed in the current workbench session
- **AND** detected repo or workspace context is available
- **THEN** the empty-state default `Enter` starter SHALL remain the next-change starter
