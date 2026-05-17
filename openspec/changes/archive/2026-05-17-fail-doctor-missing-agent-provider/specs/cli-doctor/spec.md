## MODIFIED Requirements

### Requirement: Verify all providers
The command SHALL verify the status of every provider defined in the `providers` configuration map. When `agent.provider` is set, it SHALL reference an existing provider entry; otherwise the AI provider check SHALL fail with an actionable message.

#### Scenario: Multiple providers configured
- **WHEN** `lango.json` contains both "openai" and "anthropic" in `providers`
- **THEN** the doctor output includes checks for both "OpenAI" and "Anthropic"

#### Scenario: Agent provider missing from providers map
- **WHEN** `agent.provider` is set to "anthropic"
- **AND** the `providers` configuration map has no "anthropic" entry
- **THEN** the AI provider check SHALL fail
- **AND** the failure message SHALL identify the missing `agent.provider` reference

### Requirement: Legacy API Key Verification
The system SHALL verify the legacy API key configuration ONLY IF no modern providers are configured and `agent.provider` is empty.

#### Scenario: API key configured via environment (fallback)
- **WHEN** no providers configured AND GOOGLE_API_KEY environment variable is set
- **AND** `agent.provider` is empty
- **THEN** check passes with warning "Implicit Gemini config found"

#### Scenario: Agent provider missing does not use legacy API key fallback
- **WHEN** no providers are configured
- **AND** `agent.provider` is set to "anthropic"
- **AND** GOOGLE_API_KEY environment variable is set
- **THEN** the AI provider check SHALL fail
- **AND** the failure message SHALL identify the missing `agent.provider` reference

#### Scenario: API key missing
- **WHEN** no providers configured AND no API key found
- **AND** `agent.provider` is empty
- **THEN** check fails with message "No AI providers configured"
