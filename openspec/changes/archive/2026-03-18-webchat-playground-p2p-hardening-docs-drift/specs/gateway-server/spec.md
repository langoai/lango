# Gateway Server (Delta)

## Changes

### Protected Route Group

- The protected route group SHALL include `GET /playground` alongside `GET /status`
- **WHEN** `server.httpEnabled` is `true`
- **THEN** `/playground` SHALL be registered with `RequireAuth` middleware
