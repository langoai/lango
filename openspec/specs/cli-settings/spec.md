## Purpose

Define the `lango settings` command that provides a comprehensive, interactive menu-based configuration editor for all aspects of the encrypted configuration profile.

## Requirements

### Requirement: Configuration Coverage
The settings editor SHALL support editing all configuration sections:
1. **Providers** — Add, edit, delete multi-provider configurations
2. **Agent** — Provider, Model, MaxTokens, Temperature, PromptsDir, Fallback
3. **Server** — Host, Port, HTTP/WebSocket toggles
4. **Channels** — Telegram, Discord, Slack enable/disable + tokens
5. **Tools** — Exec timeout, Browser, Filesystem limits
6. **Session** — TTL, Max history turns
7. **Security** — Interceptor (PII, policy, timeout, tools), Signer (provider incl. aws-kms/gcp-kms/azure-kv/pkcs11, RPC, KeyID)
8. **Auth** — OIDC provider management (add, edit, delete)
9. **Knowledge** — Enabled, max context per layer, auto approve skills, max skills per day
10. **Skill** — Enabled, skills directory
11. **Observational Memory** — Enabled, provider, model, thresholds, budget, context limits
12. **Embedding & RAG** — Provider, model, dimensions, local URL, RAG settings
13. **Graph Store** — Enabled, backend, DB path, traversal depth, expansion results
14. **Multi-Agent** — Orchestration toggle
15. **A2A Protocol** — Enabled, base URL, agent name/description
16. **Payment** — Wallet, chain ID, RPC URL, USDC contract, limits, X402
17. **Cron Scheduler** — Enabled, timezone, max concurrent jobs, session mode, history retention
18. **Background Tasks** — Enabled, yield time, max concurrent tasks
19. **Workflow Engine** — Enabled, max concurrent steps, default timeout, state directory
20. **Librarian** — Enabled, observation threshold, inquiry cooldown, max inquiries, auto-save confidence, provider, model
21. **P2P Network** — Enabled, listen addrs, bootstrap peers, relay, mDNS, max peers, handshake, gossip, ZK settings, trust score
22. **P2P ZKP** — Proof cache, proving scheme, SRS mode/path, credential age
23. **P2P Pricing** — Enabled, per query price, tool-specific prices
24. **P2P Owner Protection** — Owner name/email/phone, extra terms, block conversations
25. **P2P Sandbox** — Tool isolation (enabled, timeout, memory), container sandbox (runtime, image, network, rootfs, pool)
26. **Security Keyring** — OS keyring enabled
27. **Security DB Encryption** — SQLCipher enabled, cipher page size
28. **Security KMS** — Region, key ID, endpoint, fallback, timeout, retries, Azure, PKCS#11

#### Scenario: Menu categories
- **WHEN** user launches `lango settings`
- **THEN** the menu SHALL display categories in order: Providers, Agent, Server, Channels, Tools, Session, Security, Auth, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store, Multi-Agent, A2A Protocol, Payment, Cron Scheduler, Background Tasks, Workflow Engine, Librarian, P2P Network, P2P ZKP, P2P Pricing, P2P Owner Protection, P2P Sandbox, Security Keyring, Security DB Encryption, Security KMS, Save & Exit, Cancel

### Requirement: User Interface
The settings editor SHALL provide menu-based navigation with categories, free navigation between categories, and shared `tuicore.FormModel` for all forms. Provider and OIDC provider list views SHALL support managing collections.

#### Scenario: Launch settings
- **WHEN** user runs `lango settings`
- **THEN** the editor SHALL display a welcome screen followed by the configuration menu

#### Scenario: Save from settings
- **WHEN** user selects "Save & Exit" from the menu
- **THEN** the configuration SHALL be saved as an encrypted profile

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
