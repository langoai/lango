# CLI Onboard Spec

## Goal
The `lango onboard` command must provide a comprehensive, interactive configuration editor that allows users to modify all aspects of their encrypted configuration profile without manual editing.

## Requirements

### Configuration Coverage
The onboarding tool MUST support editing the following configuration sections:
1.  **Agent**:
    - Select Provider (Anthropic, OpenAI, Gemini, Ollama)
    - Set Model ID
    - Set Max Tokens (integer)
    - Set Temperature (float)
    - Set System Prompt Path (file path)
    - Select Fallback Provider (empty, Anthropic, OpenAI, Gemini, Ollama)
    - Set Fallback Model ID

2.  **Server**:
    - Set Host (default: localhost)
    - Set Port (integer, 1-65535)
    - Toggle HTTP Enabled (boolean)
    - Toggle WebSocket Enabled (boolean)

3.  **Channels**:
    - Enable/Disable each supported channel (Telegram, Discord, Slack)
    - Set Bot Tokens for enabled channels
    - Set App Token/Signing Secret for Slack if enabled

4.  **Tools**:
    - Configure Exec Tool: Default Timeout, Allow Background
    - Toggle Browser Enabled (boolean)
    - Toggle Browser Headless mode (boolean)
    - Set Browser Session Timeout (duration)
    - Configure Filesystem Tool: Max Read Size

5.  **Security**:
    - Set Session DB Path
    - Set Session TTL (duration)
    - Set Max History Turns (integer)
    - Toggle Privacy Interceptor Enabled
    - Toggle PII Redaction
    - Toggle Approval Requirement
    - Select Signer Provider (local, rpc, enclave)
    - Set RPC URL
    - Set Key ID
    - Configure Passphrase (for local encryption)

6.  **Knowledge**:
    - Toggle Knowledge System Enabled (boolean)
    - Set Max Learnings (integer)
    - Set Max Knowledge (integer)
    - Set Max Context Per Layer (integer)
    - Toggle Auto Approve Skills (boolean)
    - Set Max Skills Per Day (integer)

7.  **Providers**:
    - Add, edit, remove multi-provider configurations

#### Scenario: Agent fallback configuration
- **WHEN** user navigates to Agent settings
- **THEN** the form SHALL display fields for system_prompt_path, fallback_provider, and fallback_model
- **AND** fallback_provider SHALL be an InputSelect with options: empty, anthropic, openai, gemini, ollama

#### Scenario: Browser tool fields in Tools form
- **WHEN** user navigates to Tools settings
- **THEN** the form SHALL display browser_enabled toggle before browser_headless
- **AND** the form SHALL display browser_session_timeout as a duration text field after browser_headless

#### Scenario: Session max history in Security form
- **WHEN** user navigates to Security settings
- **THEN** the form SHALL display max_history_turns as an integer field after Session TTL

#### Scenario: Knowledge menu and form
- **WHEN** user views the Configuration Menu
- **THEN** a "Knowledge" category SHALL appear between Security and Providers
- **AND** selecting it SHALL display the Knowledge configuration form with 6 fields

### User Interface
- **Navigation**:
    - Users MUST be able to navigate between configuration categories freely.
    - Uses a menu-based system (e.g., Main Menu -> Category -> Form).
    - The menu SHALL include categories: Agent, Server, Channels, Tools, Security, Knowledge, Providers, Save & Exit, Cancel.

#### Scenario: Knowledge category in menu
- **WHEN** user views the configuration menu
- **THEN** "Knowledge" category SHALL be listed after "Security" and before "Providers"
- **Validation**:
    - Input fields MUST validate data types (int, float, bool).
    - Port numbers MUST be within valid range (1-65535).
    - Essential fields (like Provider) MUST NOT be empty.
- **Feedback**:
    - Invalid inputs MUST display an error message immediately or upon submission.
    - Changes MUST be explicitly saved or discarded.

### Requirement: Encrypted config profile storage via onboard
The `lango onboard` command SHALL save configuration via `configstore.Store.Save()` to the encrypted SQLite profile store (`~/.lango/lango.db`) instead of writing plain-text `lango.json` via `config.Save()`.

#### Scenario: Save new profile via onboard
- **WHEN** user completes the onboard wizard and selects "Save & Exit"
- **THEN** the configuration SHALL be saved as an encrypted profile via `configstore.Store.Save()`
- **AND** no `lango.json` file SHALL be created

#### Scenario: Save to named profile
- **WHEN** user runs `lango onboard --profile myprofile` and saves
- **THEN** the configuration SHALL be saved under the profile name "myprofile"

#### Scenario: New profile activation
- **WHEN** a profile with the given name does not exist before onboard
- **THEN** the saved profile SHALL be activated via `configstore.Store.SetActive()`

#### Scenario: Existing profile not re-activated
- **WHEN** a profile with the given name already exists before onboard
- **THEN** the profile's active status SHALL remain unchanged after save

### Persistence
- Passwords/Secrets (API Keys, Tokens) MUST be handled securely (though typically stored in env vars, the config references them).
- The tool should generate a `.lango.env` template if new env vars are required.

### Onboard command description reflects actual flow
The `lango onboard` command Long description SHALL accurately list all configurable sections.

#### Scenario: Long description content
- **WHEN** user runs `lango onboard --help`
- **THEN** the description SHALL list Agent, Server, Channels, Tools, Security, Knowledge, and Providers as configurable sections

## Success Criteria
1.  User can launch `lango onboard`, navigate to "Server" settings, change the port, save, and verify the encrypted profile is updated.
2.  User can navigate to "Agent" settings, switch provider to Ollama, and save.
3.  Invalid inputs (e.g., Port 99999) are rejected by the UI.
4.  No `lango.json` file is created after saving via onboard.

### Requirement: Observational Memory onboard form
The system SHALL include an Observational Memory configuration form in the onboard TUI wizard. The form SHALL have fields for: enabled (bool), provider (text), model (text), message token threshold (int, positive validation), observation token threshold (int, positive validation), and max message token budget (int, positive validation). The menu entry SHALL appear between Knowledge and Providers with the label "Observational Memory".

#### Scenario: Navigate to OM form
- **WHEN** user selects "Observational Memory" from the onboard menu
- **THEN** the wizard displays the OM configuration form with current values from config

#### Scenario: Edit OM settings
- **WHEN** user modifies fields in the OM form and presses ESC
- **THEN** the changes are saved to the in-memory config state and the wizard returns to the menu

#### Scenario: Invalid threshold value
- **WHEN** user enters a non-positive number in a threshold field
- **THEN** the form displays a validation error "must be a positive integer"

### Requirement: Observational Memory config state mapping
The system SHALL map OM form field values to the Config.ObservationalMemory struct fields when the form is submitted. The mapping SHALL handle: om_enabled to Enabled, om_provider to Provider, om_model to Model, om_msg_threshold to MessageTokenThreshold, om_obs_threshold to ObservationTokenThreshold, om_max_budget to MaxMessageTokenBudget.

#### Scenario: Save OM configuration
- **WHEN** user edits OM fields and saves the config
- **THEN** the output encrypted profile includes the updated observationalMemory section with all field values

### Requirement: Profile flag for onboard command
The `lango onboard` command SHALL accept a `--profile` flag to specify the profile name to create or edit. The default value SHALL be "default".

#### Scenario: Default profile name
- **WHEN** user runs `lango onboard` without `--profile`
- **THEN** the wizard SHALL operate on the "default" profile

#### Scenario: Custom profile name
- **WHEN** user runs `lango onboard --profile staging`
- **THEN** the wizard SHALL operate on the "staging" profile

### Requirement: Pre-load existing profile into wizard
The onboard wizard SHALL load an existing profile's configuration as the initial form values when editing a returning user's profile. If no profile exists, the wizard SHALL use `config.DefaultConfig()`.

#### Scenario: Edit existing profile
- **WHEN** user runs `lango onboard` and a "default" profile exists
- **THEN** the wizard forms SHALL be pre-populated with the existing profile's values

#### Scenario: New user onboard
- **WHEN** user runs `lango onboard` and no "default" profile exists
- **THEN** the wizard forms SHALL be pre-populated with default config values

### Requirement: Bootstrap before TUI
The onboard command SHALL run `bootstrap.Run()` to initialize the database, crypto, and configstore before starting the BubbleTea TUI program. This ensures passphrase acquisition does not conflict with TUI terminal capture.

#### Scenario: Passphrase then TUI
- **WHEN** user runs `lango onboard`
- **THEN** the passphrase prompt SHALL appear before the TUI wizard starts

#### Scenario: Bootstrap failure
- **WHEN** bootstrap fails (e.g., DB error, wrong passphrase)
- **THEN** the onboard command SHALL return the bootstrap error without starting the TUI

### Requirement: Updated post-save messaging
After saving, the onboard command SHALL display the profile name, storage path (`~/.lango/lango.db`), and profile management commands (`lango config list`, `lango config use`).

#### Scenario: Post-save output
- **WHEN** user saves configuration via onboard
- **THEN** the output SHALL include the encrypted profile name and storage path
- **AND** the output SHALL include profile management command hints

### Requirement: Save menu text reflects encrypted storage
The "Save & Exit" menu item description SHALL read "Save encrypted profile" instead of "Write config to file".

#### Scenario: Menu description
- **WHEN** user views the configuration menu
- **THEN** the "Save & Exit" item description SHALL be "Save encrypted profile"
