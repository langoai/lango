---
title: WebSocket
---

# WebSocket

Lango supports WebSocket connections for real-time streaming of agent responses. Clients receive incremental updates as the agent processes a request.

## Configuration

> **Settings:** `lango settings` → Server

```json
{
  "server": {
    "wsEnabled": true,
    "allowedOrigins": [
      "http://localhost:3000",
      "https://your-app.example.com"
    ]
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `server.wsEnabled` | `bool` | `true` | Enable WebSocket support |
| `server.allowedOrigins` | `[]string` | `[]` | Allowed CORS origins (empty = allow all) |

!!! warning "CORS in Production"

    Always set `server.allowedOrigins` explicitly in production to restrict which domains can establish WebSocket connections.

## Events

| Event | Payload | Description |
|-------|---------|-------------|
| `agent.thinking` | `{sessionKey}` | Sent before agent execution begins |
| `agent.chunk` | `{sessionKey, chunk}` | Streamed text chunk during LLM generation |
| `agent.done` | `{sessionKey}` | Sent after agent execution completes |

### Event Scoping

Events are scoped to the requesting user's session. A client only receives events for its own session, not events from other users or sessions.

### Backward Compatibility

WebSocket streaming is additive. The full agent response is still returned in the standard RPC result. Clients that do not use WebSocket continue to work without modification.

## Related

- [HTTP API](http-api.md) -- REST endpoints and server configuration
- [Security](../security/index.md) -- Authentication and access control
