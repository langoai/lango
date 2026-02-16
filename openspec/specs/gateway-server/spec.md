## Requirements

### Requirement: Gateway Initialization
The gateway server SHALL be initialized without requiring an `AuthManager` or `RPCProvider`. The `gateway.New()` function SHALL accept `nil` for optional parameters (rpcProvider, authManager). The gateway SHALL serve HTTP and WebSocket endpoints for direct chat without OIDC authentication. The Config struct SHALL include `AllowedOrigins []string` for WebSocket CORS control. The WebSocket upgrader SHALL use `makeOriginChecker(cfg.AllowedOrigins)` instead of allowing all origins.

#### Scenario: Gateway startup without auth
- **WHEN** the gateway is created with `nil` authManager and `nil` rpcProvider
- **THEN** it SHALL start successfully
- **THEN** it SHALL serve the `chat.message` RPC endpoint
- **THEN** it SHALL NOT register `/companion` WebSocket endpoint
- **THEN** it SHALL NOT require OIDC configuration
- **THEN** all endpoints SHALL be accessible without authentication

#### Scenario: Gateway startup with auth
- **WHEN** the gateway is created with a configured authManager
- **THEN** `/ws` and `/status` SHALL require valid `lango_session` cookie
- **THEN** `/health` and `/auth/*` SHALL remain publicly accessible
- **THEN** `/companion` SHALL be accessible without OIDC auth (origin restriction only)

### Requirement: File Structure
The gateway package SHALL be organized into focused files where no single file exceeds 200 lines of non-test code. The decomposition SHALL be: `server.go` (server lifecycle, types, routes, health), `websocket.go` (WebSocket connection management, Client struct, broadcast), `handlers.go` (RPC handler implementations), `auth.go` (OIDC authentication, already separated).

#### Scenario: File Size Compliance
- **GIVEN** the gateway package after decomposition
- **WHEN** line counts are measured for non-test `.go` files
- **THEN** no single file SHALL exceed 200 lines

#### Scenario: Build and Test Integrity
- **GIVEN** the decomposed gateway package
- **WHEN** `go build ./internal/gateway/...` is run
- **THEN** it SHALL succeed with no errors
- **WHEN** `go test ./internal/gateway/...` is run
- **THEN** all existing tests SHALL pass without modification

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
- **WHEN** `Shutdown()` is called
- **THEN** it SHALL close all WebSocket clients and stop the HTTP server

### Requirement: websocket.go (Connection Management)
The `websocket.go` file SHALL contain the Client struct, WebSocket upgrade handlers, read/write pump goroutines, send helpers, client close logic, broadcast methods, and client removal. The `handleWebSocketConnection` function SHALL extract the authenticated session key from the request context via `SessionFromContext` and bind it to `Client.SessionKey`.

#### Scenario: Client Connection with Auth
- **WHEN** an authenticated client connects to `/ws`
- **THEN** a Client SHALL be created with `SessionKey` set from `SessionFromContext(r.Context())`

#### Scenario: Client Connection without Auth
- **WHEN** a client connects to `/ws` with no auth configured
- **THEN** a Client SHALL be created with empty `SessionKey`

#### Scenario: RPC Dispatch
- **WHEN** a Client receives a JSON message in readPump
- **THEN** it SHALL parse it as RPCRequest and dispatch to the matching registered handler
- **THEN** it SHALL pass the client reference as the first argument to the handler
- **THEN** it SHALL send back RPCResponse with the handler result or error

#### Scenario: Broadcast
- **WHEN** `Broadcast()` is called
- **THEN** the message SHALL be sent to all clients with type "ui"
- **WHEN** `BroadcastToCompanions()` is called
- **THEN** the message SHALL be sent to all clients with type "companion"

### Requirement: handlers.go (RPC Handlers)
The `handlers.go` file SHALL contain all RPC handler method implementations. All handlers SHALL accept `client *Client` as the first parameter. The `handleChatMessage` handler SHALL use the client's authenticated session key when available, preventing session fixation.

#### Scenario: Chat Message Handler
- **WHEN** `handleChatMessage` receives a valid message from an authenticated client
- **THEN** it SHALL use `client.SessionKey` as the session key, ignoring any client-provided `sessionKey`
- **WHEN** `handleChatMessage` receives a message from an unauthenticated client (auth disabled)
- **THEN** it SHALL use the client-provided `sessionKey` or "default" if not provided
- **WHEN** agent is nil
- **THEN** it SHALL return "agent not configured" error

#### Scenario: Session key available in agent context
- **WHEN** a `chat.message` RPC request is processed with session key "default"
- **THEN** `session.WithSessionKey(ctx, "default")` SHALL be called before `agent.RunAndCollect`
- **AND** `session.SessionKeyFromContext` within the agent pipeline SHALL return "default"

#### Scenario: Authenticated session key propagated
- **WHEN** an authenticated client sends a `chat.message` RPC request
- **THEN** the client's authenticated session key SHALL be injected into the context
- **AND** downstream approval routing SHALL use that session key

#### Scenario: Security Proxy Handlers
- **WHEN** `handleSignResponse`, `handleEncryptResponse`, or `handleDecryptResponse` is called
- **THEN** it SHALL delegate to the corresponding RPCProvider method
- **WHEN** provider is nil
- **THEN** it SHALL return "provider not configured" error

#### Scenario: Approval Workflow
- **WHEN** `RequestApproval` is called
- **THEN** it SHALL broadcast an approval request to companions
- **THEN** it SHALL wait for a response or timeout after 30 seconds
- **WHEN** no companion is connected
- **THEN** it SHALL return an error immediately
