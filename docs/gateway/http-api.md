---
title: HTTP API
---

# HTTP API

Lango exposes an HTTP API for health monitoring, agent discovery, authentication, and chat interaction.

## Server Configuration

```yaml
server:
  host: localhost
  port: 18789
  httpEnabled: true
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `server.host` | `string` | `localhost` | Bind address for the HTTP server |
| `server.port` | `int` | `18789` | Port number |
| `server.httpEnabled` | `bool` | `true` | Enable the HTTP server |

## Endpoints

### Health Check

```
GET /health
```

Returns the server health status. Use this for monitoring, load balancer probes, and Docker health checks.

```bash
curl http://localhost:18789/health
```

### Agent Card (A2A)

```
GET /.well-known/agent.json
```

Returns the agent's A2A agent card when the A2A protocol is enabled (`a2a.enabled: true`). This endpoint follows the [Agent-to-Agent protocol](../features/a2a-protocol.md) specification for remote agent discovery.

### Authentication

Authentication endpoints are available when OIDC is configured. See [Authentication](../security/authentication.md) for details on the OAuth login flow and token management.

### Chat

The main chat endpoint accepts user messages and returns agent responses. When WebSocket is enabled, responses are streamed in real time via WebSocket events alongside the standard HTTP response.

## Related

- [WebSocket](websocket.md) -- Real-time streaming events
- [Authentication](../security/authentication.md) -- OIDC and OAuth configuration
- [A2A Protocol](../features/a2a-protocol.md) -- Agent-to-Agent discovery
