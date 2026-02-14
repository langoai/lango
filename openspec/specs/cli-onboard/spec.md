# CLI Onboard Spec

## Goal
The `lango onboard` command must provide a comprehensive, interactive configuration editor that allows users to modify all aspects of their `lango.json` configuration file without manual editing.

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

### Persistence
- Configuration MUST be saved to `lango.json`.
- Passwords/Secrets (API Keys, Tokens) MUST be handled securely (though typically stored in env vars, the config references them).
- The tool should generate a `.lango.env` template if new env vars are required.

### Onboard command description reflects actual flow
The `lango onboard` command Long description SHALL accurately list all configurable sections.

#### Scenario: Long description content
- **WHEN** user runs `lango onboard --help`
- **THEN** the description SHALL list Agent, Server, Channels, Tools, Security, Knowledge, and Providers as configurable sections

## Success Criteria
1.  User can launch `lango onboard`, navigate to "Server" settings, change the port, save, and verify `lango.json` is updated.
2.  User can navigate to "Agent" settings, switch provider to Ollama, and save.
3.  Invalid inputs (e.g., Port 99999) are rejected by the UI.

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
- **THEN** the output lango.json includes the updated observationalMemory section with all field values
