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

## ADDED Requirements

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
