## Purpose

Define the `lango settings` command that provides a comprehensive, interactive menu-based configuration editor for all aspects of the encrypted configuration profile.
## Requirements
### Requirement: Configuration Coverage
The settings editor SHALL support editing all configuration sections, including RunLedger (Task OS) configuration.

#### Scenario: RunLedger category appears in Automation
- **WHEN** user launches `lango settings`
- **THEN** the `Automation` section SHALL include `RunLedger` alongside Cron Scheduler, Background Tasks, and Workflow Engine

### Requirement: User Interface
The settings editor SHALL provide menu-based navigation with a two-level hierarchy, free navigation between categories at Level 2, and shared `tuicore.FormModel` for all forms. Provider and OIDC provider list views SHALL support managing collections. Pressing Esc at Level 1 of StepMenu SHALL navigate back to StepWelcome. Pressing Esc at Level 2 SHALL navigate back to Level 1 with cursor restored. The help bar at Level 1 SHALL omit the Tab hint. The help bar at Level 2 SHALL include the Tab hint.

#### Scenario: Launch settings
- **WHEN** user runs `lango settings`
- **THEN** the editor SHALL display a welcome screen followed by the Level 1 section list

#### Scenario: Save from settings
- **WHEN** user selects "Save & Exit" from Level 1
- **THEN** the configuration SHALL be saved as an encrypted profile

#### Scenario: Esc at Welcome screen quits
- **WHEN** user presses Esc at the Welcome screen (StepWelcome)
- **THEN** the TUI SHALL quit

#### Scenario: Esc at Level 1 navigates back to Welcome
- **WHEN** user presses Esc at Level 1 while not in search mode
- **THEN** the editor SHALL navigate back to StepWelcome without quitting

#### Scenario: Esc at Level 2 navigates back to Level 1
- **WHEN** user presses Esc at Level 2 while not in search mode
- **THEN** the menu SHALL return to Level 1 with cursor restored to the section position

#### Scenario: Esc at Menu during search cancels search
- **WHEN** user presses Esc at the settings menu while search mode is active
- **THEN** the search SHALL be cancelled and the menu SHALL remain at StepMenu

#### Scenario: Ctrl+C always quits
- **WHEN** user presses Ctrl+C at any step
- **THEN** the TUI SHALL quit immediately with Cancelled flag set

#### Scenario: Menu help bar at Level 1
- **WHEN** the settings menu is at Level 1 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search, Back (no Tab hint)

#### Scenario: Menu help bar at Level 2
- **WHEN** the settings menu is at Level 2 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search, Tab (filter label), Back

### Requirement: Skill configuration form
The settings editor SHALL provide a Skill configuration form with the following fields:
- **Enabled** (`skill_enabled`) — Boolean toggle for enabling the file-based skill system
- **Skills Directory** (`skill_dir`) — Text input for the directory path containing SKILL.md files

#### Scenario: Edit skill settings
- **WHEN** user selects "Skill" from the settings menu
- **THEN** the editor SHALL display a form with Enabled toggle and Skills Directory text field pre-populated from `config.Skill`

#### Scenario: Save skill settings
- **WHEN** user edits skill fields and navigates back (Esc)
- **THEN** the changes SHALL be applied to `config.Skill.Enabled` and `config.Skill.SkillsDir`

### Requirement: Cron Scheduler configuration form
The settings editor SHALL provide a Cron Scheduler configuration form with the following fields:
- **Enabled** (`cron_enabled`) — Boolean toggle
- **Timezone** (`cron_timezone`) — Text input for timezone (e.g., "UTC", "Asia/Seoul")
- **Max Concurrent Jobs** (`cron_max_jobs`) — Integer input
- **Session Mode** (`cron_session_mode`) — Select: isolated, main
- **History Retention** (`cron_history_retention`) — Text input for retention duration
- **Default Deliver To** (`cron_default_deliver`) — Text input, comma-separated channel names

#### Scenario: Edit cron settings
- **WHEN** user selects "Cron Scheduler" from the settings menu
- **THEN** the editor SHALL display a form with all cron fields pre-populated from `config.Cron`

### Requirement: Background Tasks configuration form
The settings editor SHALL provide a Background Tasks configuration form with the following fields:
- **Enabled** (`bg_enabled`) — Boolean toggle
- **Yield Time (ms)** (`bg_yield_ms`) — Integer input
- **Max Concurrent Tasks** (`bg_max_tasks`) — Integer input
- **Default Deliver To** (`bg_default_deliver`) — Text input, comma-separated channel names

#### Scenario: Edit background settings
- **WHEN** user selects "Background Tasks" from the settings menu
- **THEN** the editor SHALL display a form with all background fields pre-populated from `config.Background`

### Requirement: Workflow Engine configuration form
The settings editor SHALL provide a Workflow Engine configuration form with the following fields:
- **Enabled** (`wf_enabled`) — Boolean toggle
- **Max Concurrent Steps** (`wf_max_steps`) — Integer input
- **Default Timeout** (`wf_timeout`) — Text input for duration (e.g., "10m")
- **State Directory** (`wf_state_dir`) — Text input for directory path
- **Default Deliver To** (`wf_default_deliver`) — Text input, comma-separated channel names

#### Scenario: Edit workflow settings
- **WHEN** user selects "Workflow Engine" from the settings menu
- **THEN** the editor SHALL display a form with all workflow fields pre-populated from `config.Workflow`

### Requirement: Librarian configuration form
The settings editor SHALL provide a Librarian configuration form with the following fields:
- **Enabled** (`lib_enabled`) — Boolean toggle for enabling the proactive librarian system
- **Observation Threshold** (`lib_obs_threshold`) — Integer input (positive) for minimum observation count to trigger analysis
- **Inquiry Cooldown Turns** (`lib_cooldown`) — Integer input (non-negative) for turns between inquiries per session
- **Max Pending Inquiries** (`lib_max_inquiries`) — Integer input (non-negative) for maximum pending inquiries per session
- **Auto-Save Confidence** (`lib_auto_save`) — Select input with options: "high", "medium", "low"
- **Provider** (`lib_provider`) — Select input with "" (empty = agent default) + registered providers
- **Model** (`lib_model`) — Text input for model ID

#### Scenario: Edit librarian settings
- **WHEN** user selects "Librarian" from the settings menu
- **THEN** the editor SHALL display a form with all 7 fields pre-populated from `config.Librarian`

#### Scenario: Save librarian settings
- **WHEN** user edits librarian fields and navigates back (Esc)
- **THEN** the config state SHALL be updated with the new values via `UpdateConfigFromForm()`

### Requirement: Settings forms for default delivery channels
The Cron, Background, and Workflow settings forms SHALL each include a "Default Deliver To" text input field that accepts comma-separated channel names. The state update handler SHALL map these fields to the respective config DefaultDeliverTo slices using the splitCSV helper.

#### Scenario: Cron default deliver field
- **WHEN** the user opens the Cron Scheduler settings form
- **THEN** the form SHALL display a "Default Deliver To" field with placeholder "telegram,discord,slack (comma-separated)"

#### Scenario: Background default deliver field
- **WHEN** the user opens the Background Tasks settings form
- **THEN** the form SHALL display a "Default Deliver To" field with placeholder "telegram,discord,slack (comma-separated)"

#### Scenario: Workflow default deliver field
- **WHEN** the user opens the Workflow Engine settings form
- **THEN** the form SHALL display a "Default Deliver To" field with placeholder "telegram,discord,slack (comma-separated)"

#### Scenario: State update mapping
- **WHEN** the user enters "telegram,discord" in the cron default deliver field
- **THEN** the config state SHALL update Cron.DefaultDeliverTo to ["telegram", "discord"]

### Requirement: Observational Memory context limit fields in settings form
The Observational Memory settings form SHALL include fields for configuring context limits:
- **Max Reflections in Context** (`om_max_reflections`) — Integer input (non-negative, 0 = unlimited)
- **Max Observations in Context** (`om_max_observations`) — Integer input (non-negative, 0 = unlimited)

The state update handler SHALL map these fields to `ObservationalMemory.MaxReflectionsInContext` and `ObservationalMemory.MaxObservationsInContext`.

#### Scenario: Edit context limit fields
- **WHEN** user selects "Observational Memory" from the settings menu
- **THEN** the form SHALL display "Max Reflections in Context" and "Max Observations in Context" fields pre-populated from `config.ObservationalMemory`

#### Scenario: Save context limit values
- **WHEN** user sets Max Reflections in Context to 10 and Max Observations in Context to 50
- **THEN** the config state SHALL update `ObservationalMemory.MaxReflectionsInContext` to 10 and `ObservationalMemory.MaxObservationsInContext` to 50

#### Scenario: Zero means unlimited
- **WHEN** user sets Max Reflections in Context to 0
- **THEN** the value SHALL be accepted (0 = unlimited) and stored as 0

### Requirement: Security form PII pattern fields
The Security configuration form SHALL include fields for managing PII patterns: disabled builtin patterns (comma-separated text), custom patterns (name:regex comma-separated text), Presidio enabled (bool), Presidio URL (text), and Presidio language (text).

#### Scenario: Disabled patterns field
- **WHEN** the Security form is created
- **THEN** it SHALL contain field with key "interceptor_pii_disabled"

#### Scenario: Custom patterns field
- **WHEN** the Security form is created with custom patterns {"a": "\\d+"}
- **THEN** it SHALL contain field with key "interceptor_pii_custom" showing "a:\\d+" format

#### Scenario: Presidio fields
- **WHEN** the Security form is created
- **THEN** it SHALL contain fields "presidio_enabled", "presidio_url", "presidio_language"

### Requirement: State update for PII fields
The ConfigState.UpdateConfigFromForm SHALL map the new PII form keys to their corresponding config fields.

#### Scenario: Update disabled patterns
- **WHEN** form field "interceptor_pii_disabled" has value "passport,ipv4"
- **THEN** config PIIDisabledPatterns SHALL be ["passport", "ipv4"]

#### Scenario: Update custom patterns
- **WHEN** form field "interceptor_pii_custom" has value "my_id:\\bID-\\d+\\b"
- **THEN** config PIICustomPatterns SHALL contain {"my_id": "\\bID-\\d+\\b"}

#### Scenario: Update Presidio enabled
- **WHEN** form field "presidio_enabled" is checked
- **THEN** config Presidio.Enabled SHALL be true

### Requirement: Security form signer provider options
The Security form's signer provider dropdown SHALL include options for all supported providers: local, rpc, enclave, aws-kms, gcp-kms, azure-kv, pkcs11.

#### Scenario: KMS providers available in signer dropdown
- **WHEN** user opens the Security form
- **THEN** the signer provider dropdown SHALL include "aws-kms", "gcp-kms", "azure-kv", and "pkcs11" as options

### Requirement: P2P Network settings form
The settings TUI SHALL provide a "P2P Network" form with 14 fields covering core P2P networking: enabled, listen addresses, bootstrap peers, relay, mDNS, max peers, handshake timeout, session token TTL, auto-approve known peers, gossip interval, ZK handshake, ZK attestation, require signed challenge, and min trust score.

#### Scenario: User enables P2P networking
- **WHEN** user navigates to "P2P Network" and sets Enabled to true
- **THEN** the config's `p2p.enabled` field SHALL be set to true upon save

#### Scenario: User sets listen addresses
- **WHEN** user enters comma-separated multiaddrs in "Listen Addresses"
- **THEN** the config's `p2p.listenAddrs` SHALL contain each address as a separate array element

### Requirement: P2P ZKP settings form
The settings TUI SHALL provide a "P2P ZKP" form with fields for proof cache directory, proving scheme (plonk/groth16), SRS mode (unsafe/file), SRS path, and max credential age.

#### Scenario: User selects groth16 proving scheme
- **WHEN** user selects "groth16" from the proving scheme dropdown
- **THEN** the config's `p2p.zkp.provingScheme` SHALL be set to "groth16"

### Requirement: P2P Pricing settings form
The settings TUI SHALL provide a "P2P Pricing" form with fields for enabled, price per query, and tool-specific prices (as key:value comma-separated text).

#### Scenario: User sets tool prices
- **WHEN** user enters "exec:0.10,browser:0.50" in the Tool Prices field
- **THEN** the config's `p2p.pricing.toolPrices` SHALL be a map with keys "exec" and "browser"

### Requirement: P2P Owner Protection settings form
The settings TUI SHALL provide a "P2P Owner Protection" form with fields for owner name, email, phone, extra terms, and block conversations. The block conversations field SHALL default to checked when the config value is nil.

#### Scenario: User sets block conversations with nil default
- **WHEN** the config's `blockConversations` is nil
- **THEN** the form SHALL display the checkbox as checked (default true)

#### Scenario: User unchecks block conversations
- **WHEN** user unchecks "Block Conversations"
- **THEN** the config's `p2p.ownerProtection.blockConversations` SHALL be a pointer to false

### Requirement: P2P Sandbox settings form
The settings TUI SHALL provide a "P2P Sandbox" form with fields for tool isolation (enabled, timeout, max memory) and container sandbox (enabled, runtime, image, network mode, read-only rootfs, CPU quota, pool size, pool idle timeout). Container-specific fields SHALL only be visible when Container Sandbox is enabled.

#### Scenario: User configures container sandbox
- **WHEN** user enables container sandbox and selects "docker" runtime
- **THEN** the config's `p2p.toolIsolation.container.enabled` SHALL be true and `runtime` SHALL be "docker"

#### Scenario: Container read-only rootfs defaults to true
- **WHEN** the config's `readOnlyRootfs` is nil
- **THEN** the form SHALL display the checkbox as checked (default true)

### Requirement: Security Keyring settings form
The settings TUI SHALL provide a "Security Keyring" form with a single field for OS keyring enabled/disabled.

#### Scenario: User enables keyring
- **WHEN** user checks "OS Keyring Enabled"
- **THEN** the config's `security.keyring.enabled` SHALL be set to true

### Requirement: Security DB Encryption settings form
The settings TUI SHALL provide a "Security DB Encryption" form with fields for SQLCipher encryption enabled and cipher page size.

#### Scenario: User enables DB encryption
- **WHEN** user checks "SQLCipher Encryption" and sets page size to 4096
- **THEN** the config SHALL have `security.dbEncryption.enabled` true and `cipherPageSize` 4096

#### Scenario: Cipher page size validation
- **WHEN** user enters 0 or a negative number for cipher page size
- **THEN** the form SHALL display a validation error "must be a positive integer"

### Requirement: Security KMS settings form
The settings TUI SHALL provide a "Security KMS" form with conditional field visibility based on the selected backend. Cloud KMS fields (region, endpoint) appear for aws-kms/gcp-kms/azure-kv. Azure-specific fields appear for azure-kv. PKCS#11 fields appear for pkcs11. Common fields (key ID, fallback, timeout, retries) appear for all non-local backends.

#### Scenario: User configures AWS KMS
- **WHEN** user selects "aws-kms" and enters region and key ARN
- **THEN** the config's `security.kms.region` and `security.kms.keyId` SHALL contain the entered values

#### Scenario: PKCS#11 PIN is password field
- **WHEN** the KMS form is displayed with pkcs11 backend selected
- **THEN** the PKCS#11 PIN field SHALL use InputPassword type to mask the value

#### Scenario: Local backend hides KMS fields
- **WHEN** user selects "local" as the KMS backend
- **THEN** all KMS-specific fields SHALL be hidden

### Requirement: Grouped Section Layout
The settings menu SHALL organize categories into named sections using a two-level hierarchical navigation. Level 1 SHALL display 7 named sections (with category counts) plus Save & Cancel. Selecting a section at Level 1 SHALL drill into Level 2 showing only that section's categories. Esc at Level 2 SHALL return to Level 1 with cursor restored. Tab SHALL only toggle Basic/Advanced filtering at Level 2.

The sections SHALL be, in order:
1. **Core** — Providers, Agent, Channels, Tools, Server (advanced), Session (advanced), Logging (advanced), Gatekeeper (advanced), Output Manager (advanced)
2. **AI & Knowledge** — Context Profile, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store (advanced), Librarian (advanced), Retrieval (advanced), Auto-Adjust (advanced), Context Budget (advanced), Agent Memory (advanced), Multi-Agent (advanced), A2A Protocol (advanced), Hooks (advanced), Ontology (advanced)
3. **Automation** — Cron Scheduler, Background Tasks (advanced), Workflow Engine (advanced), RunLedger (advanced), Provenance (advanced)
4. **Payment & Account** — Payment, Smart Account (advanced), SA Session Keys (advanced), SA Paymaster (advanced), SA Modules (advanced)
5. **P2P & Economy** — P2P Network (advanced), P2P Workspace (advanced), P2P ZKP (advanced), P2P Pricing (advanced), P2P Owner Protection (advanced), P2P Sandbox (advanced), Economy (advanced), Economy Risk (advanced), Economy Negotiation (advanced), Economy Escrow (advanced), On-Chain Escrow (advanced), Economy Pricing (advanced)
6. **Integrations** — MCP Settings, MCP Server List (advanced), Observability (advanced), Alerting (advanced)
7. **Security** — Security, Auth (advanced), Security DB Encryption (advanced), Security KMS (advanced), OS Sandbox (advanced)
8. *(untitled)* — Save & Exit, Cancel

#### Scenario: Level 1 section list displayed
- **WHEN** user views the settings menu at Level 1
- **THEN** 7 named section items SHALL be rendered with category counts, followed by a separator and Save & Cancel items

#### Scenario: Level 2 categories displayed
- **WHEN** user selects a section at Level 1
- **THEN** the menu SHALL show only the categories belonging to that section, filtered by the current Basic/Advanced toggle

#### Scenario: Hidden categories in basic mode at Level 2
- **WHEN** settings menu is at Level 2 with basic mode and the section has only advanced categories
- **THEN** a "No basic settings. Press Tab to show all." message SHALL be displayed

### Requirement: Keyword Search
The settings menu SHALL support real-time keyword search to filter categories.

#### Scenario: Activate search
- **WHEN** user presses `/` in normal mode
- **THEN** the menu SHALL enter search mode, display a focused text input with `/ ` prompt and "Type to search..." placeholder, and reset the cursor to 0

#### Scenario: Filter categories
- **WHEN** user types a search query
- **THEN** the menu SHALL filter categories by case-insensitive substring match against title, description, and ID, updating results in real-time

#### Scenario: Empty search query
- **WHEN** the search input is empty or whitespace-only
- **THEN** all categories SHALL be displayed (no filtering)

#### Scenario: No results
- **WHEN** the search query matches no categories
- **THEN** the menu SHALL display "No matching items" in muted italic text

#### Scenario: Select from search results
- **WHEN** user presses Enter during search mode
- **THEN** the selected filtered category SHALL be activated, search mode SHALL exit, and the search input SHALL be cleared

#### Scenario: Cancel search
- **WHEN** user presses Esc during search mode
- **THEN** search mode SHALL be cancelled, the filtered list SHALL be cleared, and the full grouped menu SHALL be restored

#### Scenario: Navigate search results
- **WHEN** user presses up/down (or shift+tab/tab) during search mode
- **THEN** the cursor SHALL move within the filtered results list

### Requirement: Search Match Highlighting
The settings menu SHALL highlight matching substrings in search results.

#### Scenario: Highlight matching text
- **WHEN** categories are displayed during an active search with a non-empty query
- **THEN** the first matching substring in each category's title and description SHALL be rendered in amber/warning color with bold styling

#### Scenario: Selected item highlight
- **WHEN** the cursor is on a filtered category during search
- **THEN** the matching substring SHALL additionally be underlined

### Requirement: Search Help Bar
The help bar SHALL update based on the current mode and navigation level.

#### Scenario: Level 1 help bar
- **WHEN** the menu is at Level 1 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search (`/`), Back (`Esc`)

#### Scenario: Level 2 help bar
- **WHEN** the menu is at Level 2 in normal mode
- **THEN** the help bar SHALL display: Navigate, Select, Search (`/`), Tab (filter label), Back (`Esc`)

#### Scenario: Search mode help bar
- **WHEN** the menu is in search mode
- **THEN** the help bar SHALL display: Navigate, Select, Cancel (`Esc`)

### Requirement: Breadcrumb navigation in settings editor
The settings editor SHALL display a breadcrumb navigation header that reflects the current editor step. The breadcrumb SHALL use `tui.Breadcrumb()` with the following segments per step:
- **StepWelcome / StepMenu (Level 1)**: "Settings"
- **StepMenu (Level 2)**: "Settings" > section title (from `menu.ActiveSectionTitle()`)
- **StepForm**: "Settings" > form title (from `activeForm.Title`)
- **StepProvidersList**: "Settings" > "Providers"
- **StepAuthProvidersList**: "Settings" > "Auth Providers"
- **StepMCPServersList**: "Settings" > "MCP Servers"

The last breadcrumb segment SHALL be rendered in `Primary` color with bold weight. Preceding segments SHALL be rendered in `Muted` color. Segments SHALL be separated by " > " in `Dim` color.

#### Scenario: Breadcrumb at Level 1
- **WHEN** user is at StepMenu Level 1
- **THEN** the breadcrumb SHALL display "Settings" as a single segment

#### Scenario: Breadcrumb at Level 2
- **WHEN** user is at StepMenu Level 2 for the "Core" section
- **THEN** the breadcrumb SHALL display "Settings > Core"

#### Scenario: Breadcrumb at form
- **WHEN** user is editing the Agent form (StepForm)
- **THEN** the breadcrumb SHALL display "Settings > Agent Configuration"

#### Scenario: Breadcrumb at providers list
- **WHEN** user is at StepProvidersList
- **THEN** the breadcrumb SHALL display "Settings > Providers"

### Requirement: Styled containers for menu and list views
The settings menu body, providers list body, and auth providers list body SHALL each be wrapped in a `lipgloss.RoundedBorder()` container with `tui.Muted` border color and padding `(0, 1)`. The welcome screen SHALL be wrapped in a `lipgloss.RoundedBorder()` container with `tui.Primary` border color and padding `(1, 3)`.

#### Scenario: Menu container
- **WHEN** user is at StepMenu
- **THEN** the menu items SHALL be rendered inside a rounded-border container

#### Scenario: Welcome container
- **WHEN** user is at StepWelcome
- **THEN** the welcome message SHALL be rendered inside a primary-colored rounded-border box

### Requirement: Help bars in all interactive views
Every interactive settings view SHALL display a help bar at the bottom using `tui.HelpBar()` with `tui.HelpEntry()` badges. The help bars SHALL contain:
- **Welcome**: Enter (Start), Esc (Quit)
- **Menu (normal)**: up/down (Navigate), Enter (Select), / (Search), Esc (Back)
- **Menu (searching)**: up/down (Navigate), Enter (Select), Esc (Cancel)
- **Providers list**: up/down (Navigate), Enter (Select), d (Delete), Esc (Back)
- **Auth providers list**: up/down (Navigate), Enter (Select), d (Delete), Esc (Back)

#### Scenario: Menu help bar in normal mode
- **WHEN** user is at StepMenu in normal mode (not searching)
- **THEN** the help bar SHALL show Navigate, Select, Search, and Back entries

#### Scenario: Menu help bar in search mode
- **WHEN** user is at StepMenu in search mode
- **THEN** the help bar SHALL show Navigate, Select, and Cancel entries

### Requirement: Design system tokens in tui package
The `internal/cli/tui/styles.go` file SHALL export the following design tokens:
- **Colors**: `Primary` (#7C3AED), `Success` (#10B981), `Warning` (#F59E0B), `Error` (#EF4444), `Muted` (#6B7280), `Foreground` (#F9FAFB), `Background` (#1F2937), `Highlight` (#3B82F6), `Accent` (#04B575), `Dim` (#626262), `Separator` (#374151)
- **Styles**: `TitleStyle`, `SubtitleStyle`, `SuccessStyle`, `WarningStyle`, `ErrorStyle`, `MutedStyle`, `HighlightStyle`, `BoxStyle`, `ListItemStyle`, `SelectedItemStyle`, `SectionHeaderStyle`, `SeparatorLineStyle`, `CursorStyle`, `ActiveItemStyle`, `SearchBarStyle`, `FormTitleBarStyle`, `FieldDescStyle`
- **Functions**: `Breadcrumb(segments ...string)`, `HelpEntry(key, label string)`, `HelpBar(entries ...string)`, `KeyBadge(key string)`, `FormatPass(msg)`, `FormatWarn(msg)`, `FormatFail(msg)`, `FormatMuted(msg)`

#### Scenario: Breadcrumb rendering
- **WHEN** `tui.Breadcrumb("Settings", "Agent")` is called
- **THEN** the result SHALL be "Settings" in muted color, " > " separator in dim color, and "Agent" in primary bold

#### Scenario: HelpEntry rendering
- **WHEN** `tui.HelpEntry("Esc", "Back")` is called
- **THEN** the result SHALL be a key badge with "Esc" followed by "Back" label in dim color

### Requirement: Inline field descriptions
All settings form fields SHALL include a `Description` string providing human-readable guidance. The description SHALL be shown only when the field is focused.

#### Scenario: Description displayed on focus
- **WHEN** the user navigates to a field with a Description
- **THEN** the form SHALL render the description text below that field

#### Scenario: Description hidden when not focused
- **WHEN** the user moves focus away from a field
- **THEN** the description for that field SHALL no longer be rendered

### Requirement: Field input validation
Numeric and range-sensitive fields SHALL have `Validate` functions that return clear error messages.

#### Scenario: Temperature validation
- **WHEN** the user enters a value outside 0.0-2.0 for the Temperature field
- **THEN** the validator SHALL return "must be between 0.0 and 2.0"

#### Scenario: Port validation
- **WHEN** the user enters a value outside 1-65535 for the Port field
- **THEN** the validator SHALL return "port out of range"

#### Scenario: Positive integer validation
- **WHEN** the user enters a non-positive value for fields requiring positive integers (Max Read Size, Max History Turns, Knowledge Max Context, Max Concurrent Jobs, Max Concurrent Tasks, Max Concurrent Steps, Max Peers, Observation Threshold, Max Bulk Import, Import Concurrency)
- **THEN** the validator SHALL return "must be a positive integer"

#### Scenario: Non-negative integer validation
- **WHEN** the user enters a negative value for fields allowing zero (Yield Time, Max Reflections in Context, Max Observations in Context, Inquiry Cooldown, Max Pending Inquiries, Approval Timeout, Embedding Dimensions, RAG Max Results)
- **THEN** the validator SHALL return "must be a non-negative integer" (with optional "(0 = unlimited)" suffix where applicable)

#### Scenario: Float range validation
- **WHEN** the user enters a value outside 0.0-1.0 for Min Trust Score
- **THEN** the validator SHALL return "must be between 0.0 and 1.0"

### Requirement: Auto-fetch model options from provider API
Form builders for Agent, Observational Memory, Embedding, and Librarian SHALL attempt to fetch available models from the configured provider API at form creation time.

#### Scenario: Successful model fetch
- **WHEN** the provider API returns a list of models within the 15-second timeout
- **THEN** the model field SHALL be converted from InputText to InputSearchSelect with the fetched models as options, and the current model SHALL always be included

#### Scenario: Failed model fetch with error feedback
- **WHEN** the provider API fails, times out, or returns empty
- **THEN** the model field SHALL remain as InputText and the description SHALL show the failure reason

#### Scenario: Embedding model field with filtered models
- **WHEN** the Embedding form fetches models
- **THEN** FetchEmbeddingModelOptions SHALL filter for embedding-pattern models ("embed", "embedding") and fall back to full list if no matches

#### Scenario: Esc key with open dropdown in form
- **WHEN** user presses Esc while a search-select dropdown is open in StepForm
- **THEN** editor passes Esc to form (closes dropdown) instead of exiting the form

#### Scenario: Agent form model fetch
- **WHEN** the Agent form is created and the configured provider has a valid API key
- **THEN** the Model ID field SHALL be populated with models from `FetchModelOptions(cfg.Agent.Provider, ...)`

#### Scenario: Observational Memory model fetch with provider inheritance
- **WHEN** the Observational Memory form is created with an empty provider
- **THEN** the model fetch SHALL use the Agent provider as fallback

#### Scenario: Librarian model fetch with provider inheritance
- **WHEN** the Librarian form is created with an empty provider
- **THEN** the model fetch SHALL use the Agent provider as fallback

#### Scenario: Embedding model fetch
- **WHEN** the Embedding form is created with a non-empty provider
- **THEN** the Model field SHALL attempt to fetch models from the embedding provider

### Requirement: Agent form reactive model list
The Agent configuration form SHALL wire `OnChange` on the provider field to asynchronously fetch and update the model field's options when the provider changes.

#### Scenario: Provider change triggers model refresh
- **WHEN** a user changes the provider field in the Agent form
- **THEN** the model field SHALL show a loading indicator and asynchronously fetch models from the new provider

#### Scenario: Fallback provider change triggers fallback model refresh
- **WHEN** a user changes the fallback provider field in the Agent form
- **THEN** the fallback model field SHALL asynchronously refresh its options from the new fallback provider

### Requirement: Knowledge forms reactive model list
The Observational Memory, Embedding, and Librarian configuration forms SHALL wire `OnChange` on their provider fields to refresh the corresponding model field options.

#### Scenario: OM provider change triggers OM model refresh
- **WHEN** a user changes the OM provider field
- **THEN** the OM model field SHALL asynchronously fetch models from the new provider (or agent provider if empty)

#### Scenario: Embedding provider change triggers embedding model refresh
- **WHEN** a user changes the embedding provider field
- **THEN** the embedding model field SHALL asynchronously fetch embedding-filtered models

#### Scenario: Librarian provider change triggers librarian model refresh
- **WHEN** a user changes the librarian provider field
- **THEN** the librarian model field SHALL asynchronously fetch models from the new provider (or agent provider if empty)

### Requirement: Async Cmd wrappers for model fetching
The settings package SHALL provide `FetchModelOptionsCmd()` and `FetchEmbeddingModelOptionsCmd()` functions that return `tea.Cmd` for async model fetching, producing `FieldOptionsLoadedMsg` results.

#### Scenario: FetchModelOptionsCmd returns loaded message
- **WHEN** `FetchModelOptionsCmd("model", "openai", cfg, "")` is executed
- **THEN** it SHALL return a `FieldOptionsLoadedMsg` with `FieldKey="model"` and `ProviderID="openai"`

### Requirement: Unified embedding provider field
The Embedding & RAG form SHALL use a single "Provider" field (key `emb_provider_id`) mapped to `cfg.Embedding.Provider`. The state update handler SHALL clear the deprecated `cfg.Embedding.ProviderID` field when saving.

#### Scenario: Embedding form shows single provider field
- **WHEN** the user opens the Embedding & RAG form
- **THEN** the form SHALL display one "Provider" select field, not separate Provider and ProviderID fields

#### Scenario: State update clears deprecated ProviderID
- **WHEN** the `emb_provider_id` field is saved via UpdateConfigFromForm
- **THEN** `cfg.Embedding.Provider` SHALL be set to the value AND `cfg.Embedding.ProviderID` SHALL be set to empty string

### Requirement: Conditional field visibility in channel forms
Channel token fields SHALL be visible only when the parent channel is enabled.

#### Scenario: Telegram token hidden when disabled
- **WHEN** the Telegram Enabled toggle is unchecked
- **THEN** the Telegram Bot Token field SHALL be hidden

#### Scenario: Telegram token shown when enabled
- **WHEN** the user checks the Telegram Enabled toggle
- **THEN** the Telegram Bot Token field SHALL become visible

#### Scenario: Discord token visibility
- **WHEN** the Discord Enabled toggle is toggled
- **THEN** the Discord Bot Token field visibility SHALL match the toggle state

#### Scenario: Slack token visibility
- **WHEN** the Slack Enabled toggle is toggled
- **THEN** the Slack Bot Token and App Token fields visibility SHALL match the toggle state

### Requirement: Conditional visibility in security form
Security sub-fields SHALL be visible only when their parent toggle is enabled.

#### Scenario: PII fields hidden when interceptor disabled
- **WHEN** the Privacy Interceptor toggle is unchecked
- **THEN** all interceptor sub-fields (Redact PII, Approval Policy, Timeout, Notify Channel, Sensitive Tools, Exempt Tools, Disabled PII Patterns, Custom PII Patterns, Presidio) SHALL be hidden

#### Scenario: Presidio detail fields nested under both interceptor and presidio
- **WHEN** the interceptor is enabled but Presidio is disabled
- **THEN** the Presidio URL and Presidio Language fields SHALL be hidden

#### Scenario: Presidio fields visible when both enabled
- **WHEN** both the Privacy Interceptor and Presidio toggles are checked
- **THEN** the Presidio URL and Presidio Language fields SHALL be visible

#### Scenario: Signer Key ID visibility based on provider
- **WHEN** the signer provider is "local" or "enclave"
- **THEN** the Key ID field SHALL be hidden

#### Scenario: Signer RPC URL visibility
- **WHEN** the signer provider is "rpc"
- **THEN** the RPC URL field SHALL be visible

### Requirement: Conditional visibility in P2P sandbox form
P2P container sandbox fields SHALL be visible only when the container sandbox is enabled.

#### Scenario: Container fields hidden when container disabled
- **WHEN** the Container Sandbox Enabled toggle is unchecked
- **THEN** container-specific fields (Runtime, Image, Network Mode, Read-Only RootFS, CPU Quota, Pool Size, Pool Idle Timeout) SHALL be hidden

### Requirement: Conditional visibility in KMS form
KMS backend-specific fields SHALL be visible based on the selected backend type.

#### Scenario: Azure fields visible for azure-kv backend
- **WHEN** the KMS backend is "azure-kv"
- **THEN** the Azure Vault URL and Azure Key Version fields SHALL be visible

#### Scenario: PKCS11 fields visible for pkcs11 backend
- **WHEN** the KMS backend is "pkcs11"
- **THEN** the PKCS11 Module Path, Slot ID, PIN, and Key Label fields SHALL be visible

### Requirement: Model Fetcher API
The settings package SHALL export `FetchModelOptions` and `NewProviderFromConfig` as public functions so other CLI packages (e.g., onboard) can reuse model auto-fetch logic.

#### Scenario: Exported function availability
- **WHEN** another package imports the settings package
- **THEN** `settings.FetchModelOptions(providerID, cfg, currentModel)` SHALL be callable
- **AND** `settings.NewProviderFromConfig(id, pCfg)` SHALL be callable

### Requirement: Model fetcher provider support
The `NewProviderFromConfig` function SHALL support creating lightweight provider instances for: OpenAI, Anthropic, Gemini/Google, Ollama (via OpenAI-compatible endpoint), and GitHub (via OpenAI-compatible endpoint).

#### Scenario: Ollama default base URL
- **WHEN** creating an Ollama provider with empty BaseURL
- **THEN** the base URL SHALL default to "http://localhost:11434/v1"

#### Scenario: GitHub default base URL
- **WHEN** creating a GitHub provider with empty BaseURL
- **THEN** the base URL SHALL default to "https://models.inference.ai.azure.com"

#### Scenario: Provider without API key
- **WHEN** creating a non-Ollama provider with empty API key
- **THEN** `NewProviderFromConfig` SHALL return nil

### Requirement: Economy settings forms
The settings TUI SHALL provide 5 Economy configuration forms:
- `NewEconomyForm(cfg)` — economy.enabled, budget.defaultMax, budget.hardLimit, budget.alertThresholds
- `NewEconomyRiskForm(cfg)` — risk.escrowThreshold, risk.highTrustScore, risk.mediumTrustScore
- `NewEconomyNegotiationForm(cfg)` — negotiate.enabled, maxRounds, timeout, autoNegotiate, maxDiscount
- `NewEconomyEscrowForm(cfg)` — escrow.enabled, defaultTimeout, maxMilestones, autoRelease, disputeWindow
- `NewEconomyPricingForm(cfg)` — pricing.enabled, trustDiscount, volumeDiscount, minPrice

#### Scenario: User edits economy base settings
- **WHEN** user selects "Economy" from the settings menu
- **THEN** the editor SHALL display a form with Enabled toggle, Budget Default Max, Hard Limit, and Alert Thresholds fields pre-populated from `config.Economy`

#### Scenario: User edits economy risk settings
- **WHEN** user selects "Economy Risk" from the settings menu
- **THEN** the editor SHALL display a form with escrow threshold, high trust score, and medium trust score fields

#### Scenario: User edits economy negotiation settings
- **WHEN** user selects "Economy Negotiation" from the settings menu
- **THEN** the editor SHALL display a form with enabled toggle, max rounds, timeout, auto-negotiate, and max discount fields

#### Scenario: User edits economy escrow settings
- **WHEN** user selects "Economy Escrow" from the settings menu
- **THEN** the editor SHALL display a form with enabled toggle, default timeout, max milestones, auto-release, and dispute window fields

#### Scenario: User edits economy pricing settings
- **WHEN** user selects "Economy Pricing" from the settings menu
- **THEN** the editor SHALL display a form with enabled toggle, trust discount, volume discount, and min price fields

### Requirement: Observability settings form
The settings TUI SHALL provide an Observability configuration form with fields for observability.enabled, tokens (enabled, persistHistory, retentionDays), health (enabled, interval), audit (enabled, retentionDays), and metrics (enabled, format).

#### Scenario: User edits observability settings
- **WHEN** user selects "Observability" from the settings menu
- **THEN** the editor SHALL display a form with all observability fields pre-populated from `config.Observability`

### Requirement: Economy and observability state update
The `UpdateConfigFromForm()` function SHALL handle all economy and observability form field keys, mapping them to the corresponding config struct fields.

#### Scenario: Economy form fields saved
- **WHEN** user edits economy form fields and navigates back
- **THEN** the config state SHALL be updated for all economy.* fields including budget, risk, negotiation, escrow, and pricing sub-configs

#### Scenario: Observability form fields saved
- **WHEN** user edits observability form fields and navigates back
- **THEN** the config state SHALL be updated for all observability.* fields including tokens, health, audit, and metrics sub-configs

### Requirement: TUI on-chain escrow form
The system SHALL provide a TUI form (`NewEconomyEscrowOnChainForm`) for configuring on-chain escrow settings with 10 fields: enabled, mode, hubAddress, vaultFactoryAddress, vaultImplementation, arbitratorAddress, tokenAddress, pollInterval, receiptTimeout, maxRetries.

#### Scenario: Form creation
- **WHEN** the user selects "On-Chain Escrow" from the settings menu
- **THEN** a form with 10 fields matching `EscrowOnChainConfig` and `EscrowSettlementConfig` is displayed

#### Scenario: Mode validation
- **WHEN** the user enters a value other than "hub" or "vault" for the mode field
- **THEN** a validation error "must be 'hub' or 'vault'" is shown

#### Scenario: Max retries validation
- **WHEN** the user enters a negative number for max retries
- **THEN** a validation error "must be a non-negative integer" is shown

### Requirement: Menu category for on-chain escrow
The system SHALL include an `economy_escrow_onchain` category in the Economy section of the settings menu with title "On-Chain Escrow" and description "Hub/Vault mode, contracts, settlement".

#### Scenario: Menu navigation
- **WHEN** the user navigates the settings menu to the Economy section
- **THEN** "On-Chain Escrow" appears as a selectable category

### Requirement: Editor wiring for on-chain escrow
The system SHALL wire the `economy_escrow_onchain` menu selection to the `NewEconomyEscrowOnChainForm` in `editor.go`.

#### Scenario: Menu selection handler
- **WHEN** the user selects `economy_escrow_onchain` from the menu
- **THEN** `handleMenuSelection` returns the on-chain escrow form model

### Requirement: RunLedger configuration form
The settings editor SHALL provide a RunLedger configuration form with the following fields:

- **Enabled** (`runledger_enabled`) — Boolean toggle
- **Shadow Mode** (`runledger_shadow`) — Boolean toggle
- **Write-Through** (`runledger_write_through`) — Boolean toggle
- **Authoritative Read** (`runledger_authoritative_read`) — Boolean toggle
- **Workspace Isolation** (`runledger_workspace_isolation`) — Boolean toggle
- **Stale TTL** (`runledger_stale_ttl`) — Duration text input
- **Max Run History** (`runledger_max_history`) — Integer input
- **Validator Timeout** (`runledger_validator_timeout`) — Duration text input
- **Planner Max Retries** (`runledger_planner_retries`) — Integer input

#### Scenario: Edit RunLedger settings
- **WHEN** user selects `RunLedger` from the settings menu
- **THEN** the editor SHALL display a form with all RunLedger fields pre-populated from `config.RunLedger`

#### Scenario: Save RunLedger settings
- **WHEN** user edits RunLedger fields and navigates back or saves
- **THEN** the config state SHALL be updated through `UpdateConfigFromForm`
- **AND** all edited values SHALL persist into `config.RunLedger`

### Requirement: Provenance configuration form
The settings editor SHALL provide a Provenance configuration form with the following fields:

- **Enabled** (`provenance_enabled`) — Boolean toggle
- **Auto on Step Complete** (`provenance_auto_on_step_complete`) — Boolean toggle
- **Auto on Policy** (`provenance_auto_on_policy`) — Boolean toggle
- **Max Per Session** (`provenance_max_per_session`) — Integer input
- **Retention Days** (`provenance_retention_days`) — Integer input

#### Scenario: Edit Provenance settings
- **WHEN** user selects `Provenance` from the settings menu
- **THEN** the editor SHALL display a form with all provenance fields pre-populated from `config.Provenance`

#### Scenario: Save Provenance settings
- **WHEN** user edits Provenance fields and navigates back or saves
- **THEN** the config state SHALL be updated through `UpdateConfigFromForm`
- **AND** all edited values SHALL persist into `config.Provenance`

### Requirement: Orchestration state update mapping
The `UpdateConfigFromForm()` function SHALL handle orchestration form field keys, mapping them to `Config.Agent.Orchestration` sub-fields: `orchestration_mode` → `Mode`, `orc_cb_failure_threshold` → `CircuitBreaker.FailureThreshold`, `orc_cb_reset_timeout` → `CircuitBreaker.ResetTimeout`, `orc_budget_tool_call_limit` → `Budget.ToolCallLimit`, `orc_budget_delegation_limit` → `Budget.DelegationLimit`, `orc_budget_alert_threshold` → `Budget.AlertThreshold`, `orc_recovery_max_retries` → `Recovery.MaxRetries`, `orc_recovery_cooldown` → `Recovery.CircuitBreakerCooldown`.

#### Scenario: Orchestration fields saved to config
- **WHEN** user sets `orchestration_mode=structured`, `orc_cb_failure_threshold=5`, `orc_budget_alert_threshold=0.75`
- **THEN** `Config.Agent.Orchestration.Mode` SHALL be "structured", `CircuitBreaker.FailureThreshold` SHALL be 5, `Budget.AlertThreshold` SHALL be 0.75

#### Scenario: Invalid orchestration values ignored
- **WHEN** user enters "not-a-number" for `orc_cb_failure_threshold`
- **THEN** the config value SHALL remain unchanged (parse error silently skipped)

### Requirement: Trace store state update mapping
The `UpdateConfigFromForm()` function SHALL handle trace store form field keys, mapping them to `Config.Observability.TraceStore` sub-fields: `obs_trace_max_age` → `MaxAge`, `obs_trace_max_traces` → `MaxTraces`, `obs_trace_failed_multiplier` → `FailedTraceMultiplier`, `obs_trace_cleanup_interval` → `CleanupInterval`.

#### Scenario: Trace store fields saved to config
- **WHEN** user sets `obs_trace_max_age=168h`, `obs_trace_max_traces=5000`
- **THEN** `Config.Observability.TraceStore.MaxAge` SHALL be 168h, `MaxTraces` SHALL be 5000

#### Scenario: Duration parse for trace store
- **WHEN** user enters "30m" for `obs_trace_cleanup_interval`
- **THEN** `Config.Observability.TraceStore.CleanupInterval` SHALL be 30 minutes

### Requirement: OS Sandbox settings form
The settings TUI SHALL provide an "OS Sandbox" category under the Security section with 9 fields mapping to `cfg.Sandbox.*`, using `os_sandbox_*` field key prefix.

#### Scenario: Form contains all sandbox config fields
- **WHEN** `NewOSSandboxForm(cfg)` is called
- **THEN** the form SHALL contain 9 fields: os_sandbox_enabled, os_sandbox_fail_closed, os_sandbox_workspace_path, os_sandbox_network_mode, os_sandbox_allowed_ips, os_sandbox_allowed_write_paths, os_sandbox_timeout, os_sandbox_seccomp_profile, os_sandbox_seatbelt_profile

#### Scenario: OS sandbox fields do not affect P2P sandbox config
- **WHEN** `os_sandbox_enabled` is toggled in the form
- **THEN** `cfg.Sandbox.Enabled` SHALL change and `cfg.P2P.ToolIsolation.Enabled` SHALL NOT change

#### Scenario: Menu includes OS Sandbox category
- **WHEN** the settings menu is rendered
- **THEN** the Security section SHALL contain an "OS Sandbox" entry with ID `os_sandbox`

#### Scenario: OS Sandbox category is always enabled
- **WHEN** `categoryIsEnabled("os_sandbox")` is called
- **THEN** it SHALL return true regardless of other config settings
