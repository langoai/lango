## ADDED Requirements

### Requirement: Provenance configuration form
The settings editor SHALL provide a Provenance configuration form in the Automation section with the following fields:

- **Enabled** (`provenance_enabled`) — Boolean toggle
- **Auto on Step Complete** (`provenance_auto_on_step_complete`) — Boolean toggle
- **Auto on Policy** (`provenance_auto_on_policy`) — Boolean toggle
- **Max Per Session** (`provenance_max_per_session`) — Integer input
- **Retention Days** (`provenance_retention_days`) — Integer input

#### Scenario: Edit provenance settings
- **WHEN** the user selects `Provenance` from the settings menu
- **THEN** the editor SHALL display a form with all provenance fields pre-populated from `config.Provenance`

#### Scenario: Save provenance settings
- **WHEN** the user edits provenance fields and navigates back or saves
- **THEN** the config state SHALL be updated through `UpdateConfigFromForm`
- **AND** all edited values SHALL persist into `config.Provenance`

### Requirement: Provenance menu category
The settings menu SHALL include a `provenance` category in the Automation section with title `Provenance`.

#### Scenario: Automation section shows Provenance
- **WHEN** the user navigates the settings menu to the Automation section
- **THEN** the section includes `RunLedger` and `Provenance` as separate categories
