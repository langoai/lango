## MODIFIED Requirements

### Requirement: server.go (Core Server)
The `server.go` file SHALL contain the Server struct definition, Config struct with `AllowedOrigins`, RPC protocol types (RPCRequest, RPCResponse, RPCError, RPCHandler), the constructor `New()`, route setup with auth middleware, handler registration, server Start/Shutdown lifecycle, and HTTP endpoint handlers (health, status). The `RPCHandler` type SHALL be `func(client *Client, params json.RawMessage) (interface{}, error)` to provide handler access to the calling client's session context. The Server struct SHALL include `shutdownCtx context.Context` and `shutdownCancel context.CancelFunc` fields. The constructor `New()` SHALL initialize these via `context.WithCancel(context.Background())`. The `handleChatMessage()` method SHALL use `s.shutdownCtx` as the parent context for all per-request contexts (both `deadline.New()` and `context.WithTimeout()` paths). The `Shutdown()` method SHALL call `s.shutdownCancel()` before closing WebSocket connections and stopping the HTTP server, so that all in-flight request contexts are immediately cancelled. `Shutdown()` SHALL be safe to call before `Start()`, after failed startup, and more than once.

#### Scenario: Shutdown before start is safe
- **WHEN** `Shutdown()` is called on a newly constructed gateway server before `Start()`
- **THEN** it SHALL return `nil`
- **AND** it SHALL NOT panic

#### Scenario: Shutdown after failed start is safe
- **GIVEN** the configured gateway listen address is already occupied
- **WHEN** `Start()` fails
- **AND** `Shutdown()` is called for cleanup
- **THEN** shutdown SHALL NOT panic

#### Scenario: Repeated shutdown is safe
- **WHEN** `Shutdown()` is called more than once on a gateway server
- **THEN** subsequent calls SHALL NOT panic
