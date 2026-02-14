## ADDED Requirements

### Requirement: Secret value registration
The system SHALL provide a SecretScanner that registers known secret values for detection.

#### Scenario: Register secret
- **WHEN** a secret with name "API_KEY" and value "sk-abc123def" is registered
- **THEN** the scanner SHALL store it for future scanning

#### Scenario: Ignore short values
- **WHEN** a secret value shorter than 4 characters is registered
- **THEN** the scanner SHALL ignore it to avoid false positives

### Requirement: Output text scanning
The system SHALL scan text for known secret values and replace them with masked placeholders.

#### Scenario: Single secret detected
- **WHEN** text containing "sk-abc123def" is scanned
- **AND** that value is registered as "API_KEY"
- **THEN** the text SHALL be returned with the value replaced by `[SECRET:API_KEY]`

#### Scenario: Multiple secrets detected
- **WHEN** text containing multiple registered secret values is scanned
- **THEN** all occurrences SHALL be replaced with their respective `[SECRET:name]` placeholders

#### Scenario: No secrets in text
- **WHEN** text not containing any registered secret values is scanned
- **THEN** the text SHALL be returned unchanged

### Requirement: Tool result scanning
The PIIRedactingModelAdapter SHALL scan tool results for secret values before they reach the LLM.

#### Scenario: Tool result contains secret
- **WHEN** a tool result with role "tool" contains a registered secret value
- **THEN** the value SHALL be replaced with `[SECRET:name]` before forwarding to the LLM

### Requirement: Model response scanning
The PIIRedactingModelAdapter SHALL scan model responses for secret values before they reach output channels.

#### Scenario: Model response contains secret
- **WHEN** a model response text contains a registered secret value
- **THEN** the value SHALL be replaced with `[SECRET:name]` in the response

### Requirement: Scanner thread safety
The SecretScanner SHALL be safe for concurrent access from multiple goroutines.

#### Scenario: Concurrent register and scan
- **WHEN** multiple goroutines call Register and Scan concurrently
- **THEN** no data races SHALL occur

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
