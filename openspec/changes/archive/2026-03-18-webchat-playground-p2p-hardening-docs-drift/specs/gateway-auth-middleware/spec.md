# Gateway Auth Middleware (Delta)

## Changes

### Export

- The auth middleware constructor SHALL be exported as `RequireAuth` (was `requireAuth`)
- **WHEN** `auth` is `nil` (no OIDC configured), the middleware SHALL pass all requests through unchanged
- **WHEN** `auth` is non-nil, the middleware SHALL validate the `lango_session` cookie and return 401 if invalid

### Session Isolation

- **GIVEN** a WebSocket client connects without OIDC authentication
- **WHEN** `SessionFromContext` returns an empty string
- **THEN** the server SHALL assign the client's unique `clientID` as its `SessionKey`
- **SO THAT** `BroadcastToSession` delivers events only to that specific client
