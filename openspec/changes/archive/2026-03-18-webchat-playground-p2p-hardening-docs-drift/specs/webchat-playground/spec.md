# WebChat Playground

## Overview

The gateway SHALL serve an embedded HTML page at `GET /playground` that provides a browser-based chat interface for interacting with the agent via the existing WebSocket JSON-RPC protocol.

## Requirements

### Serving

- **GIVEN** `server.httpEnabled` is `true`
- **WHEN** a client requests `GET /playground`
- **THEN** the server SHALL return the embedded HTML with `Content-Type: text/html; charset=utf-8`
- **AND** the HTML SHALL be embedded in the binary using Go's `go:embed` directive

### Authentication

- **GIVEN** OIDC authentication is configured
- **THEN** `/playground` SHALL require authentication (same middleware as `/ws` and `/status`)
- **GIVEN** OIDC is not configured (dev mode)
- **THEN** `/playground` SHALL be accessible without authentication

### WebSocket Integration

- **WHEN** the playground page loads
- **THEN** it SHALL establish a WebSocket connection to `/ws` using `location`-based URL construction
- **AND** it SHALL use the JSON-RPC 2.0 `chat.message` method for sending messages
- **AND** it SHALL handle the following events: `agent.thinking`, `agent.chunk`, `agent.done`, `agent.error`, `agent.progress`, `agent.warning`

### UI Capabilities

- The playground SHALL render basic markdown: code blocks, inline code, bold, italic
- The playground SHALL support dark and light modes via `prefers-color-scheme`
- The playground SHALL display connection status (connected, disconnected, reconnecting)
- The playground SHALL auto-reconnect with exponential backoff on disconnect
- The playground SHALL have zero external CDN dependencies
