# CLI Settings Spec

## Goal
The `lango settings` command provides a comprehensive, interactive menu-based configuration editor for all aspects of the encrypted configuration profile.

## Requirements

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
9. **Knowledge** — Enabled, max learnings/knowledge/context, auto-approve, max skills/day
10. **Observational Memory** — Enabled, provider, model, thresholds, budget
11. **Embedding & RAG** — Provider, model, dimensions, local URL, RAG settings
12. **Graph Store** — Enabled, backend, DB path, traversal depth, expansion results
13. **Multi-Agent** — Orchestration toggle
14. **A2A Protocol** — Enabled, base URL, agent name/description
15. **Payment** — Wallet, chain ID, RPC URL, USDC contract, limits, X402

### User Interface
- Menu-based navigation with 17 categories (15 sections + Save & Exit + Cancel)
- Free navigation between categories
- Uses shared `tuicore.FormModel` for all forms
- Provider and OIDC provider list views for managing collections

#### Scenario: Menu categories
- **WHEN** user launches `lango settings`
- **THEN** the menu SHALL display categories in order: Providers, Agent, Server, Channels, Tools, Session, Security, Auth, Knowledge, Observational Memory, Embedding & RAG, Graph Store, Multi-Agent, A2A Protocol, Payment, Save & Exit, Cancel

### Command
```
lango settings [--profile <name>]
```
- Default profile: "default"
- Loads existing profile or creates new with defaults
- Saves via `configstore.Store.Save()` to encrypted profile

#### Scenario: Launch settings
- **WHEN** user runs `lango settings`
- **THEN** the editor SHALL display a welcome screen followed by the configuration menu

#### Scenario: Save from settings
- **WHEN** user selects "Save & Exit" from the menu
- **THEN** the configuration SHALL be saved as an encrypted profile
