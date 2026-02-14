## ADDED Requirements

### Requirement: Auto-register config secrets
The application SHALL automatically register all sensitive configuration values with the SecretScanner during initialization. This includes provider API keys, provider client secrets, channel bot tokens, channel app tokens, channel signing secrets, and auth provider client secrets.

#### Scenario: Provider API key registered
- **WHEN** the application starts with a provider config containing an API key
- **THEN** the API key value is registered with the SecretScanner for output redaction

#### Scenario: Channel token registered
- **WHEN** the application starts with channel configs containing bot tokens
- **THEN** all non-empty token values are registered with the SecretScanner

#### Scenario: Empty values skipped
- **WHEN** a config field is empty
- **THEN** the empty value is not registered with the SecretScanner
