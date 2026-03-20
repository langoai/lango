# P2P REST API (Delta)

## Changes

### Authentication

- **GIVEN** OIDC authentication is configured
- **WHEN** an unauthenticated client requests any `/api/p2p/*` endpoint
- **THEN** the server SHALL return 401 Unauthorized

- **GIVEN** OIDC is not configured (dev mode)
- **WHEN** a client requests any `/api/p2p/*` endpoint
- **THEN** the request SHALL pass through without authentication (backward compatible)

### Implementation

- `registerP2PRoutes` SHALL accept a `*gateway.AuthManager` parameter
- The P2P route group SHALL apply `gateway.RequireAuth(auth)` middleware
