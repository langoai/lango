## Purpose

Capability spec for companion-protocol. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Companion handshake
The gateway SHALL accept WebSocket connections on `/companion` endpoint and perform mutual identification.

#### Scenario: Successful handshake
- **WHEN** companion connects and sends `{"method": "companion.hello", "params": {"deviceId": "...", "capabilities": ["sign", "encrypt", "decrypt"]}}`
- **THEN** server responds with `{"result": {"serverId": "...", "version": "1.0"}}`

### Requirement: Crypto request routing
The server SHALL forward crypto requests to connected companions via WebSocket.

#### Scenario: Sign request
- **WHEN** server needs signature and companion is connected
- **THEN** server sends `{"method": "sign.request", "params": {"requestId": "...", "keyId": "...", "payload": "base64..."}}`

#### Scenario: Sign response
- **WHEN** companion returns `{"method": "sign.response", "params": {"requestId": "...", "signature": "base64..."}}`
- **THEN** server forwards signature to waiting goroutine

### Requirement: Approval request
The server SHALL send approval requests to companions for sensitive operations.

#### Scenario: Secrets get approval
- **WHEN** AI calls secrets.get and approval is required
- **THEN** server sends `{"method": "approval.request", "params": {"requestId": "...", "tool": "secrets.get", "args": {...}}}`

#### Scenario: Approval granted
- **WHEN** companion sends `{"method": "approval.response", "params": {"requestId": "...", "approved": true}}`
- **THEN** server allows the operation to proceed
