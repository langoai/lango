## Purpose

Capability spec for phantom-feature-wiring. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: phantom-feature-wiring capability documented
The phantom-feature-wiring capability SHALL be documented through the sections in this spec. This requirement is a structural placeholder that satisfies the canonical openspec format; detailed behavior contracts live in the descriptive sections of this file.

#### Scenario: Spec file is readable
- **WHEN** the phantom-feature-wiring spec.md file is read
- **THEN** it SHALL describe the capability's behavior in sections below

# Phantom Feature Wiring

All config-exposed settings must have corresponding runtime behavior. No config field should exist without implementation.

## Agent Config Wiring

### SystemPromptPath
- When `agent.systemPromptPath` is set to a file path, the system prompt is loaded from that file via `os.ReadFile()`
- If the file does not exist or is empty, falls back to the hardcoded default prompt
- The loaded prompt is passed to both `NewContextAwareModelAdapter()` and `NewAgent()`

### Temperature / MaxTokens
- `agent.temperature` and `agent.maxTokens` are passed to `ProviderProxy` via `WithTemperature()` and `WithMaxTokens()` options
- In `Generate()`, if the request params have zero values, the proxy's configured defaults are applied
- Values from the ADK request config (if set) take precedence over proxy defaults

### FallbackProvider / FallbackModel
- `agent.fallbackProvider` and `agent.fallbackModel` are passed via `WithFallback()` option
- When the primary provider returns an error in `Generate()`, the proxy attempts one retry with the fallback provider
- If fallback also fails, the fallback error is returned (wrapped)
- If no fallback is configured, the primary error is returned directly

## Session Store Wiring

### MaxHistoryTurns
- `session.maxHistoryTurns` is passed via `WithMaxHistoryTurns()` store option
- After each `AppendMessage()`, if the count exceeds the limit, the oldest messages are deleted
- Deletion is ordered by timestamp ascending

### Session TTL
- `session.ttl` is passed via `WithTTL()` store option
- On `Get()`, if `time.Since(session.UpdatedAt) > ttl`, returns an "expired" error
- Writing to an expired session is still allowed (refreshes UpdatedAt)

## Gateway Wiring

### HTTPEnabled
- When `server.httpEnabled` is `false`, the `/health` and `/status` HTTP endpoints are not registered
- WebSocket endpoints are controlled separately by `server.wsEnabled`

### Auth Routes
- When `auth.providers` contains OIDC provider configs, `AuthManager` is created and passed to Gateway
- Gateway registers `/auth/login/{provider}` and `/auth/callback/{provider}` routes when `auth != nil`

## Security Tool Wiring

### Crypto Tools
- When security is initialized (`app.Crypto != nil`), 5 crypto tools are registered: `crypto_encrypt`, `crypto_decrypt`, `crypto_sign`, `crypto_hash`, `crypto_keys`
- Tools delegate to `tools/crypto.Tool` methods

### Secrets Tools
- When security is initialized (`app.Secrets != nil`), 4 secrets tools are registered: `secrets_store`, `secrets_get`, `secrets_list`, `secrets_delete`
- Tools delegate to `tools/secrets.Tool` methods

### Tool Approval
- When `security.interceptor.approvalRequired` is `true` and `sensitiveTools` is non-empty, all tools are wrapped with approval logic
- Sensitive tools check `Gateway.HasCompanions()`:
  - If companion connected: calls `Gateway.RequestApproval()` and waits for response (30s timeout)
  - If no companion: logs warning and proceeds (fail-open)
- Non-sensitive tools pass through unwrapped

## Knowledge Wiring

### MaxSkillsPerDay
- `knowledge.maxSkillsPerDay` is passed to `knowledge.NewStore()` constructor
- `SaveSkill()` checks a daily counter keyed by date string (`YYYY-MM-DD`)
- Returns error when the daily limit is reached

### Learning SessionKey
- Session key is stored in context via `WithSessionKey()` / `SessionKeyFromContext()`
- `wrapWithLearning()` extracts session key from context instead of passing empty string
- Enables per-session learning tracking

## CLI Wiring

### Security Command
- `lango security migrate-passphrase` is registered in main.go
- Uses lazy config loading: `NewSecurityCmd(func() (*config.Config, error))` accepts a closure
- Config is loaded at command execution time, not at registration time

## Dead Code

### Companion Discovery (deleted)
- `internal/companion/discovery.go` (254 LOC) deleted
- mDNS service discovery was never imported; WebSocket companion connections use Gateway directly
- `zeroconf` dependency removed via `go mod tidy`

### PairingEnabled (removed)
- `TelegramConfig.PairingEnabled` field removed from config, telegram, and channels packages
- No code ever read or acted on this field
