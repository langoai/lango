## Purpose

Capability spec for oauth-login. See requirements below for scope and behavior contracts.

## REMOVED

**Status**: Removed (2026-02-14)
**Reason**: OAuth with AI providers risks account bans. Use API key authentication instead.

**Migration**: Remove `clientId`, `clientSecret`, and `scopes` from provider configuration. Set `apiKey` using environment variable references (e.g., `${GOOGLE_API_KEY}`).

---

## Requirements

### Requirement: OAuth Login Command
The system SHALL NOT provide a CLI command `lango login [provider]` for AI provider authentication.

#### Scenario: OAuth login command is unavailable
- **WHEN** a user attempts to use OAuth-based AI provider login
- **THEN** the system SHALL direct them to API-key based configuration instead

### Requirement: OAuth Callback Handling
The system SHALL NOT rely on localhost OAuth callback handling for AI provider authentication.

#### Scenario: OAuth callback flow is absent
- **WHEN** a user expects a localhost callback for AI provider login
- **THEN** the product SHALL not expose that callback flow

### Requirement: Secure Token Storage
The system SHALL NOT store AI provider OAuth tokens in `~/.lango/tokens/<provider>.json`.

#### Scenario: OAuth token file is not part of provider auth
- **WHEN** AI provider authentication is configured
- **THEN** the product SHALL use API-key based configuration rather than OAuth token files

### Requirement: Automatic Token Refresh
The system SHALL NOT implement automatic OAuth token refresh for AI providers.

#### Scenario: OAuth token refresh is not used
- **WHEN** AI provider authentication is evaluated
- **THEN** automatic OAuth token refresh SHALL not be part of the supported flow
