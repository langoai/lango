## Purpose

Capability spec for oauth-login. See requirements below for scope and behavior contracts.

## REMOVED

**Status**: Removed (2026-02-14)
**Reason**: OAuth with AI providers risks account bans. Use API key authentication instead.

**Migration**: Remove `clientId`, `clientSecret`, and `scopes` from provider configuration. Set `apiKey` using environment variable references (e.g., `${GOOGLE_API_KEY}`).

---

## Requirements

### Requirement: OAuth Login Command
The system previously provided a CLI command `lango login [provider]` to initiate OAuth authentication with Google and GitHub.

### Requirement: OAuth Callback Handling
The system previously handled OAuth callbacks on localhost.

### Requirement: Secure Token Storage
The system previously stored OAuth tokens in `~/.lango/tokens/<provider>.json`.

### Requirement: Automatic Token Refresh
The system previously refreshed expired access tokens automatically.
