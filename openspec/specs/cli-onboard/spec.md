# CLI Onboard Spec

## Goal
The `lango onboard` command must provide a comprehensive, interactive configuration editor that allows users to modify all aspects of their encrypted configuration profile without manual editing.

## Requirements

### Configuration Coverage
The onboarding tool MUST support editing the following configuration sections:
1.  **Agent**:
    - Select Provider (dynamically populated from registered providers; falls back to Anthropic, OpenAI, Gemini, Ollama when no providers are registered)
    - Set Model ID
    - Set Max Tokens (integer)
    - Set Temperature (float)
    - Set System Prompt Path (file path)
    - Select Fallback Provider (empty + same dynamic provider list as Provider)
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

5.  **Session**:
    - Set Database Path (default: `~/.lango/data.db`)
    - Set Session TTL (duration)
    - Set Max History Turns (integer)

6.  **Security**:
    - Toggle Privacy Interceptor Enabled
    - Toggle PII Redaction
    - Toggle Approval Requirement
    - Set Approval Timeout (integer, non-negative)
    - Select Notify Channel (empty, telegram, discord, slack)
    - Set Sensitive Tools (comma-separated)
    - Select Signer Provider (local, rpc, enclave)
    - Set RPC URL
    - Set Key ID

7.  **Knowledge**:
    - Toggle Knowledge System Enabled (boolean)
    - Set Max Learnings (integer)
    - Set Max Knowledge (integer)
    - Set Max Context Per Layer (integer)
    - Toggle Auto Approve Skills (boolean)
    - Set Max Skills Per Day (integer)

8.  **Providers**:
    - Add, edit, and delete multi-provider configurations

#### Scenario: Agent fallback configuration
- **WHEN** user navigates to Agent settings
- **THEN** the form SHALL display fields for system_prompt_path, fallback_provider, and fallback_model
- **AND** fallback_provider SHALL be an InputSelect with options: empty + registered provider IDs

#### Scenario: Agent provider options from registered providers
- **WHEN** user navigates to Agent settings and providers are registered in config
- **THEN** the Provider and Fallback Provider dropdowns SHALL list the registered provider IDs

#### Scenario: Agent provider options with no registered providers
- **WHEN** user navigates to Agent settings and no providers are registered
- **THEN** the Provider dropdown SHALL fall back to default options: anthropic, openai, gemini, ollama

#### Scenario: Browser tool fields in Tools form
- **WHEN** user navigates to Tools settings
- **THEN** the form SHALL display browser_enabled toggle before browser_headless
- **AND** the form SHALL display browser_session_timeout as a duration text field after browser_headless

#### Scenario: Session settings in dedicated form
- **WHEN** user selects "Session" from the configuration menu
- **THEN** the wizard SHALL display a Session form with Database Path, Session TTL, and Max History Turns fields

#### Scenario: Security form without session fields
- **WHEN** user selects "Security" from the configuration menu
- **THEN** the form SHALL NOT include Database Path, Session TTL, Max History Turns, or DB Passphrase fields

#### Scenario: Delete provider from list
- **WHEN** user presses `d` on a provider in the Providers list view
- **THEN** the provider SHALL be removed from the in-memory config state
- **AND** the list SHALL refresh to reflect the deletion

#### Scenario: Cannot delete Add New Provider option
- **WHEN** user presses `d` while the cursor is on "Add New Provider"
- **THEN** no deletion SHALL occur

#### Scenario: Knowledge menu and form
- **WHEN** user views the Configuration Menu
- **THEN** a "Knowledge" category SHALL appear between Security and Observational Memory
- **AND** selecting it SHALL display the Knowledge configuration form with 6 fields

### User Interface
- **Navigation**:
    - Users MUST be able to navigate between configuration categories freely.
    - Uses a menu-based system (e.g., Main Menu -> Category -> Form).
    - The menu SHALL include categories in this order: Providers, Agent, Server, Channels, Tools, Session, Security, Knowledge, Observational Memory, Embedding & RAG, Save & Exit, Cancel.
    - Form cursor navigation SHALL NOT panic when navigating past the first or last field.

#### Scenario: Providers category appears first
- **WHEN** user views the configuration menu
- **THEN** "Providers" SHALL be the first category in the menu, before "Agent"

#### Scenario: Form cursor at top boundary
- **WHEN** user presses up or shift+tab while cursor is on the first field
- **THEN** the cursor SHALL remain on the first field without error

#### Scenario: Session category in menu
- **WHEN** user views the configuration menu
- **THEN** "Session" category SHALL be listed before "Security"

#### Scenario: Knowledge category in menu
- **WHEN** user views the configuration menu
- **THEN** "Knowledge" category SHALL be listed after "Security" and before "Observational Memory"

#### Scenario: Provider creation field order
- **WHEN** user selects "Add New Provider" from the Providers list
- **THEN** the form SHALL display Type selector before Provider Name field
- **AND** the Provider Name field label SHALL be "Provider Name"

- **Validation**:
    - Input fields MUST validate data types (int, float, bool).
    - Port numbers MUST be within valid range (1-65535).
    - Essential fields (like Provider) MUST NOT be empty.
- **Feedback**:
    - Invalid inputs MUST display an error message immediately or upon submission.
    - Changes MUST be explicitly saved or discarded.
    - Provider list help footer SHALL display available key bindings including delete.

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
The system SHALL include an Observational Memory configuration form in the onboard TUI wizard. The form SHALL have fields for: enabled (bool), provider (select, dynamically populated from registered providers with empty option for agent default), model (text), message token threshold (int, positive validation), observation token threshold (int, positive validation), and max message token budget (int, positive validation). The menu entry SHALL appear between Knowledge and Providers with the label "Observational Memory".

#### Scenario: Navigate to OM form
- **WHEN** user selects "Observational Memory" from the onboard menu
- **THEN** the wizard displays the OM configuration form with current values from config

#### Scenario: Edit OM settings
- **WHEN** user modifies fields in the OM form and presses ESC
- **THEN** the changes are saved to the in-memory config state and the wizard returns to the menu

#### Scenario: Invalid threshold value
- **WHEN** user enters a non-positive number in a threshold field
- **THEN** the form displays a validation error "must be a positive integer"

#### Scenario: OM provider field is a select dropdown
- **WHEN** user navigates to Observational Memory settings
- **THEN** the provider field SHALL be an InputSelect dropdown
- **AND** the first option SHALL be an empty string representing "use agent default"
- **AND** subsequent options SHALL be registered provider IDs from buildProviderOptions

#### Scenario: OM provider options with no registered providers
- **WHEN** user navigates to Observational Memory settings and no providers are registered
- **THEN** the provider dropdown SHALL fall back to default options: empty string, anthropic, openai, gemini, ollama

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
