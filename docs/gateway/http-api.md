---
title: HTTP API
---

# HTTP API

Lango exposes an HTTP API for health monitoring, agent discovery, authentication, and chat interaction.

## Server Configuration

> **Settings:** `lango settings` → Server

```json
{
  "server": {
    "host": "localhost",
    "port": 18789,
    "httpEnabled": true
  }
}
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

### P2P Network

When P2P networking is enabled (`p2p.enabled: true`), the gateway exposes read-only endpoints for querying the running node's state. These endpoints are public (no authentication required) and return only node metadata.

#### `GET /api/p2p/status`

Returns the local node's peer ID, listen addresses, and connected peer count.

```bash
curl http://localhost:18789/api/p2p/status
```

```json
{
  "peerId": "12D3KooW...",
  "listenAddrs": ["/ip4/0.0.0.0/tcp/9000"],
  "connectedPeers": 2,
  "mdnsEnabled": true
}
```

#### `GET /api/p2p/peers`

Returns the list of currently connected peers with their IDs and multiaddresses.

```bash
curl http://localhost:18789/api/p2p/peers
```

```json
{
  "peers": [
    {
      "peerId": "12D3KooW...",
      "addrs": ["/ip4/172.20.0.3/tcp/9002"]
    }
  ],
  "count": 1
}
```

#### `GET /api/p2p/identity`

Returns the local DID derived from the wallet and the libp2p peer ID.

```bash
curl http://localhost:18789/api/p2p/identity
```

```json
{
  "did": "did:lango:02abc...",
  "peerId": "12D3KooW..."
}
```

If no identity provider is configured, `did` is `null`.

#### `GET /api/p2p/reputation`

Returns the trust score and exchange history for a peer. The `peer_did` query parameter is required.

```bash
curl "http://localhost:18789/api/p2p/reputation?peer_did=did:lango:02abc..."
```

```json
{
  "peerDid": "did:lango:02abc...",
  "trustScore": 0.85,
  "successfulExchanges": 42,
  "failedExchanges": 3,
  "timeoutCount": 1,
  "firstSeen": "2026-02-20T10:00:00Z",
  "lastInteraction": "2026-02-24T14:30:00Z"
}
```

If the reputation system is not enabled or the peer has no history, the response indicates the default state (new peers are trusted by default).

#### `GET /api/p2p/pricing`

Returns tool pricing configuration. Use the optional `tool` query parameter to query a specific tool's price.

```bash
# Get all tool pricing
curl http://localhost:18789/api/p2p/pricing
```

```json
{
  "enabled": true,
  "currency": "USDC",
  "perQuery": "0.10",
  "toolPrices": {
    "knowledge_search": "0.25",
    "browser_navigate": "0.50"
  }
}
```

```bash
# Get pricing for a specific tool
curl "http://localhost:18789/api/p2p/pricing?tool=knowledge_search"
```

```json
{
  "tool": "knowledge_search",
  "price": "0.25",
  "currency": "USDC"
}
```

!!! note
    These REST endpoints query the **running server's persistent P2P node**. The CLI commands (`lango p2p status`, etc.) create ephemeral nodes for one-off operations. For monitoring and automation, prefer the REST API.

## Related

- [WebSocket](websocket.md) -- Real-time streaming events
- [Authentication](../security/authentication.md) -- OIDC and OAuth configuration
- [A2A Protocol](../features/a2a-protocol.md) -- Agent-to-Agent discovery
