# CLI Onboard Spec

## Goal
The `lango onboard` command provides a guided 5-step wizard for first-time setup. For the full configuration editor, users should use `lango settings`.

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
- **AND** type options SHALL be: anthropic, openai, gemini, ollama

#### Scenario: Step 2 Agent Config
- **WHEN** user advances to Step 2
- **THEN** the wizard SHALL display a form with fields: provider (select), model (text), maxtokens (int), temp (text)
- **AND** provider options SHALL be populated from config.Providers

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

#### Scenario: Step 4 Security form
- **WHEN** user advances to Step 4
- **THEN** the wizard SHALL display a form with fields: interceptor_enabled (bool), interceptor_pii (bool), interceptor_policy (select)
- **AND** policy options SHALL be: dangerous, all, configured, none

#### Scenario: Step 5 Test Results
- **WHEN** user advances to Step 5
- **THEN** the wizard SHALL run 5 configuration validation checks:
  1. Provider exists in providers map with non-empty type
  2. API key is set (non-empty, not placeholder)
  3. Agent model is set
  4. Channel token present (if channel enabled)
  5. config.Validate() passes
- **AND** results SHALL be displayed using pass/warn/fail indicators

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

### Encrypted Profile Storage
The `lango onboard` command SHALL save configuration via `configstore.Store.Save()` to the encrypted SQLite profile store. The `--profile` flag controls the profile name (default: "default").

### Post-save Messaging
After saving, the command SHALL display the profile name, storage path, and next steps including `lango serve`, `lango doctor`, and `lango settings` for fine-tuning.

#### Scenario: Post-save mentions settings
- **WHEN** user saves configuration via onboard
- **THEN** the output SHALL include "lango settings" as a next step for fine-tuning
