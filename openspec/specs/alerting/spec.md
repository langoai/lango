## ADDED Requirements

### Requirement: Alert dispatcher monitors policy block rate
The alerting dispatcher SHALL subscribe to PolicyDecisionEvent on the EventBus and track block verdicts within a sliding 5-minute window. When the number of blocks exceeds the configured `PolicyBlockRate` threshold, the dispatcher SHALL publish an AlertEvent with type "policy_block_rate" and severity "warning".

#### Scenario: Policy block rate exceeds threshold
- **WHEN** the number of policy block decisions in the last 5 minutes exceeds the configured PolicyBlockRate threshold
- **THEN** the dispatcher publishes an AlertEvent with type="policy_block_rate", severity="warning", and a message containing the count and threshold

#### Scenario: Policy block rate within threshold
- **WHEN** the number of policy block decisions in the last 5 minutes is at or below the threshold
- **THEN** no AlertEvent is published

### Requirement: Alert deduplication
The alerting dispatcher SHALL deduplicate alerts by type within each 5-minute window. Only one alert per type per window SHALL be published to prevent alert storms.

#### Scenario: Duplicate alert suppressed
- **WHEN** an alert of the same type was already published within the current 5-minute window
- **THEN** no additional AlertEvent is published for that type until the window advances

### Requirement: Alert configuration
The system SHALL provide an AlertingConfig struct with fields: Enabled (bool), AdminChannel (string), PolicyBlockRate (int), and RecoveryRetries (int). Default values SHALL be Enabled=false, PolicyBlockRate=10, RecoveryRetries=5.

#### Scenario: Alerting disabled by default
- **WHEN** the system starts with default configuration
- **THEN** the alerting dispatcher is not created and no alert monitoring occurs

#### Scenario: Alerting enabled with custom thresholds
- **WHEN** alerting.enabled=true and alerting.policyBlockRateThreshold=20 in configuration
- **THEN** the dispatcher uses 20 as the policy block rate threshold

### Requirement: CLI alerts list command
The system SHALL provide a `lango alerts list` CLI command that queries the gateway `/alerts` endpoint and displays recent alerts. The command SHALL support `--days` flag (default: 7) and `--output table|json` format.

#### Scenario: List alerts in table format
- **WHEN** user runs `lango alerts list`
- **THEN** the system displays alerts from the last 7 days in table format with columns: time, type, severity, message

#### Scenario: List alerts in JSON format
- **WHEN** user runs `lango alerts list --output json`
- **THEN** the system outputs alerts as a JSON array

### Requirement: CLI alerts summary command
The system SHALL provide a `lango alerts summary` CLI command that queries the gateway `/alerts` endpoint and displays aggregated alert counts by type.

#### Scenario: Summary with alerts
- **WHEN** user runs `lango alerts summary` and alerts exist
- **THEN** the system displays a count of alerts grouped by type

### Requirement: HTTP alerts endpoint
The system SHALL expose a GET `/alerts` HTTP endpoint that queries the audit log for records with action="alert" and returns them as JSON. The endpoint SHALL accept a `days` query parameter (default: 7).

#### Scenario: Query alerts via HTTP
- **WHEN** a GET request is made to `/alerts?days=3`
- **THEN** the system returns a JSON response containing alert records from the last 3 days
