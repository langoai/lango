## ADDED Requirements

### Requirement: A2A-over-P2P Message Protocol

The system SHALL implement A2A message exchange over libp2p streams using protocol ID `/lango/a2a/1.0.0`. All messages SHALL be JSON-encoded. Each `Request` SHALL carry a `type` field (`tool_invoke`, `capability_query`, or `agent_card`), a `sessionToken`, a UUID `requestId`, and an optional `payload` map. Each `Response` SHALL carry the matching `requestId`, a `status` field (`"ok"`, `"error"`, or `"denied"`), an optional `result` map, an optional `error` string, an optional `attestationProof` byte slice, and a `timestamp`.

#### Scenario: Tool invoke request routed to executor
- **WHEN** an incoming stream delivers a `Request` with `type="tool_invoke"` and `payload.toolName="search"`
- **THEN** the `Handler` SHALL call the registered `ToolExecutor` with the tool name and params map

#### Scenario: Agent card request served
- **WHEN** an incoming stream delivers a `Request` with `type="agent_card"`
- **THEN** the `Handler` SHALL call the `CardProvider` function and return its result with `status="ok"`

#### Scenario: Capability query returns agent card
- **WHEN** an incoming stream delivers a `Request` with `type="capability_query"`
- **THEN** the `Handler` SHALL return the agent card contents with `status="ok"` as a capability listing

#### Scenario: Unknown request type returns error
- **WHEN** an incoming stream delivers a `Request` with an unrecognized `type` value
- **THEN** the `Handler` SHALL return a `Response` with `status="error"` and an error describing the unknown type

---

### Requirement: Session Token Validation on Every Request

The `Handler` SHALL validate the session token on every incoming request before dispatching to the type-specific handler. Token validation SHALL iterate over all active sessions in the `SessionStore` and check for a matching token using `SessionStore.Validate`. If no session matches, the handler MUST return a `Response` with `status="denied"` and `error="invalid or expired session token"`.

#### Scenario: Valid session token grants access
- **WHEN** a `Request` arrives with a `sessionToken` that matches an active non-expired session
- **THEN** the handler SHALL resolve the peer DID and proceed with the request

#### Scenario: Invalid session token denied
- **WHEN** a `Request` arrives with a `sessionToken` that does not match any active session
- **THEN** the handler SHALL return `{"status": "denied", "error": "invalid or expired session token"}`

#### Scenario: Expired session token denied
- **WHEN** a `Request` arrives with a token from a session whose `ExpiresAt` is in the past
- **THEN** the handler SHALL return `{"status": "denied"}` and the expired session SHALL be removed from the store

---

### Requirement: Firewall Enforcement on Tool Invocations

The `Handler.handleToolInvoke` method MUST call `Firewall.FilterQuery(peerDID, toolName)` before executing any tool. A non-nil error from the firewall SHALL cause the handler to return a `Response` with `status="denied"`. The tool executor SHALL NOT be called if the firewall rejects the query.

#### Scenario: Firewall blocks unauthorized tool
- **WHEN** a peer requests a tool that is not in its allow list
- **THEN** `handleToolInvoke` SHALL return `{"status": "denied"}` without calling the `ToolExecutor`

#### Scenario: Missing toolName in payload
- **WHEN** a `tool_invoke` request arrives with no `toolName` field in the payload
- **THEN** the handler SHALL return `{"status": "error", "error": "missing toolName in payload"}`

---

### Requirement: Response Sanitization and ZK Attestation on Tool Results

After successful tool execution, the `Handler` SHALL pass the result through `Firewall.SanitizeResponse` to remove sensitive fields. If a `ZKAttestFunc` is configured on the firewall, the handler SHALL compute a SHA-256 hash of the sanitized result and the local agent DID and include the resulting attestation proof in `Response.AttestationProof`.

#### Scenario: Tool result sanitized before returning
- **WHEN** a tool returns a result containing a sensitive field (e.g., `"token": "secret"`)
- **THEN** the `Response.Result` SHALL have the sensitive field removed

#### Scenario: ZK attestation included when available
- **WHEN** the firewall has a `ZKAttestFunc` configured and a tool invocation succeeds
- **THEN** `Response.AttestationProof` SHALL contain a non-empty byte slice

---

### Requirement: P2PRemoteAgent Adapter

The `P2PRemoteAgent` SHALL implement a remote agent adapter that wraps a peer ID and session token to send requests over P2P streams. `InvokeTool` SHALL open a new libp2p stream to the peer's ID using protocol `/lango/a2a/1.0.0`, encode the tool invoke request, and decode the response. Non-"ok" responses MUST return an error using the `Response.Error` field. `QueryCapabilities` and `FetchAgentCard` SHALL use the same stream-open-encode-decode pattern.

#### Scenario: InvokeTool sends request and returns result
- **WHEN** `P2PRemoteAgent.InvokeTool(ctx, "search", params)` is called
- **THEN** a new stream to the target peer SHALL be opened, a `tool_invoke` request encoded, and the `Response.Result` returned on `status="ok"`

#### Scenario: Remote error response propagated
- **WHEN** the remote `Handler` returns `{"status": "error", "error": "tool not found"}`
- **THEN** `InvokeTool` SHALL return an error containing "tool not found"

#### Scenario: Stream open failure returns error
- **WHEN** `host.NewStream` fails (e.g., peer unreachable)
- **THEN** `InvokeTool` SHALL return a wrapped error containing "open stream to"

#### Scenario: ZK attestation proof logged on receipt
- **WHEN** `InvokeTool` receives a `Response` with a non-empty `AttestationProof`
- **THEN** the adapter SHALL log "response has ZK attestation" at debug level
