## ADDED Requirements

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

## MODIFIED Requirements

### Configuration Coverage
The settings editor SHALL support editing all configuration sections previously handled by `lango onboard`:
1. **Providers** — Add, edit, delete multi-provider configurations
2. **Agent** — Provider, Model, MaxTokens, Temperature, PromptsDir, Fallback
3. **Server** — Host, Port, HTTP/WebSocket toggles
4. **Channels** — Telegram, Discord, Slack enable/disable + tokens
5. **Tools** — Exec timeout, Browser, Filesystem limits
6. **Session** — Database path, TTL, Max history turns
7. **Security** — Interceptor (PII, policy, timeout, tools), Signer (provider, RPC, KeyID)
8. **Auth** — OIDC provider management (add, edit, delete)
9. **Knowledge** — Enabled, max learnings/knowledge/context
10. **Skill** — Enabled, skills directory
11. **Observational Memory** — Enabled, provider, model, thresholds, budget
12. **Embedding & RAG** — Provider, model, dimensions, local URL, RAG settings
13. **Graph Store** — Enabled, backend, DB path, traversal depth, expansion results
14. **Multi-Agent** — Orchestration toggle
15. **A2A Protocol** — Enabled, base URL, agent name/description
16. **Payment** — Wallet, chain ID, RPC URL, USDC contract, limits, X402
17. **Cron Scheduler** — Enabled, timezone, max concurrent jobs, session mode, history retention
18. **Background Tasks** — Enabled, yield time, max concurrent tasks
19. **Workflow Engine** — Enabled, max concurrent steps, default timeout, state directory
20. **Librarian** — Enabled, observation threshold, inquiry cooldown, max inquiries, auto-save confidence, provider, model

### User Interface
- Menu-based navigation with 22 categories (20 sections + Save & Exit + Cancel)
- Free navigation between categories
- Uses shared `tuicore.FormModel` for all forms
- Provider and OIDC provider list views for managing collections

#### Scenario: Menu categories
- **WHEN** user launches `lango settings`
- **THEN** the menu SHALL display categories in order: Providers, Agent, Server, Channels, Tools, Session, Security, Auth, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store, Multi-Agent, A2A Protocol, Payment, Cron Scheduler, Background Tasks, Workflow Engine, Librarian, Save & Exit, Cancel
