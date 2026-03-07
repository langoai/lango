## MODIFIED Requirements

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
21. **P2P Network** — Enabled, listen addrs, bootstrap peers, relay, mDNS, max peers, handshake timeout, session token TTL, auto-approve, gossip interval, ZK handshake/attestation, signed challenge, min trust score
22. **P2P ZKP** — Proof cache dir, proving scheme, SRS mode/path, max credential age
23. **P2P Pricing** — Enabled, per query price, tool-specific prices
24. **P2P Owner Protection** — Owner name/email/phone, extra terms, block conversations
25. **P2P Sandbox** — Tool isolation (enabled, timeout, memory), container sandbox (runtime, image, network, rootfs, CPU, pool)
26. **Security Keyring** — OS keyring enabled
27. **Security DB Encryption** — SQLCipher enabled, cipher page size
28. **Security KMS** — Region, key ID, endpoint, fallback, timeout, retries, Azure vault/version, PKCS#11 module/slot/PIN/key label
29. **Economy** — Enabled, budget (defaultMax, hardLimit, alertThresholds)
30. **Economy Risk** — Escrow threshold, high trust score, medium trust score
31. **Economy Negotiation** — Enabled, max rounds, timeout, auto-negotiate, max discount
32. **Economy Escrow** — Enabled, default timeout, max milestones, auto-release, dispute window
33. **Economy Pricing** — Enabled, trust discount, volume discount, min price
34. **Observability** — Enabled, tokens (enabled, persist, retention), health (enabled, interval), audit (enabled, retention), metrics (enabled, format)

#### Scenario: Menu categories
- **WHEN** user launches `lango settings`
- **THEN** the menu SHALL display all categories including Economy (5 sub-forms), Observability, grouped under "Economy" and "Infrastructure" sections respectively

#### Scenario: Provider form includes github
- **WHEN** user opens the provider add/edit form
- **THEN** the Type select field options SHALL include "github" alongside openai, anthropic, gemini, and ollama

### Requirement: Grouped Section Layout
The settings menu SHALL organize categories into named sections. Each section SHALL have a title header rendered above its categories with a visual separator line between sections.

The sections SHALL be, in order:
1. **Core** — Providers, Agent, Server, Session
2. **Communication** — Channels, Tools, Multi-Agent, A2A Protocol
3. **AI & Knowledge** — Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store, Librarian
4. **Economy** — Economy, Economy Risk, Economy Negotiation, Economy Escrow, Economy Pricing
5. **Infrastructure** — Payment, Cron Scheduler, Background Tasks, Workflow Engine, Observability
6. **P2P Network** — P2P Network, P2P ZKP, P2P Pricing, P2P Owner Protection, P2P Sandbox
7. **Security** — Security, Auth, Security Keyring, Security DB Encryption, Security KMS
8. *(untitled)* — Save & Exit, Cancel

#### Scenario: Section headers displayed
- **WHEN** user views the settings menu in normal (non-search) mode
- **THEN** named section headers SHALL be rendered above each group of categories with separator lines between sections

#### Scenario: Flat cursor across sections
- **WHEN** user navigates with arrow keys
- **THEN** the cursor SHALL move through all categories across sections as a flat list, skipping section headers

## ADDED Requirements

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
