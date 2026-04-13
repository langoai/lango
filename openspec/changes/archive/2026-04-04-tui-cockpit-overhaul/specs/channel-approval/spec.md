## MODIFIED Requirements

### Requirement: Approval request context
The approval request SHALL carry ID, ToolName, SessionKey, Params, Summary, CreatedAt, and additionally SafetyLevel, Category, and Activity as optional string fields for tier classification.

#### Scenario: Request fields
- **WHEN** an approval request is created
- **THEN** it contains ID, ToolName, SessionKey, Params, Summary, CreatedAt

#### Scenario: Summary populated
- **WHEN** the interceptor builds a request for tool "exec" with command "rm -rf /"
- **THEN** Summary is a human-readable string like `Execute command: rm -rf /`

#### Scenario: Empty summary backward compatibility
- **WHEN** a provider receives a request with empty Summary
- **THEN** the provider falls back to displaying ToolName only

#### Scenario: SafetyLevel populated
- **WHEN** the approval middleware creates a request for a dangerous tool
- **THEN** `SafetyLevel` is set to `"dangerous"`

#### Scenario: Category and Activity populated
- **WHEN** the approval middleware creates a request for a filesystem write tool
- **THEN** `Category` is set to `"filesystem"` and `Activity` is set to `"write"`

#### Scenario: Fields omitted for legacy providers
- **WHEN** a channel provider (Slack/Telegram/Discord) receives a request with SafetyLevel/Category/Activity
- **THEN** the provider ignores these fields gracefully (they are optional, omitempty)
