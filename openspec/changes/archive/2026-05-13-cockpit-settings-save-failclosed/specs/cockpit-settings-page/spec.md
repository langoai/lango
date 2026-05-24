## ADDED Requirements

### Requirement: Settings page fails closed when persistence is unavailable
The cockpit Settings page SHALL distinguish between editable configuration state and persistence availability.

#### Scenario: Nil config-profile store shows degraded save note
- **WHEN** the Settings page renders with no config-profile store
- **THEN** the page SHALL explain that settings changes cannot be saved because the config profile store is not configured

#### Scenario: Nil config-profile store save callback returns actionable error
- **WHEN** the embedded Settings page save callback runs with no config-profile store
- **THEN** the callback SHALL return an actionable error explaining that the config profile store is not configured
