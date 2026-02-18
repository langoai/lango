## MODIFIED Requirements

### Requirement: server.go (Core Server)
The `server.go` file SHALL contain the Server struct definition, Config struct with `AllowedOrigins`, RPC protocol types (RPCRequest, RPCResponse, RPCError, RPCHandler), the constructor `New()`, route setup with auth middleware, handler registration, server Start/Shutdown lifecycle, and HTTP endpoint handlers (health, status). The `RPCHandler` type SHALL be `func(client *Client, params json.RawMessage) (interface{}, error)` to provide handler access to the calling client's session context.

#### Scenario: Server Constructor
- **WHEN** `gateway.New()` is called with config, agent, provider, store, and auth parameters
- **THEN** it SHALL return a fully initialized Server
- **THEN** it SHALL register all RPC handlers (chat.message, sign.response, encrypt.response, decrypt.response, companion.hello, approval.response)
- **THEN** it SHALL wire up the provider sender if provider is non-nil
- **THEN** it SHALL configure the WebSocket upgrader with `makeOriginChecker(cfg.AllowedOrigins)`

#### Scenario: Route Protection
- **WHEN** routes are configured
- **THEN** `/health` SHALL be public (no auth middleware)
- **THEN** `/auth/*` SHALL be public with rate limiting
- **THEN** `/ws` and `/status` SHALL be in a protected route group with `requireAuth` middleware
- **THEN** `/companion` SHALL be separate (no OIDC auth, origin restriction via upgrader)

#### Scenario: Server Lifecycle
- **WHEN** `Start()` is called
- **THEN** it SHALL listen on the configured host:port
- **WHEN** `Start()` returns after `Shutdown()` has been called
- **THEN** it SHALL return `nil` (not `http.ErrServerClosed`), treating graceful shutdown as a normal exit
- **WHEN** `Start()` returns with any other error
- **THEN** it SHALL return that error to the caller
- **WHEN** `Shutdown()` is called
- **THEN** it SHALL close all WebSocket clients and stop the HTTP server

#### Scenario: Graceful shutdown does not produce error
- **WHEN** `Shutdown()` is called on a running server
- **THEN** `Start()` SHALL return `nil`
- **THEN** the caller SHALL NOT log an error for the normal shutdown path
