## ADDED Requirements

### Requirement: ApprovalPolicy type
The system SHALL define an `ApprovalPolicy` string type with four constants: `"dangerous"` (default), `"all"`, `"configured"`, and `"none"`.

#### Scenario: Policy constants
- **WHEN** ApprovalPolicy constants are referenced
- **THEN** ApprovalPolicyDangerous SHALL equal "dangerous", ApprovalPolicyAll SHALL equal "all", ApprovalPolicyConfigured SHALL equal "configured", ApprovalPolicyNone SHALL equal "none"

### Requirement: InterceptorConfig ApprovalPolicy field
The system SHALL add an `ApprovalPolicy` field to `InterceptorConfig`. The legacy `ApprovalRequired` boolean field SHALL be marked deprecated.

#### Scenario: New config field
- **WHEN** InterceptorConfig is defined
- **THEN** it SHALL include `ApprovalPolicy ApprovalPolicy` field with mapstructure tag "approvalPolicy"

### Requirement: ExemptTools config field
The system SHALL add an `ExemptTools []string` field to `InterceptorConfig` that lists tools exempt from approval regardless of policy.

#### Scenario: Exempt tool configuration
- **WHEN** a tool name appears in ExemptTools
- **THEN** that tool SHALL bypass approval regardless of the active policy

### Requirement: Default approval policy
New installations (no existing config) SHALL default to `ApprovalPolicy: "dangerous"` with `Interceptor.Enabled: true`.

#### Scenario: Fresh installation defaults
- **WHEN** no config file exists
- **THEN** the default InterceptorConfig SHALL have Enabled=true and ApprovalPolicy="dangerous"

### Requirement: Legacy config migration
The system SHALL migrate legacy `approvalRequired` + `sensitiveTools` configuration to the new ApprovalPolicy field during config loading, after Unmarshal and before Validate.

#### Scenario: Legacy approvalRequired=true with sensitiveTools
- **WHEN** config has `approvalRequired: true` and `sensitiveTools: ["exec"]`
- **AND** `approvalPolicy` is not set
- **THEN** ApprovalPolicy SHALL be migrated to "configured"

#### Scenario: Legacy approvalRequired=true without sensitiveTools
- **WHEN** config has `approvalRequired: true` and no sensitiveTools
- **AND** `approvalPolicy` is not set
- **THEN** ApprovalPolicy SHALL be migrated to "dangerous"

#### Scenario: Legacy approvalRequired=false
- **WHEN** config has `approvalRequired: false`
- **AND** `approvalPolicy` is not set
- **THEN** ApprovalPolicy SHALL remain empty (inherits default from viper)

#### Scenario: Explicit approvalPolicy takes precedence
- **WHEN** config has `approvalPolicy: "all"` explicitly set
- **THEN** the migration SHALL not modify it regardless of approvalRequired value

### Requirement: needsApproval decision function
The system SHALL implement a `needsApproval(tool, interceptorConfig)` function that determines whether a tool requires approval.

#### Scenario: ExemptTools bypass
- **WHEN** a tool name is in ExemptTools
- **THEN** needsApproval SHALL return false regardless of policy

#### Scenario: SensitiveTools override
- **WHEN** a tool name is in SensitiveTools and not in ExemptTools
- **THEN** needsApproval SHALL return true regardless of policy

#### Scenario: Dangerous policy with dangerous tool
- **WHEN** policy is "dangerous" and tool SafetyLevel is Dangerous (or zero)
- **THEN** needsApproval SHALL return true

#### Scenario: Dangerous policy with safe tool
- **WHEN** policy is "dangerous" and tool SafetyLevel is Safe
- **THEN** needsApproval SHALL return false

#### Scenario: All policy
- **WHEN** policy is "all"
- **THEN** needsApproval SHALL return true for any tool (unless exempt)

#### Scenario: Configured policy
- **WHEN** policy is "configured"
- **THEN** needsApproval SHALL return false for tools not in SensitiveTools

#### Scenario: None policy
- **WHEN** policy is "none"
- **THEN** needsApproval SHALL return false for all tools

#### Scenario: Unknown policy fail-safe
- **WHEN** policy is an unrecognized value
- **THEN** needsApproval SHALL return true (fail-safe)

### Requirement: Policy-based approval gate
The system SHALL apply tool approval wrapping based on ApprovalPolicy instead of the legacy boolean check. When policy is "none", no wrapping SHALL occur. For all other policies, each tool SHALL be passed through `wrapWithApproval`.

#### Scenario: None policy skips wrapping
- **WHEN** ApprovalPolicy is "none"
- **THEN** tools SHALL NOT be wrapped with approval logic

#### Scenario: Dangerous policy wraps tools
- **WHEN** ApprovalPolicy is "dangerous"
- **THEN** all tools SHALL be passed through wrapWithApproval (which internally checks needsApproval)
