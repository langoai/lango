## MODIFIED Requirements

### Requirement: server.go (Core Server)
The `server.go` file SHALL contain the Server struct definition, Config struct with `AllowedOrigins`, RPC protocol types (RPCRequest, RPCResponse, RPCError, RPCHandler), the constructor `New()`, route setup with auth middleware, handler registration, server Start/Shutdown lifecycle, and HTTP endpoint handlers (health, status). The `RPCHandler` type SHALL be `func(client *Client, params json.RawMessage) (interface{}, error)` to provide handler access to the calling client's session context. The Server struct SHALL include `shutdownCtx context.Context` and `shutdownCancel context.CancelFunc` fields. The constructor `New()` SHALL initialize these via `context.WithCancel(context.Background())`. The `handleChatMessage()` method SHALL use `s.shutdownCtx` as the parent context for all per-request contexts (both `deadline.New()` and `context.WithTimeout()` paths). The `Shutdown()` method SHALL call `s.shutdownCancel()` before closing WebSocket connections and stopping the HTTP server, so that all in-flight request contexts are immediately cancelled.

#### Scenario: Server Constructor
- **WHEN** `gateway.New()` is called with config, agent, provider, store, and auth parameters
- **THEN** it SHALL return a fully initialized Server
- **THEN** it SHALL register all RPC handlers (chat.message, sign.response, encrypt.response, decrypt.response, companion.hello, approval.response)
- **THEN** it SHALL wire up the provider sender if provider is non-nil
- **THEN** it SHALL configure the WebSocket upgrader with `makeOriginChecker(cfg.AllowedOrigins)`
- **THEN** it SHALL initialize `shutdownCtx` and `shutdownCancel` via `context.WithCancel(context.Background())`

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
- **THEN** it SHALL call `shutdownCancel()` first to cancel all in-flight request contexts
- **THEN** it SHALL close all WebSocket clients and stop the HTTP server

#### Scenario: Graceful shutdown does not produce error
- **WHEN** `Shutdown()` is called on a running server
- **THEN** `Start()` SHALL return `nil`
- **THEN** the caller SHALL NOT log an error for the normal shutdown path

#### Scenario: Shutdown cancels in-flight request contexts
- **WHEN** `handleChatMessage()` is processing a request with a context derived from `shutdownCtx`
- **AND** `Shutdown()` is called
- **THEN** the request context SHALL be cancelled immediately
- **THEN** `RunStreaming` SHALL observe `ctx.Err() != nil` and return

#### Scenario: Shutdown cancels pending approval waits
- **WHEN** `RequestApproval()` is waiting for a companion response with a context derived from `shutdownCtx`
- **AND** `Shutdown()` is called
- **THEN** `RequestApproval()` SHALL return `context.Canceled` immediately
- **THEN** it SHALL NOT wait for the 30-second approval timeout

#### Scenario: Request contexts use shutdownCtx as parent
- **WHEN** `handleChatMessage()` creates a per-request context
- **THEN** it SHALL use `s.shutdownCtx` as the parent (not `context.Background()`)
- **THEN** the idle-timeout path (`deadline.New()`) SHALL use `s.shutdownCtx` as parent
- **THEN** the fixed-timeout path (`context.WithTimeout()`) SHALL use `s.shutdownCtx` as parent
