## MODIFIED Requirements

### Security Configuration
The Security form MUST be expanded to include the following sections:

#### Privacy Interceptor
-   **Enable/Disable**: Toggle switch for the interceptor.
-   **Redact PII**: Toggle switch for automatic PII redaction.
-   **Approval Policy**: Select input (options: "dangerous", "all", "configured", "none"). The value SHALL be read from `cfg.Security.Interceptor.ApprovalPolicy`; if empty, default to "dangerous".
-   **Approval Timeout**: Integer input for timeout seconds.
-   **Notify Channel**: Select input for notification channel.
-   **Sensitive Tools**: Text input (comma-separated tool names).
-   **Exempt Tools**: Text input (comma-separated tool names exempt from approval). The value SHALL be read from `cfg.Security.Interceptor.ExemptTools`.

#### Scenario: ApprovalPolicy select replaces boolean
- **WHEN** user opens the Security form in TUI onboard
- **THEN** the form SHALL display an InputSelect for "Approval Policy" with options ["dangerous", "all", "configured", "none"] instead of the legacy "Approval Req." boolean toggle

#### Scenario: ExemptTools text field present
- **WHEN** user opens the Security form in TUI onboard
- **THEN** the form SHALL display an InputText field for "Exempt Tools" below the "Sensitive Tools" field

#### Scenario: Policy value saved to config
- **WHEN** user selects an approval policy and submits the form
- **THEN** `UpdateConfigFromForm` SHALL set `Security.Interceptor.ApprovalPolicy` to the selected `config.ApprovalPolicy` value

#### Scenario: ExemptTools value saved to config
- **WHEN** user enters comma-separated tool names in Exempt Tools and submits
- **THEN** `UpdateConfigFromForm` SHALL parse and set `Security.Interceptor.ExemptTools` as a trimmed string slice

## REMOVED Requirements

### Requirement: Approval Required boolean toggle
**Reason**: Replaced by ApprovalPolicy select input that supports four granular modes.
**Migration**: The `interceptor_approval` InputBool field is removed from `NewSecurityForm`. The `interceptor_approval` case in `UpdateConfigFromForm` is removed. Legacy `approvalRequired` config values are migrated by the config loader.
