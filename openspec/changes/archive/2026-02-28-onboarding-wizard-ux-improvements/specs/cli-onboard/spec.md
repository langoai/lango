## MODIFIED Requirements

### Guided Wizard Flow
The onboard wizard SHALL guide users through 5 sequential steps:
1. **Provider Setup** — Provider type, name, API key, base URL
2. **Agent Config** — Provider selection, model, max tokens, temperature
3. **Channel Setup** — Channel selector (Telegram/Discord/Slack/Skip) then channel-specific form
4. **Security & Auth** — Privacy interceptor enabled, PII redaction, approval policy
5. **Test Configuration** — Validates configuration and displays results

#### Scenario: Step 1 Provider Setup
- **WHEN** user starts the onboard wizard
- **THEN** the wizard SHALL display a form with fields: type (select), id (text), apikey (password), baseurl (text)
- **AND** type options SHALL be: anthropic, openai, gemini, ollama, github
- **AND** every field SHALL have a non-empty Description for inline help

#### Scenario: Step 2 Agent Config
- **WHEN** user advances to Step 2
- **THEN** the wizard SHALL display a form with fields: provider (select), model (text or select), maxtokens (int), temp (text)
- **AND** provider options SHALL be populated from config.Providers, with fallback list including github
- **AND** the model field SHALL attempt auto-fetch via `settings.FetchModelOptions()`; on success it becomes InputSelect, on failure it remains InputText with placeholder
- **AND** every field SHALL have a non-empty Description for inline help

#### Scenario: Step 2 Temperature validation
- **WHEN** user enters a temperature value
- **THEN** the validator SHALL accept values between 0.0 and 2.0 inclusive
- **AND** SHALL reject non-numeric values and values outside the range

#### Scenario: Step 2 Max Tokens validation
- **WHEN** user enters a max tokens value
- **THEN** the validator SHALL accept positive integers only
- **AND** SHALL reject zero, negative integers, and non-integer values

#### Scenario: Step 3 Channel forms descriptions
- **WHEN** user selects any channel (Telegram, Discord, Slack)
- **THEN** every channel form field SHALL have a non-empty Description for inline help

#### Scenario: Step 4 Security form with conditional visibility
- **WHEN** user advances to Step 4
- **THEN** the wizard SHALL display interceptor_enabled (bool) with Description
- **AND** interceptor_pii and interceptor_policy SHALL have VisibleWhen tied to interceptor_enabled.Checked
- **AND** when interceptor is disabled, only interceptor_enabled SHALL be visible (1 field)
- **AND** when interceptor is enabled, all 3 fields SHALL be visible
- **AND** interceptor_pii label SHALL be "  Redact PII" and interceptor_policy label SHALL be "  Approval Policy" (indented)

#### Scenario: GitHub provider suggestion
- **WHEN** the agent provider is "github"
- **THEN** suggestModel SHALL return "gpt-4o"
