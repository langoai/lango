# WebChat Playground

## Overview

The gateway SHALL serve an embedded HTML page at `GET /playground` that provides a browser-based chat interface for interacting with the agent via the existing WebSocket JSON-RPC protocol.

## Purpose

Capability spec for webchat-playground. See requirements below for scope and behavior contracts.

## Requirements

### Serving
The gateway SHALL serve the embedded playground HTML at `GET /playground` when HTTP serving is enabled.

#### Scenario: Playground HTML is served from the binary
- **GIVEN** `server.httpEnabled` is `true`
- **WHEN** a client requests `GET /playground`
- **THEN** the server SHALL return the embedded HTML with `Content-Type: text/html; charset=utf-8`
- **AND** the HTML SHALL be embedded in the binary using Go's `go:embed` directive

### Authentication
The playground SHALL follow the same authentication posture as the existing WebSocket and status surfaces.

#### Scenario: OIDC-protected playground
- **GIVEN** OIDC authentication is configured
- **WHEN** a user requests `/playground`
- **THEN** the route SHALL require authentication using the same middleware as `/ws` and `/status`

#### Scenario: Dev-mode playground remains open
- **GIVEN** OIDC authentication is not configured
- **WHEN** a user requests `/playground`
- **THEN** `/playground` SHALL be accessible without authentication

### WebSocket Integration
The playground SHALL connect to the existing WebSocket JSON-RPC surface rather than inventing a second chat transport.

#### Scenario: Playground uses the existing WebSocket chat protocol
- **WHEN** the playground page loads
- **THEN** it SHALL establish a WebSocket connection to `/ws` using `location`-based URL construction
- **AND** it SHALL use the JSON-RPC 2.0 `chat.message` method for sending messages
- **AND** it SHALL handle `agent.thinking`, `agent.chunk`, `agent.done`, `agent.error`, `agent.progress`, and `agent.warning` events

### UI Capabilities
The playground SHALL provide a lightweight browser chat UI without external CDN dependencies.

#### Scenario: Playground UI provides core chat affordances
- **WHEN** a user interacts with the playground
- **THEN** it SHALL render basic markdown, support dark and light modes, display connection status, auto-reconnect with exponential backoff, and avoid external CDN dependencies
