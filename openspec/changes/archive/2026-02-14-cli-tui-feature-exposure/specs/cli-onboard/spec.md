## ADDED Requirements

### Requirement: Observational Memory onboard form
The system SHALL include an Observational Memory configuration form in the onboard TUI wizard. The form SHALL have fields for: enabled (bool), provider (text), model (text), message token threshold (int, positive validation), observation token threshold (int, positive validation), and max message token budget (int, positive validation). The menu entry SHALL appear between Knowledge and Providers with the label "Observational Memory".

#### Scenario: Navigate to OM form
- **WHEN** user selects "Observational Memory" from the onboard menu
- **THEN** the wizard displays the OM configuration form with current values from config

#### Scenario: Edit OM settings
- **WHEN** user modifies fields in the OM form and presses ESC
- **THEN** the changes are saved to the in-memory config state and the wizard returns to the menu

#### Scenario: Invalid threshold value
- **WHEN** user enters a non-positive number in a threshold field
- **THEN** the form displays a validation error "must be a positive integer"

### Requirement: Observational Memory config state mapping
The system SHALL map OM form field values to the Config.ObservationalMemory struct fields when the form is submitted. The mapping SHALL handle: om_enabled to Enabled, om_provider to Provider, om_model to Model, om_msg_threshold to MessageTokenThreshold, om_obs_threshold to ObservationTokenThreshold, om_max_budget to MaxMessageTokenBudget.

#### Scenario: Save OM configuration
- **WHEN** user edits OM fields and saves the config
- **THEN** the output lango.json includes the updated observationalMemory section with all field values
