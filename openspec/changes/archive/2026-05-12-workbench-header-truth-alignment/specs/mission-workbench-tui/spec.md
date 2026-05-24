## ADDED Requirements

### Requirement: Workbench header reflects incomplete profile setup
The standalone workbench header SHALL show a setup-required model summary when the active profile does not yet have a usable provider/model path.

#### Scenario: Incomplete profile shows setup-required header
- **WHEN** bare `lango` renders the workbench with an incomplete profile
- **THEN** the header SHALL show `Model: Setup required`
- **AND** the empty-state body and composer guidance SHALL continue pointing the operator to setup recovery commands

### Requirement: Workbench header preserves ready profile summary
The standalone workbench header SHALL keep the concrete provider/model summary for ready profiles.

#### Scenario: Ready profile keeps provider and model summary
- **WHEN** bare `lango` renders the workbench with a ready profile
- **THEN** the header SHALL show the configured provider and model summary instead of `Setup required`
