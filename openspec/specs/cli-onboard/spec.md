# CLI Onboard Spec

## Goal
The `lango onboard` command provides a guided 5-step wizard for first-time setup. For the full configuration editor, users should use `lango settings`.

## Purpose

Capability spec for cli-onboard. See requirements below for scope and behavior contracts.
## Requirements

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

#### Scenario: Step 3 Channel Selector
- **WHEN** user advances to Step 3
- **THEN** the wizard SHALL display a channel selector with options: Telegram, Discord, Slack, Skip
- **AND** selecting a channel SHALL enable it and show the channel-specific token form
- **AND** selecting "Skip" SHALL advance to Step 4

#### Scenario: Step 3 Telegram form
- **WHEN** user selects Telegram from the channel selector
- **THEN** the form SHALL display a single telegram_token password field

#### Scenario: Step 3 Slack form
- **WHEN** user selects Slack from the channel selector
- **THEN** the form SHALL display slack_token and slack_app_token password fields

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
- **AND** policy options SHALL be: dangerous, all, configured, none

#### Scenario: GitHub provider suggestion
- **WHEN** the agent provider is "github"
- **THEN** suggestModel SHALL return "gpt-4o"

#### Scenario: Step 5 Test Results
- **WHEN** user advances to Step 5
- **THEN** the wizard SHALL run 5 configuration validation checks:
  1. Provider exists in providers map with non-empty type
  2. API key is set (non-empty, not placeholder)
  3. Agent model is set
  4. Channel token present (if channel enabled)
  5. config.Validate() passes
- **AND** results SHALL be displayed using pass/warn/fail indicators

### Agent step reactive model list
The Onboard Agent step form SHALL wire `OnChange` on the provider field to asynchronously fetch and update the model field when the provider changes. The model field SHALL use `InputSearchSelect` type.

#### Scenario: Provider change in onboard triggers model refresh
- **WHEN** a user changes the provider in the Agent step of the onboard wizard
- **THEN** the model field SHALL show loading state, fetch models from the new provider, and update the placeholder with `suggestModel(newProvider)`

#### Scenario: Model fetch error shows feedback
- **WHEN** model fetching fails during onboard Agent step
- **THEN** the model field SHALL fall back to `InputText` with an error message in the description

### Wizard forwards async messages
The onboard Wizard's `Update()` method SHALL forward non-key, non-window messages to the active form so that `FieldOptionsLoadedMsg` and other async results reach the form's update handler.

#### Scenario: FieldOptionsLoadedMsg reaches active form
- **WHEN** the Wizard receives a `FieldOptionsLoadedMsg` while on a form step
- **THEN** the message SHALL be forwarded to `activeForm.Update()` for processing

### Navigation
- `Ctrl+N` SHALL save the current form and advance to the next step
- `Ctrl+P` SHALL save the current form and go back one step
- `Ctrl+C` SHALL cancel and quit without saving
- `Esc` on Step 1 SHALL quit; on other steps SHALL go back

#### Scenario: Navigate forward
- **WHEN** user presses Ctrl+N on any step
- **THEN** the current form values SHALL be saved to the config state
- **AND** the wizard SHALL advance to the next step

#### Scenario: Navigate backward
- **WHEN** user presses Ctrl+P on Step 2+
- **THEN** the current form values SHALL be saved to the config state
- **AND** the wizard SHALL go back to the previous step

#### Scenario: Complete wizard
- **WHEN** user presses Enter on Step 5 (Test Results)
- **THEN** the wizard SHALL save configuration and exit

### Progress Indicator
The wizard SHALL display a progress bar showing the current step, total steps, and step name. A vertical step list SHALL show completed steps (check mark), current step (pointer), and pending steps (circle).

#### Scenario: Progress bar display
- **WHEN** user is on Step 2
- **THEN** the progress bar SHALL show "[Step 2/5]" with a partially filled bar
- **AND** the step list SHALL show Step 1 with a check mark and Step 2 with a pointer

### Configuration Validation
The test step SHALL validate:
1. Provider exists and has a non-empty type
2. API key is set (empty → fail, placeholder → warn, ollama → pass without key)
3. Agent model is non-empty
4. Channel tokens are present for enabled channels (no channels → warn)
5. config.Validate() passes

#### Scenario: Test step runs the five validation categories
- **WHEN** user advances to the Test Configuration step
- **THEN** the wizard SHALL validate provider presence, API key status, agent model presence, enabled channel tokens, and `config.Validate()` success
- **AND** SHALL surface the results using pass, warn, or fail indicators

### Encrypted Profile Storage
The `lango onboard` command SHALL save configuration via `configstore.Store.Save()` to the encrypted SQLite profile store. The `--profile` flag controls the profile name (default: "default").

#### Scenario: Onboard saves to the encrypted profile store
- **WHEN** user completes the onboard wizard
- **THEN** the command SHALL persist the configuration through `configstore.Store.Save()`
- **AND** SHALL use the `--profile` flag value, or `default` when the flag is omitted

### Post-save Messaging
After saving, the command SHALL display the profile name, storage path, and next steps including `lango serve`, `lango doctor`, and `lango settings` for fine-tuning.

#### Scenario: Post-save mentions settings
- **WHEN** user saves configuration via onboard
- **THEN** the output SHALL include "lango settings" as a next step for fine-tuning

### Onboard preset support
The onboard command SHALL accept a `--preset` flag (minimal, researcher, collaborator, full) that initializes the wizard config from the named preset instead of DefaultConfig().

#### Scenario: Onboard with preset
- **WHEN** user runs `lango onboard --preset researcher`
- **THEN** wizard starts with researcher preset config (Knowledge, Graph, etc. pre-enabled)

#### Scenario: Invalid preset
- **WHEN** user runs `lango onboard --preset invalid`
- **THEN** command returns error listing valid presets

### Config-aware next steps
After onboard completion, the system SHALL display recommended features that are currently disabled, with the settings category name for each.

#### Scenario: Default config recommendations
- **WHEN** onboard completes with default config (Knowledge, ObsMemory, Cron, MCP all disabled)
- **THEN** next steps shows all four as recommendations with their settings category names

#### Scenario: Researcher preset recommendations
- **WHEN** onboard completes with researcher preset (Knowledge enabled)
- **THEN** Knowledge is NOT listed in recommendations (already enabled); Cron and MCP are listed

### Preset hints in next steps
The next steps output SHALL include quick preset commands for creating additional profiles.

#### Scenario: Preset commands shown
- **WHEN** onboard completes
- **THEN** output includes example commands: `lango config create <name> --preset researcher/collaborator/full`

### Requirement: Onboard post-save output mentions all primary next steps
After saving, the onboard command SHALL display next-step guidance that mentions the default workbench entry point, the live runtime entry point, the verification command, and the full configuration editor.

#### Scenario: Post-save mentions primary entry points
- **WHEN** user saves configuration via onboard
- **THEN** the output SHALL include `lango` as the default mission workbench entry point
- **AND** SHALL include `lango serve` as the live runtime entry point
- **AND** SHALL include `lango doctor` as the verification step
- **AND** SHALL include `lango settings` as the full-editor follow-up

### Requirement: Onboard command output routing
`lango onboard` SHALL write its preset banner, cancel message, and post-save guidance through the Cobra command output stream so wrappers and test harnesses can capture non-TUI completion output without intercepting process-global stdout.

#### Scenario: Onboard next-steps guidance writes to command output
- **WHEN** onboard completes successfully
- **THEN** the post-save guidance writes to the Cobra command output stream

### Requirement: Onboard command requires an interactive terminal
The `lango onboard` command SHALL fail before bootstrap or TUI startup when the current stdin is not an interactive terminal.

#### Scenario: Non-interactive onboard fails with scripted guidance
- **WHEN** `lango onboard` is invoked while stdin is not an interactive terminal
- **THEN** the command SHALL return an error that says onboard requires an interactive terminal
- **AND** the error SHALL guide scripted setup toward `lango config create --preset <name>` or `lango config import`
- **AND** the command SHALL NOT start the onboard wizard or save a profile

### Requirement: Advanced feature guidance stays outside the five-step wizard
Product guidance for advanced systems SHALL describe them as settings-editor or config-import/export tasks instead of as dedicated onboarding submenu paths.

#### Scenario: Advanced features are routed to settings or config workflows
- **WHEN** a user needs to configure advanced systems such as embedding, graph, multi-agent, A2A, prompts, or OIDC providers
- **THEN** the product guidance SHALL direct them to `lango settings` or `lango config import/export`
- **AND** SHALL NOT describe nonexistent dedicated onboard submenu paths for those advanced systems

### Requirement: Onboard test step shares the workbench readiness contract

The onboard Test Configuration step SHALL evaluate provider/model/API-key completeness with the same agent-readiness rules used by the workbench setup-recovery flow.

#### Scenario: Missing remote API key fails onboard validation
- **WHEN** the selected agent provider and model target a non-ollama provider and the provider API key is empty
- **THEN** the Test Configuration step SHALL fail the API key check
- **AND** SHALL continue treating the profile as incomplete until the key is provided

#### Scenario: Ollama passes onboard validation without an API key
- **WHEN** the selected agent provider and model target an ollama provider and the provider API key is empty
- **THEN** the Test Configuration step SHALL pass the API key check
- **AND** SHALL keep that profile eligible for ready-state workbench behavior after save
