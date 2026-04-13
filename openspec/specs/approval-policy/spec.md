## Purpose

The Approval Policy capability provides a policy-based approval system with four modes (dangerous, all, configured, none) that replaces the legacy boolean approval gate. It includes exempt tool overrides, automatic legacy config migration, and a decision function that combines policy, SafetyLevel, and explicit tool lists.
## Requirements
### Requirement: ApprovalPolicy type
The system SHALL define an `ApprovalPolicy` string type with four constants: `"dangerous"` (default), `"all"`, `"configured"`, and `"none"`.

#### Scenario: Policy constants
- **WHEN** ApprovalPolicy constants are referenced
- **THEN** ApprovalPolicyDangerous SHALL equal "dangerous", ApprovalPolicyAll SHALL equal "all", ApprovalPolicyConfigured SHALL equal "configured", ApprovalPolicyNone SHALL equal "none"

### Requirement: InterceptorConfig ApprovalPolicy field
The system SHALL include an `ApprovalPolicy` field in `InterceptorConfig`.

#### Scenario: Config field
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
The system SHALL apply tool approval wrapping based on ApprovalPolicy instead of the legacy boolean check. When policy is "none", no wrapping SHALL occur. For all other policies, each tool SHALL be passed through `wrapWithApproval` with the GrantStore for persistent grant tracking.

#### Scenario: None policy skips wrapping
- **WHEN** ApprovalPolicy is "none"
- **THEN** tools SHALL NOT be wrapped with approval logic

#### Scenario: Dangerous policy wraps tools with GrantStore
- **WHEN** ApprovalPolicy is "dangerous"
- **THEN** all tools SHALL be passed through wrapWithApproval with a GrantStore instance

#### Scenario: Granted tool bypasses approval prompt
- **WHEN** a tool has been previously "Always Allowed" in the current session
- **AND** the same tool is invoked again
- **THEN** the tool SHALL execute without prompting the user

#### Scenario: AlwaysAllow response records grant
- **WHEN** a user responds with "Always Allow" to an approval request
- **THEN** the GrantStore SHALL record the grant for that session and tool
- **AND** subsequent invocations of the same tool in the same session SHALL auto-approve

#### Scenario: Deny response does not record grant
- **WHEN** a user denies an approval request
- **THEN** no grant SHALL be recorded
- **AND** future invocations SHALL still prompt for approval

#### Scenario: Approval denied with session key present
- **WHEN** a tool approval is denied
- **AND** the session key is present in context
- **THEN** the error message SHALL be `tool '<name>' execution denied: user did not approve the action`

#### Scenario: Approval denied with session key missing
- **WHEN** a tool approval is denied
- **AND** the session key is empty in context
- **THEN** the error message SHALL be `tool '<name>' execution denied: no approval channel available (session key missing)`

### Requirement: Tool approval system prompt guidance
The system prompt SHALL include a "Tool Approval" section in `TOOL_USAGE.md` that instructs the AI to correctly interpret approval outcomes.

#### Scenario: User denial guidance
- **WHEN** the AI receives "user did not approve the action" error
- **THEN** the system prompt guidance SHALL instruct the AI to inform the user and offer to retry or suggest alternatives

#### Scenario: Missing channel guidance
- **WHEN** the AI receives "no approval channel available" error
- **THEN** the system prompt guidance SHALL instruct the AI to inform the user of a configuration issue

### Requirement: Example config includes approvalPolicy and exemptTools
The example `config.json` SHALL include `approvalPolicy` and `exemptTools` fields in the `security.interceptor` block to document the new configuration model.

#### Scenario: Example config fields
- **WHEN** a user inspects the example config.json
- **THEN** the `security.interceptor` block SHALL contain `"approvalPolicy": "dangerous"` and `"exemptTools": []`

### Requirement: README documents approvalPolicy
The README Security configuration table SHALL include `security.interceptor.approvalPolicy` with type `string`, default `dangerous`, and description of available policies.

#### Scenario: README table entries
- **WHEN** a user reads the README Security section
- **THEN** the table SHALL list `approvalPolicy` (string, default "dangerous") and `exemptTools` ([]string) as configuration options

### Requirement: P2P approval fallback isolation
The CompositeProvider SHALL provide a dedicated P2P fallback slot (`p2pFallback`) that is used exclusively for approval requests with session keys prefixed with `"p2p:"`. P2P sessions MUST never be routed to the TTY fallback slot, preventing HeadlessProvider from auto-approving remote peer requests.

#### Scenario: P2P session with no P2P fallback configured
- **WHEN** a P2P approval request (session key `"p2p:..."`) is received and no P2P fallback is set
- **THEN** the provider SHALL return an error stating "headless auto-approve is not allowed for remote peers"

#### Scenario: P2P session routes to dedicated fallback
- **WHEN** a P2P approval request is received and a P2P fallback provider is configured
- **THEN** the request SHALL be routed to the P2P fallback provider, not the TTY fallback

#### Scenario: Non-P2P session still uses TTY fallback
- **WHEN** a non-P2P approval request (session key without `"p2p:"` prefix) is received
- **THEN** the request SHALL be routed to the TTY fallback as before

#### Scenario: HeadlessProvider as TTY fallback with P2P request
- **WHEN** HeadlessProvider is configured as TTY fallback and a P2P approval request arrives
- **THEN** HeadlessProvider SHALL NOT be called; the request SHALL use the P2P fallback or be denied

### Requirement: P2P approval wiring
When P2P is enabled, the application SHALL configure `TTYProvider` as the P2P fallback on `CompositeProvider`. This ensures P2P approval requests are always routed to an interactive provider, regardless of whether HeadlessProvider is configured as the TTY fallback.

#### Scenario: P2P enabled wiring
- **WHEN** the application initializes with `cfg.P2P.Enabled = true`
- **THEN** `composite.SetP2PFallback(&approval.TTYProvider{})` SHALL be called

### Requirement: Amount-based auto-approve for payment tools
wrapWithApproval SHALL accept an optional SpendingLimiter parameter (nil allowed). When non-nil and the tool is a payment tool (`p2p_pay` or `payment_send`), it SHALL check the amount parameter against `IsAutoApprovable` before requesting interactive approval.

#### Scenario: Auto-approve small payment
- **WHEN** tool is `p2p_pay` with amount "0.05" AND limiter.IsAutoApprovable returns true
- **THEN** the tool SHALL execute without interactive approval

#### Scenario: Require approval for large payment
- **WHEN** tool is `p2p_pay` with amount "5.00" AND limiter.IsAutoApprovable returns false
- **THEN** the tool SHALL request interactive approval via the approval provider

#### Scenario: No limiter provided
- **WHEN** limiter is nil
- **THEN** wrapWithApproval SHALL behave as before (no amount-based auto-approve)

#### Scenario: Non-payment tool unaffected
- **WHEN** tool is `exec` AND limiter is non-nil
- **THEN** wrapWithApproval SHALL ignore the limiter and follow normal approval policy

### Requirement: Startup warning for "none" policy
The application SHALL emit a WARN-level log message at startup when `approvalPolicy` is set to `"none"`, indicating that all tool calls will execute without user confirmation and that this is not recommended for production.

#### Scenario: Policy "none" startup warning
- **GIVEN** `security.interceptor.approvalPolicy` is set to `"none"`
- **WHEN** the application initializes
- **THEN** a WARN-level log message SHALL be emitted indicating that all tool calls will execute without user confirmation
- **AND** the message SHALL note this is not recommended for production

### Requirement: P2P payment approval summary
buildApprovalSummary SHALL return a human-readable summary for `p2p_pay` tool invocations including amount, peer DID (truncated), and memo.

#### Scenario: p2p_pay approval summary
- **WHEN** buildApprovalSummary is called with toolName "p2p_pay" and params containing amount, peer_did, and memo
- **THEN** it SHALL return a string containing the amount, truncated peer DID, and memo

