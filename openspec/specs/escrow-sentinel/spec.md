## Purpose

Security anomaly detection engine for the on-chain escrow system. Monitors escrow activity via eventbus subscriptions and generates alerts for suspicious patterns.
## Requirements
### Requirement: Sentinel engine with anomaly detection
The system SHALL provide a Sentinel engine that subscribes to eventbus escrow events, runs them through pluggable Detector implementations, and stores generated alerts. The engine SHALL support Start/Stop lifecycle.

#### Scenario: Engine detects rapid deal creation
- **WHEN** more than 5 escrow deals are created from the same peer within 1 minute
- **THEN** a High severity alert of type "rapid_creation" is generated

#### Scenario: Engine detects large withdrawal
- **WHEN** a single escrow release exceeds the configured threshold amount
- **THEN** a High severity alert of type "large_withdrawal" is generated

### Requirement: Five anomaly detectors
The system SHALL implement 5 detectors: RapidCreationDetector (>5 deals/peer/minute), LargeWithdrawalDetector (release > threshold), RepeatedDisputeDetector (>3 disputes/peer/hour), UnusualTimingDetector (create-to-release < 1 minute), BalanceDropDetector (>50% balance drop).

#### Scenario: Unusual timing detection (wash trading)
- **WHEN** a deal is created and released within less than 1 minute
- **THEN** a Medium severity alert of type "unusual_timing" is generated

#### Scenario: Balance drop detection
- **WHEN** contract balance drops more than 50% in a single block
- **THEN** a Critical severity alert of type "balance_drop" is generated

### Requirement: Alert management
Alerts SHALL have fields: ID, Severity (Critical/High/Medium/Low), Type, Message, DealID, Timestamp, Metadata. The engine SHALL support listing alerts by severity, listing active (unacknowledged) alerts, and acknowledging alerts by ID.

#### Scenario: Acknowledge an alert
- **WHEN** Acknowledge is called with a valid alert ID
- **THEN** the alert is marked as acknowledged and excluded from ActiveAlerts

### Requirement: Sentinel agent tools
The system SHALL provide 4 sentinel tools: sentinel_status (safe), sentinel_alerts (safe, with severity filter), sentinel_config (safe), sentinel_acknowledge (dangerous).

#### Scenario: Agent queries sentinel status
- **WHEN** agent calls sentinel_status
- **THEN** system returns running state, total alerts count, active alerts count, and detector names

### Requirement: Sentinel skill definition
The system SHALL provide a `security-sentinel.yaml` skill that allows the agent to monitor escrow activity, with allowed tools: sentinel_status, sentinel_alerts, sentinel_config, sentinel_acknowledge, escrow_status, escrow_list.

#### Scenario: Skill invocation for alerts
- **WHEN** the security-sentinel skill is invoked with action=alerts
- **THEN** the agent calls sentinel_alerts and reports severity levels with recommended actions

### Requirement: Sentinel documentation in economy.md
The system SHALL include a Security Sentinel subsection in `docs/features/economy.md` covering 5 anomaly detectors, alert severity levels, and configuration.

#### Scenario: Detector documentation
- **WHEN** a user reads the Sentinel section in economy.md
- **THEN** they find descriptions of RapidCreation, LargeWithdrawal, RepeatedDispute, UnusualTiming, and BalanceDrop detectors

### Requirement: Sentinel tools in system prompts
The system SHALL list all 4 `sentinel_*` tools in `prompts/TOOL_USAGE.md`: `sentinel_status`, `sentinel_alerts`, `sentinel_config`, `sentinel_acknowledge`.

#### Scenario: Sentinel tool names match code
- **WHEN** the agent reads TOOL_USAGE.md
- **THEN** tool names match those registered in `internal/app/tools_sentinel.go`

### Requirement: Sentinel CLI documentation
The system SHALL document the `lango economy escrow sentinel status` command in `docs/cli/economy.md`.

#### Scenario: Sentinel CLI reference
- **WHEN** a user reads `docs/cli/economy.md`
- **THEN** they find the sentinel status command with description and output format

### Requirement: README reflects sentinel
The system SHALL mention Security Sentinel anomaly detection in `README.md` features.

#### Scenario: Sentinel in README
- **WHEN** a user reads README.md
- **THEN** Security Sentinel is mentioned in the features section

