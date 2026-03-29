# Configuration Reference

Complete reference of all configuration keys available in Lango. Configuration is stored in encrypted profiles managed by [`lango config`](cli/config.md) commands. Use [`lango onboard`](cli/core.md#lango-onboard) for guided setup or [`lango settings`](cli/core.md#lango-settings) for the full interactive editor.

All configuration is managed through the **`lango settings`** TUI (interactive terminal editor) or by importing a JSON file with **`lango config import`**. Lango does not use YAML configuration files. The JSON examples below show the structure expected by `lango config import` and reflect what `lango settings` edits behind the scenes.

See [Configuration Basics](getting-started/configuration.md) for an introduction to the configuration system.

---

## Server

Gateway server settings for HTTP API and WebSocket connections.

> **Settings:** `lango settings` → Server

```json
{
  "server": {
    "host": "localhost",
    "port": 18789,
    "httpEnabled": true,
    "wsEnabled": true,
    "allowedOrigins": []
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `server.host` | `string` | `localhost` | Host address to bind to |
| `server.port` | `int` | `18789` | Port to listen on |
| `server.httpEnabled` | `bool` | `true` | Enable HTTP API endpoints |
| `server.wsEnabled` | `bool` | `true` | Enable WebSocket server |
| `server.allowedOrigins` | `[]string` | `[]` | Allowed origins for CORS. Empty = same-origin only |

---

## Agent

LLM agent settings including model selection, prompt configuration, and timeouts.

> **Settings:** `lango settings` → Agent

```json
{
  "agent": {
    "provider": "anthropic",
    "model": "claude-sonnet-4-20250514",
    "fallbackProvider": "",
    "fallbackModel": "",
    "maxTokens": 4096,
    "temperature": 0.7,
    "systemPromptPath": "",
    "promptsDir": "",
    "requestTimeout": "5m",
    "toolTimeout": "2m",
    "multiAgent": false,
    "agentsDir": "",
    "autoExtendTimeout": false,
    "maxRequestTimeout": ""
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `agent.provider` | `string` | `anthropic` | Primary AI provider ID (references `providers.<id>`) |
| `agent.model` | `string` | | Model ID to use (e.g., `claude-sonnet-4-20250514`) |
| `agent.fallbackProvider` | `string` | | Fallback provider ID when primary fails |
| `agent.fallbackModel` | `string` | | Fallback model ID |
| `agent.maxTokens` | `int` | `4096` | Maximum tokens per response |
| `agent.temperature` | `float64` | `0.7` | Sampling temperature (0.0 - 1.0) |
| `agent.systemPromptPath` | `string` | | Path to a custom system prompt file |
| `agent.promptsDir` | `string` | | Directory containing `.md` files for [system prompts](features/system-prompts.md) |
| `agent.requestTimeout` | `duration` | `5m` | Maximum duration for a single AI provider request |
| `agent.toolTimeout` | `duration` | `2m` | Maximum duration for a single tool call |
| `agent.multiAgent` | `bool` | `false` | Enable [multi-agent orchestration](features/multi-agent.md) |
| `agent.agentsDir` | `string` | `""` | Directory containing user-defined [AGENT.md](features/multi-agent.md#custom-agent-definitions) agent definitions |
| `agent.autoExtendTimeout` | `bool` | `false` | Auto-extend deadline when agent activity is detected |
| `agent.maxRequestTimeout` | `duration` | | Absolute max when auto-extend enabled (default: 3x requestTimeout) |

---

## Agent Memory

Per-agent persistent memory for cross-session context retention.

> **Settings:** `lango settings` → Agent Memory

```json
{
  "agentMemory": {
    "enabled": false
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `agentMemory.enabled` | `bool` | `false` | Enable per-agent persistent memory for sub-agents |

---

## Providers

Named AI provider configurations. Referenced by other sections via provider ID.

> **Settings:** `lango settings` → Providers

```json
{
  "providers": {
    "my-anthropic": {
      "type": "anthropic",
      "apiKey": "${ANTHROPIC_API_KEY}"
    },
    "my-openai": {
      "type": "openai",
      "apiKey": "${OPENAI_API_KEY}",
      "baseUrl": "https://api.openai.com/v1"
    },
    "local-ollama": {
      "type": "ollama",
      "baseUrl": "http://localhost:11434/v1"
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `providers.<id>.type` | `string` | | Provider type: `anthropic`, `openai`, `google`, `gemini`, `ollama` |
| `providers.<id>.apiKey` | `string` | | API key (supports `${ENV_VAR}` substitution) |
| `providers.<id>.baseUrl` | `string` | | Base URL for OpenAI-compatible or self-hosted providers |

---

## Logging

> **Settings:** `lango settings` → Logging

```json
{
  "logging": {
    "level": "info",
    "format": "console"
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `logging.level` | `string` | `info` | Log level: `debug`, `info`, `warn`, `error` |
| `logging.format` | `string` | `console` | Output format: `console`, `json` |

---

## Session

Session storage and lifecycle settings.

> **Settings:** `lango settings` → Session

```json
{
  "session": {
    "databasePath": "~/.lango/data.db",
    "ttl": "24h",
    "maxHistoryTurns": 100
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `session.databasePath` | `string` | `~/.lango/data.db` | Path to the SQLite session database |
| `session.ttl` | `duration` | | Session time-to-live before expiration (empty = no expiration) |
| `session.maxHistoryTurns` | `int` | `50` | Maximum conversation turns to retain per session |

---

## Security

### Signer

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `security.signer.provider` | `string` | `local` | Signer provider (`local`) |

### Interceptor

The security interceptor controls tool execution approval and PII protection. See [Tool Approval](security/tool-approval.md) and [PII Redaction](security/pii-redaction.md).

> **Settings:** `lango settings` → Security

```json
{
  "security": {
    "interceptor": {
      "enabled": true,
      "redactPii": false,
      "approvalPolicy": "dangerous",
      "approvalTimeoutSec": 30,
      "notifyChannel": "",
      "sensitiveTools": [],
      "exemptTools": []
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `security.interceptor.enabled` | `bool` | `true` | Enable the security interceptor |
| `security.interceptor.redactPii` | `bool` | `false` | Enable PII redaction in messages |
| `security.interceptor.approvalPolicy` | `string` | `dangerous` | Tool approval policy: `all`, `dangerous`, `configured`, `none` |
| `security.interceptor.approvalTimeoutSec` | `int` | `30` | Timeout for approval requests (seconds) |
| `security.interceptor.notifyChannel` | `string` | | Channel to send approval notifications |
| `security.interceptor.sensitiveTools` | `[]string` | | Tools that always require approval |
| `security.interceptor.exemptTools` | `[]string` | | Tools exempt from approval regardless of policy |

### PII Detection

> **Settings:** `lango settings` → Security

```json
{
  "security": {
    "interceptor": {
      "piiRegexPatterns": [],
      "piiDisabledPatterns": [],
      "piiCustomPatterns": []
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `security.interceptor.piiRegexPatterns` | `[]string` | | Built-in PII regex pattern names to enable |
| `security.interceptor.piiDisabledPatterns` | `[]string` | | Built-in PII patterns to disable |
| `security.interceptor.piiCustomPatterns` | `[]object` | | Custom PII regex patterns (name + regex pairs) |

### Presidio Integration

> **Settings:** `lango settings` → Security

```json
{
  "security": {
    "interceptor": {
      "presidio": {
        "enabled": false,
        "url": "http://localhost:5002",
        "scoreThreshold": 0.7,
        "language": "en"
      }
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `security.interceptor.presidio.enabled` | `bool` | `false` | Enable Microsoft Presidio for advanced PII detection |
| `security.interceptor.presidio.url` | `string` | `http://localhost:5002` | Presidio analyzer service URL |
| `security.interceptor.presidio.scoreThreshold` | `float64` | `0.7` | Minimum confidence score (0.0 - 1.0) |
| `security.interceptor.presidio.language` | `string` | `en` | Language for PII analysis |

---

## Auth

Configure OAuth2/OIDC authentication providers for the gateway API.

> **Settings:** `lango settings` → Auth

```json
{
  "auth": {
    "providers": {
      "google": {
        "issuerUrl": "https://accounts.google.com",
        "clientId": "${GOOGLE_CLIENT_ID}",
        "clientSecret": "${GOOGLE_CLIENT_SECRET}",
        "redirectUrl": "http://localhost:18789/auth/callback",
        "scopes": ["openid", "email", "profile"]
      }
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `auth.providers.<id>.issuerUrl` | `string` | | OIDC issuer URL |
| `auth.providers.<id>.clientId` | `string` | | OAuth2 client ID |
| `auth.providers.<id>.clientSecret` | `string` | | OAuth2 client secret |
| `auth.providers.<id>.redirectUrl` | `string` | | OAuth2 redirect URL |
| `auth.providers.<id>.scopes` | `[]string` | | OAuth2 scopes to request |

---

## Channels

Communication channel configurations.

### Telegram

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `channels.telegram.enabled` | `bool` | `false` | Enable Telegram channel |
| `channels.telegram.botToken` | `string` | | Bot token from BotFather |
| `channels.telegram.allowlist` | `[]int64` | `[]` | Allowed user/group IDs (empty = allow all) |

### Discord

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `channels.discord.enabled` | `bool` | `false` | Enable Discord channel |
| `channels.discord.botToken` | `string` | | Bot token from Discord Developer Portal |
| `channels.discord.applicationId` | `string` | | Application ID for slash commands |
| `channels.discord.allowedGuilds` | `[]string` | `[]` | Allowed guild IDs (empty = allow all) |

### Slack

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `channels.slack.enabled` | `bool` | `false` | Enable Slack channel |
| `channels.slack.botToken` | `string` | | Bot OAuth token |
| `channels.slack.appToken` | `string` | | App-level token for Socket Mode |
| `channels.slack.signingSecret` | `string` | | Signing secret for request verification |

---

## Tools

### Exec Tool

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `tools.exec.defaultTimeout` | `duration` | `30s` | Default timeout for shell command execution |
| `tools.exec.allowBackground` | `bool` | `true` | Allow background command execution |
| `tools.exec.workDir` | `string` | | Working directory for command execution |

### Filesystem Tool

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `tools.filesystem.maxReadSize` | `int` | `10485760` | Maximum file read size in bytes (10 MB) |
| `tools.filesystem.allowedPaths` | `[]string` | | Allowed filesystem paths (empty = all) |

### Browser Tool

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `tools.browser.enabled` | `bool` | `false` | Enable browser automation tool |
| `tools.browser.headless` | `bool` | `true` | Run browser in headless mode |
| `tools.browser.sessionTimeout` | `duration` | `5m` | Browser session timeout |

### Output Manager

Token-based tiered compression for large tool outputs. Applied as middleware to all tools.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `tools.outputManager.tokenBudget` | `int` | `2000` | Maximum tokens before output is compressed |
| `tools.outputManager.headRatio` | `float` | `0.7` | Fraction of budget allocated to output head |
| `tools.outputManager.tailRatio` | `float` | `0.3` | Fraction of budget allocated to output tail |

---

## Hooks

Tool execution hooks for security filtering, access control, and event publishing.

> **Settings:** `lango settings` → Hooks

```json
{
  "hooks": {
    "enabled": false,
    "securityFilter": false,
    "accessControl": false,
    "eventPublishing": false,
    "knowledgeSave": false,
    "blockedCommands": []
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `hooks.enabled` | `bool` | `false` | Enable the tool execution hook system |
| `hooks.securityFilter` | `bool` | `false` | Block dangerous commands via security filter hook |
| `hooks.accessControl` | `bool` | `false` | Enable per-agent tool access control |
| `hooks.eventPublishing` | `bool` | `false` | Publish tool execution events to the [event bus](features/multi-agent.md) |
| `hooks.knowledgeSave` | `bool` | `false` | Auto-save knowledge extracted from tool results |
| `hooks.blockedCommands` | `[]string` | `[]` | Command patterns to block when security filter is active |

---

## Context Profile

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `contextProfile` | `string` | - | Preset that auto-configures context subsystems: `off`, `lite`, `balanced`, `full` |

Profiles control which subsystems are enabled:

| Profile | Knowledge | Memory | Librarian | Graph |
|---------|-----------|--------|-----------|-------|
| `off` | - | - | - | - |
| `lite` | ✓ | ✓ | - | - |
| `balanced` | ✓ | ✓ | ✓ | - |
| `full` | ✓ | ✓ | ✓ | ✓ |

User-explicit overrides take precedence over profile defaults.

---

## Knowledge

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `knowledge.enabled` | `bool` | `false` | Enable the [knowledge system](features/knowledge.md) |
| `knowledge.maxContextPerLayer` | `int` | `5` | Maximum context items per knowledge layer |

---

## Skill

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `skill.enabled` | `bool` | `false` | Enable the [skill system](features/skills.md) |
| `skill.skillsDir` | `string` | `~/.lango/skills` | Directory for skill files |
| `skill.allowImport` | `bool` | `false` | Allow importing skills from external sources |
| `skill.maxBulkImport` | `int` | `50` | Maximum skills per bulk import |
| `skill.importConcurrency` | `int` | `5` | Concurrent import workers |
| `skill.importTimeout` | `duration` | `2m` | Timeout per skill import |

---

## Observational Memory

> **Settings:** `lango settings` → Observational Memory

```json
{
  "observationalMemory": {
    "enabled": false,
    "provider": "",
    "model": "",
    "messageTokenThreshold": 1000,
    "observationTokenThreshold": 2000,
    "maxMessageTokenBudget": 8000,
    "maxReflectionsInContext": 5,
    "maxObservationsInContext": 20
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `observationalMemory.enabled` | `bool` | `false` | Enable [observational memory](features/observational-memory.md) |
| `observationalMemory.provider` | `string` | | AI provider for memory extraction (empty = agent default) |
| `observationalMemory.model` | `string` | | Model for memory extraction (empty = agent default) |
| `observationalMemory.messageTokenThreshold` | `int` | `1000` | Minimum tokens in recent messages before triggering observation |
| `observationalMemory.observationTokenThreshold` | `int` | `2000` | Token threshold to trigger reflection |
| `observationalMemory.maxMessageTokenBudget` | `int` | `8000` | Max tokens to include from message history |
| `observationalMemory.maxReflectionsInContext` | `int` | `5` | Max reflections injected into LLM context |
| `observationalMemory.maxObservationsInContext` | `int` | `20` | Max observations injected into LLM context |

---

## Embedding & RAG

> **Settings:** `lango settings` → Embedding & RAG

```json
{
  "embedding": {
    "providerID": "my-openai",
    "provider": "",
    "model": "text-embedding-3-small",
    "dimensions": 1536,
    "local": {
      "baseUrl": "http://localhost:11434/v1",
      "model": ""
    },
    "rag": {
      "enabled": false,
      "maxResults": 5,
      "collections": []
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `embedding.providerID` | `string` | | References a key in the `providers` map |
| `embedding.provider` | `string` | | Embedding provider type (set to `local` for Ollama) |
| `embedding.model` | `string` | | Embedding model identifier |
| `embedding.dimensions` | `int` | | Embedding vector dimensionality |
| `embedding.local.baseUrl` | `string` | | Local embedding service URL (e.g., Ollama) |
| `embedding.local.model` | `string` | | Model override for local provider |
| `embedding.rag.enabled` | `bool` | `false` | Enable [RAG retrieval](features/embedding-rag.md) |
| `embedding.rag.maxResults` | `int` | | Maximum results per RAG query |
| `embedding.rag.collections` | `[]string` | | Collection names to search (empty = all) |

---

## Graph

> **Settings:** `lango settings` → Graph Store

```json
{
  "graph": {
    "enabled": false,
    "backend": "bolt",
    "databasePath": "~/.lango/graph.db",
    "maxTraversalDepth": 2,
    "maxExpansionResults": 10
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `graph.enabled` | `bool` | `false` | Enable the [knowledge graph](features/knowledge-graph.md) |
| `graph.backend` | `string` | `bolt` | Graph storage backend (`bolt`) |
| `graph.databasePath` | `string` | | Path to the graph database file |
| `graph.maxTraversalDepth` | `int` | `2` | Max depth for graph traversal in Graph RAG |
| `graph.maxExpansionResults` | `int` | `10` | Max results from graph expansion |

---

## Retrieval Coordinator

> **Settings:** `lango settings` → Retrieval / Auto-Adjust

The retrieval coordinator runs multiple search agents (FactSearch, TemporalSearch, ContextSearch) in parallel and merges results using evidence-based priority ranking.

```json
{
  "retrieval": {
    "enabled": true,
    "feedback": true,
    "autoAdjust": {
      "enabled": true,
      "mode": "shadow",
      "boostDelta": 0.05,
      "decayDelta": 0.01,
      "decayInterval": 100,
      "minScore": 0.1,
      "maxScore": 5.0,
      "warmupTurns": 50
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `retrieval.enabled` | `bool` | `false` | Enable multi-agent retrieval coordinator |
| `retrieval.feedback` | `bool` | `false` | Log context injection events for observability |
| `retrieval.autoAdjust.enabled` | `bool` | `false` | Enable relevance score auto-adjustment |
| `retrieval.autoAdjust.mode` | `string` | `shadow` | `shadow` (observe only) or `active` (apply changes) |
| `retrieval.autoAdjust.boostDelta` | `float64` | `0.05` | Score boost per context injection |
| `retrieval.autoAdjust.decayDelta` | `float64` | `0.01` | Score decay per interval |
| `retrieval.autoAdjust.decayInterval` | `int` | `100` | Turns between global decay |
| `retrieval.autoAdjust.minScore` | `float64` | `0.1` | Score floor |
| `retrieval.autoAdjust.maxScore` | `float64` | `5.0` | Score ceiling |
| `retrieval.autoAdjust.warmupTurns` | `int` | `50` | Turns before auto-adjust activates |

---

## Context Budget

> **Settings:** `lango settings` → Context Budget

Controls how the model's context window is allocated across prompt sections. Ratios must sum to 1.0 (tolerance: +/-0.001).

```json
{
  "context": {
    "modelWindow": 0,
    "responseReserve": 0,
    "allocation": {
      "knowledge": 0.30,
      "rag": 0.25,
      "memory": 0.25,
      "runSummary": 0.10,
      "headroom": 0.10
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `context.modelWindow` | `int` | `0` | Model context window in tokens (0 = auto-detect) |
| `context.responseReserve` | `int` | `0` | Tokens reserved for response (0 = use agent.maxTokens) |
| `context.allocation.knowledge` | `float64` | `0.30` | Knowledge section budget ratio |
| `context.allocation.rag` | `float64` | `0.25` | RAG section budget ratio |
| `context.allocation.memory` | `float64` | `0.25` | Memory section budget ratio |
| `context.allocation.runSummary` | `float64` | `0.10` | Run summary budget ratio |
| `context.allocation.headroom` | `float64` | `0.10` | Unallocated headroom ratio |

---

## A2A Protocol

!!! warning "Experimental"
    The A2A protocol is experimental. See [A2A Protocol](features/a2a-protocol.md).

> **Settings:** `lango settings` → A2A Protocol

```json
{
  "a2a": {
    "enabled": false,
    "baseUrl": "",
    "agentName": "",
    "agentDescription": "",
    "remoteAgents": [
      {
        "name": "code-reviewer",
        "agentCardUrl": "https://reviewer.example.com/.well-known/agent.json"
      }
    ]
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `a2a.enabled` | `bool` | `false` | Enable A2A protocol support |
| `a2a.baseUrl` | `string` | | External URL where this agent is reachable |
| `a2a.agentName` | `string` | | Name advertised in the Agent Card |
| `a2a.agentDescription` | `string` | | Description in the Agent Card |
| `a2a.remoteAgents` | `[]object` | | List of remote agents to connect to |

Each remote agent entry:

| Key | Type | Description |
|-----|------|-------------|
| `a2a.remoteAgents[].name` | `string` | Display name for the remote agent |
| `a2a.remoteAgents[].agentCardUrl` | `string` | URL to the remote agent's agent card |

---

## Payment

!!! warning "Experimental"
    The payment system is experimental. See [Payments](payments/index.md).

> **Settings:** `lango settings` → Payment

```json
{
  "payment": {
    "enabled": false,
    "walletProvider": "local",
    "network": {
      "chainId": 84532,
      "rpcUrl": "https://sepolia.base.org",
      "usdcContract": "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
    },
    "limits": {
      "maxPerTx": "1.00",
      "maxDaily": "10.00",
      "autoApproveBelow": ""
    },
    "x402": {
      "autoIntercept": false,
      "maxAutoPayAmount": ""
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `payment.enabled` | `bool` | `false` | Enable blockchain payment features |
| `payment.walletProvider` | `string` | `local` | Wallet backend: `local`, `rpc`, `composite` |
| `payment.network.chainId` | `int` | `84532` | EVM chain ID (84532 = Base Sepolia) |
| `payment.network.rpcUrl` | `string` | | JSON-RPC endpoint for the blockchain network |
| `payment.network.usdcContract` | `string` | | USDC token contract address |
| `payment.limits.maxPerTx` | `string` | `1.00` | Maximum USDC per transaction |
| `payment.limits.maxDaily` | `string` | `10.00` | Maximum daily USDC spending |
| `payment.limits.autoApproveBelow` | `string` | | Auto-approve payments below this amount |
| `payment.x402.autoIntercept` | `bool` | `false` | Enable X402 auto-interception for paid APIs |
| `payment.x402.maxAutoPayAmount` | `string` | | Maximum auto-pay amount for X402 requests |

---

## P2P Network

!!! warning "Experimental"
    The P2P networking system is experimental. See [P2P Network](features/p2p-network.md).

> **Settings:** `lango settings` → P2P Network

```json
{
  "p2p": {
    "enabled": false,
    "listenAddrs": ["/ip4/0.0.0.0/tcp/9000"],
    "bootstrapPeers": [],
    "keyDir": "~/.lango/p2p",
    "enableRelay": false,
    "enableMdns": true,
    "maxPeers": 50,
    "handshakeTimeout": "30s",
    "sessionTokenTtl": "1h",
    "autoApproveKnownPeers": false,
    "requireSignedChallenge": false,
    "firewallRules": [],
    "gossipInterval": "30s",
    "zkHandshake": false,
    "zkAttestation": false,
    "zkp": {
      "proofCacheDir": "~/.lango/zkp",
      "provingScheme": "plonk",
      "srsMode": "unsafe",
      "srsPath": "",
      "maxCredentialAge": "24h"
    },
    "toolIsolation": {
      "enabled": false,
      "timeoutPerTool": "30s",
      "maxMemoryMB": 512,
      "container": {
        "enabled": false,
        "runtime": "auto",
        "image": "lango-sandbox:latest",
        "networkMode": "none",
        "readOnlyRootfs": true,
        "poolSize": 0,
        "poolIdleTimeout": "5m"
      }
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `p2p.enabled` | `bool` | `false` | Enable P2P networking |
| `p2p.listenAddrs` | `[]string` | `["/ip4/0.0.0.0/tcp/9000"]` | Multiaddrs to listen on |
| `p2p.bootstrapPeers` | `[]string` | `[]` | Initial peers for DHT bootstrapping |
| `p2p.keyDir` | `string` | `~/.lango/p2p` | Directory for node key persistence |
| `p2p.enableRelay` | `bool` | `true` | Act as relay for NAT traversal |
| `p2p.enableMdns` | `bool` | `true` | Enable mDNS for LAN discovery |
| `p2p.maxPeers` | `int` | `50` | Maximum connected peers |
| `p2p.handshakeTimeout` | `duration` | `30s` | Maximum handshake duration |
| `p2p.sessionTokenTtl` | `duration` | `1h` | Session token lifetime |
| `p2p.autoApproveKnownPeers` | `bool` | `false` | Skip approval for known peers |
| `p2p.firewallRules` | `[]object` | `[]` | Static firewall ACL rules |
| `p2p.gossipInterval` | `duration` | `30s` | Agent card gossip interval |
| `p2p.zkHandshake` | `bool` | `false` | Enable ZK-enhanced handshake |
| `p2p.zkAttestation` | `bool` | `false` | Enable ZK attestation on responses |
| `p2p.requireSignedChallenge` | `bool` | `false` | Reject unsigned (v1.0) challenges; require v1.1 signed challenges |
| `p2p.zkp.proofCacheDir` | `string` | `~/.lango/zkp` | ZKP circuit cache directory |
| `p2p.zkp.provingScheme` | `string` | `plonk` | ZKP proving scheme: `plonk` or `groth16` |
| `p2p.zkp.srsMode` | `string` | `unsafe` | SRS generation mode: `unsafe` (deterministic) or `file` (trusted ceremony) |
| `p2p.zkp.srsPath` | `string` | | Path to SRS file (when `srsMode = "file"`) |
| `p2p.zkp.maxCredentialAge` | `string` | `24h` | Maximum age for ZK credentials before rejection |

Each firewall rule entry:

| Key | Type | Description |
|-----|------|-------------|
| `firewallRules[].peerDid` | `string` | Peer DID (`"*"` for all peers) |
| `firewallRules[].action` | `string` | `"allow"` or `"deny"` |
| `firewallRules[].tools` | `[]string` | Tool name patterns (empty = all) |
| `firewallRules[].rateLimit` | `int` | Max requests/min (0 = unlimited) |

### P2P Pricing

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `p2p.pricing.enabled` | `bool` | `false` | Enable paid P2P tool invocations |
| `p2p.pricing.perQuery` | `string` | | Default price per query in USDC (e.g., `"0.10"`) |
| `p2p.pricing.toolPrices` | `map[string]string` | | Map of tool names to specific prices in USDC |
| `p2p.pricing.trustThresholds.postPayMinScore` | `float64` | `0.8` | Minimum reputation score for post-pay pricing tier |

### P2P Settlement

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `p2p.pricing.settlement.receiptTimeout` | `duration` | `2m` | Maximum wait for on-chain receipt confirmation |
| `p2p.pricing.settlement.maxRetries` | `int` | `3` | Maximum transaction submission retries |

### P2P Owner Protection

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `p2p.ownerProtection.ownerName` | `string` | | Owner name to block from P2P responses |
| `p2p.ownerProtection.ownerEmail` | `string` | | Owner email to block from P2P responses |
| `p2p.ownerProtection.ownerPhone` | `string` | | Owner phone to block from P2P responses |
| `p2p.ownerProtection.extraTerms` | `[]string` | | Additional terms to block from P2P responses |
| `p2p.ownerProtection.blockConversations` | `bool` | `true` | Block conversation data in P2P responses |

### P2P Reputation

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `p2p.minTrustScore` | `float64` | `0.3` | Minimum trust score to accept P2P requests (0.0 - 1.0) |

### P2P Tool Isolation

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `p2p.toolIsolation.enabled` | `bool` | `false` | Enable subprocess isolation for remote peer tool invocations |
| `p2p.toolIsolation.timeoutPerTool` | `duration` | `30s` | Maximum duration for a single tool execution |
| `p2p.toolIsolation.maxMemoryMB` | `int` | `256` | Soft memory limit per subprocess in megabytes |
| `p2p.toolIsolation.container.enabled` | `bool` | `false` | Use container-based sandbox instead of subprocess |
| `p2p.toolIsolation.container.runtime` | `string` | `auto` | Container runtime: `auto`, `docker`, `gvisor`, `native` |
| `p2p.toolIsolation.container.image` | `string` | `lango-sandbox:latest` | Docker image for sandbox container |
| `p2p.toolIsolation.container.networkMode` | `string` | `none` | Docker network mode for sandbox containers |
| `p2p.toolIsolation.container.readOnlyRootfs` | `bool` | `true` | Mount container root filesystem as read-only |
| `p2p.toolIsolation.container.cpuQuotaUs` | `int` | `0` | Docker CPU quota in microseconds (0 = unlimited) |
| `p2p.toolIsolation.container.poolSize` | `int` | `0` | Pre-warmed containers in pool (0 = disabled) |
| `p2p.toolIsolation.container.poolIdleTimeout` | `duration` | `5m` | Idle timeout before pool containers are recycled |

---

## Economy

!!! warning "Experimental"
    The P2P economy layer is experimental. See [P2P Economy](features/economy.md).

> **Settings:** `lango settings` → Economy

```json
{
  "economy": {
    "enabled": false,
    "budget": {
      "defaultMax": "10.00",
      "alertThresholds": [0.5, 0.8, 0.95],
      "hardLimit": true
    },
    "risk": {
      "escrowThreshold": "5.00",
      "highTrustScore": 0.8,
      "mediumTrustScore": 0.5
    },
    "negotiate": {
      "enabled": false,
      "maxRounds": 5,
      "timeout": "5m",
      "autoNegotiate": false,
      "maxDiscount": 0.2
    },
    "escrow": {
      "enabled": false,
      "defaultTimeout": "24h",
      "maxMilestones": 10,
      "autoRelease": false,
      "disputeWindow": "1h",
      "settlement": {
        "receiptTimeout": "2m",
        "maxRetries": 3
      },
      "onChain": {
        "enabled": false,
        "mode": "hub",
        "hubAddress": "",
        "vaultFactoryAddress": "",
        "vaultImplementation": "",
        "arbitratorAddress": "",
        "tokenAddress": "",
        "pollInterval": "15s"
      }
    },
    "pricing": {
      "enabled": false,
      "trustDiscount": 0.1,
      "volumeDiscount": 0.05,
      "minPrice": "0.01"
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `economy.enabled` | `bool` | `false` | Enable the P2P economy layer |
| `economy.budget.defaultMax` | `string` | `10.00` | Default maximum budget per task in USDC |
| `economy.budget.alertThresholds` | `[]float64` | `[0.5, 0.8, 0.95]` | Budget usage percentages that trigger alerts |
| `economy.budget.hardLimit` | `bool` | `true` | Enforce budget as a hard cap (reject overspend) |
| `economy.risk.escrowThreshold` | `string` | `5.00` | USDC amount above which escrow is forced |
| `economy.risk.highTrustScore` | `float64` | `0.8` | Minimum trust score for DirectPay strategy |
| `economy.risk.mediumTrustScore` | `float64` | `0.5` | Minimum trust score for non-ZK strategies |
| `economy.negotiate.enabled` | `bool` | `false` | Enable P2P negotiation protocol |
| `economy.negotiate.maxRounds` | `int` | `5` | Maximum counter-offers per negotiation |
| `economy.negotiate.timeout` | `duration` | `5m` | Negotiation session timeout |
| `economy.negotiate.autoNegotiate` | `bool` | `false` | Auto-generate counter-offers |
| `economy.negotiate.maxDiscount` | `float64` | `0.2` | Maximum discount for auto-negotiation (0-1) |
| `economy.escrow.enabled` | `bool` | `false` | Enable milestone-based escrow |
| `economy.escrow.defaultTimeout` | `duration` | `24h` | Escrow expiration timeout |
| `economy.escrow.maxMilestones` | `int` | `10` | Maximum milestones per escrow |
| `economy.escrow.autoRelease` | `bool` | `false` | Auto-release funds when all milestones met |
| `economy.escrow.disputeWindow` | `duration` | `1h` | Time window for disputes after completion |
| `economy.escrow.settlement.receiptTimeout` | `duration` | `2m` | Max wait for on-chain receipt confirmation |
| `economy.escrow.settlement.maxRetries` | `int` | `3` | Max transaction submission retries |
| `economy.escrow.onChain.enabled` | `bool` | `false` | Enable on-chain escrow mode |
| `economy.escrow.onChain.mode` | `string` | `hub` | On-chain escrow pattern: `hub` or `vault` |
| `economy.escrow.onChain.hubAddress` | `string` | | Deployed LangoEscrowHub contract address |
| `economy.escrow.onChain.vaultFactoryAddress` | `string` | | Deployed LangoVaultFactory contract address |
| `economy.escrow.onChain.vaultImplementation` | `string` | | LangoVault implementation address for cloning |
| `economy.escrow.onChain.arbitratorAddress` | `string` | | Dispute arbitrator address |
| `economy.escrow.onChain.tokenAddress` | `string` | | ERC-20 token (USDC) contract address |
| `economy.escrow.onChain.pollInterval` | `duration` | `15s` | Event monitor polling interval |
| `economy.pricing.enabled` | `bool` | `false` | Enable dynamic pricing adjustments |
| `economy.pricing.trustDiscount` | `float64` | `0.1` | Max discount for high-trust peers (0-1) |
| `economy.pricing.volumeDiscount` | `float64` | `0.05` | Max discount for high-volume peers (0-1) |
| `economy.pricing.minPrice` | `string` | `0.01` | Minimum price floor in USDC |

---

## Smart Account

!!! warning "Experimental"
    Smart Account support is experimental. See [Smart Accounts](features/smart-accounts.md).

> **Settings:** `lango settings` → Smart Account / SA Session Keys / SA Paymaster / SA Modules

```json
{
  "smartAccount": {
    "enabled": false,
    "factoryAddress": "",
    "entryPointAddress": "",
    "safe7579Address": "",
    "fallbackHandler": "",
    "bundlerURL": "",
    "session": {
      "maxDuration": "24h",
      "defaultGasLimit": 500000,
      "maxActiveKeys": 10
    },
    "paymaster": {
      "enabled": false,
      "provider": "circle",
      "rpcURL": "",
      "tokenAddress": "",
      "paymasterAddress": "",
      "policyId": "",
      "fallbackMode": "abort"
    },
    "modules": {
      "sessionValidatorAddress": "",
      "spendingHookAddress": "",
      "escrowExecutorAddress": ""
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `smartAccount.enabled` | `bool` | `false` | Enable ERC-7579 smart account subsystem |
| `smartAccount.factoryAddress` | `string` | | Safe factory contract address |
| `smartAccount.entryPointAddress` | `string` | | ERC-4337 EntryPoint contract address |
| `smartAccount.safe7579Address` | `string` | | Safe7579 adapter contract address |
| `smartAccount.fallbackHandler` | `string` | | Safe fallback handler contract address |
| `smartAccount.bundlerURL` | `string` | | ERC-4337 bundler RPC endpoint URL |
| `smartAccount.session.maxDuration` | `duration` | `24h` | Maximum allowed session key duration |
| `smartAccount.session.defaultGasLimit` | `uint64` | `500000` | Default gas limit for session key operations |
| `smartAccount.session.maxActiveKeys` | `int` | `10` | Maximum number of active session keys |
| `smartAccount.paymaster.enabled` | `bool` | `false` | Enable paymaster for gasless transactions |
| `smartAccount.paymaster.provider` | `string` | `circle` | Paymaster provider (`circle`, `pimlico`, `alchemy`) |
| `smartAccount.paymaster.mode` | `string` | `rpc` | Paymaster mode: `rpc` (API-based) or `permit` (on-chain EIP-2612) |
| `smartAccount.paymaster.rpcURL` | `string` | | Paymaster provider RPC endpoint (required for `rpc` mode) |
| `smartAccount.paymaster.tokenAddress` | `string` | | USDC token contract address |
| `smartAccount.paymaster.paymasterAddress` | `string` | | Paymaster contract address |
| `smartAccount.paymaster.policyId` | `string` | | Provider-specific policy ID (optional) |
| `smartAccount.paymaster.fallbackMode` | `string` | `abort` | Behavior when paymaster fails (`abort`, `direct`) |
| `smartAccount.modules.sessionValidatorAddress` | `string` | | LangoSessionValidator module contract address |
| `smartAccount.modules.spendingHookAddress` | `string` | | LangoSpendingHook module contract address |
| `smartAccount.modules.escrowExecutorAddress` | `string` | | LangoEscrowExecutor module contract address |

---

## Observability

!!! warning "Experimental"
    The observability system is experimental. See [Observability](features/observability.md).

> **Settings:** `lango settings` → Observability

```json
{
  "observability": {
    "enabled": false,
    "tokens": {
      "enabled": true,
      "persistHistory": false,
      "retentionDays": 30
    },
    "health": {
      "enabled": true,
      "interval": "30s"
    },
    "audit": {
      "enabled": false,
      "retentionDays": 90
    },
    "metrics": {
      "enabled": true,
      "format": "json"
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `observability.enabled` | `bool` | `false` | Enable the observability subsystem |
| `observability.tokens.enabled` | `bool` | `true` | Enable token usage tracking |
| `observability.tokens.persistHistory` | `bool` | `false` | Persist token usage to database |
| `observability.tokens.retentionDays` | `int` | `30` | Days to retain token usage records |
| `observability.health.enabled` | `bool` | `true` | Enable health check monitoring |
| `observability.health.interval` | `duration` | `30s` | Health check probe interval |
| `observability.audit.enabled` | `bool` | `false` | Enable audit logging |
| `observability.audit.retentionDays` | `int` | `90` | Days to retain audit records |
| `observability.metrics.enabled` | `bool` | `true` | Enable metrics export endpoint |
| `observability.metrics.format` | `string` | `json` | Metrics export format (currently only `json` is implemented) |

---

## Cron

See [Cron Scheduling](automation/cron.md) for usage details and [CLI reference](cli/automation.md#cron-commands).

> **Settings:** `lango settings` → Cron Scheduler

```json
{
  "cron": {
    "enabled": false,
    "timezone": "UTC",
    "maxConcurrentJobs": 5,
    "defaultSessionMode": "isolated",
    "historyRetention": "720h",
    "defaultDeliverTo": []
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `cron.enabled` | `bool` | `false` | Enable the cron scheduling system |
| `cron.timezone` | `string` | `UTC` | Default timezone for cron expressions |
| `cron.maxConcurrentJobs` | `int` | `5` | Maximum concurrently executing jobs |
| `cron.defaultSessionMode` | `string` | `isolated` | Default session mode: `isolated` or `main` |
| `cron.historyRetention` | `duration` | `720h` | How long to retain execution history (30 days) |
| `cron.defaultDeliverTo` | `[]string` | `[]` | Default delivery channels for job results |

---

## Background

!!! warning "Experimental"
    Background tasks are experimental. See [Background Tasks](automation/background.md).

> **Settings:** `lango settings` → Background Tasks

```json
{
  "background": {
    "enabled": false,
    "yieldMs": 30000,
    "maxConcurrentTasks": 3,
    "defaultDeliverTo": []
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `background.enabled` | `bool` | `false` | Enable the background task system |
| `background.yieldMs` | `int` | `30000` | Auto-yield threshold in milliseconds |
| `background.maxConcurrentTasks` | `int` | `3` | Maximum concurrently running tasks |
| `background.defaultDeliverTo` | `[]string` | `[]` | Default delivery channels for task results |

---

## Workflow

!!! warning "Experimental"
    The workflow engine is experimental. See [Workflow Engine](automation/workflows.md) and [CLI reference](cli/automation.md#workflow-commands).

> **Settings:** `lango settings` → Workflow Engine

```json
{
  "workflow": {
    "enabled": false,
    "maxConcurrentSteps": 4,
    "defaultTimeout": "10m",
    "stateDir": "~/.lango/workflows/",
    "defaultDeliverTo": []
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `workflow.enabled` | `bool` | `false` | Enable the workflow engine |
| `workflow.maxConcurrentSteps` | `int` | `4` | Maximum steps running in parallel |
| `workflow.defaultTimeout` | `duration` | `10m` | Default timeout per workflow step |
| `workflow.stateDir` | `string` | `~/.lango/workflows/` | Directory for workflow state files |
| `workflow.defaultDeliverTo` | `[]string` | `[]` | Default delivery channels for workflow results |

---

## Librarian

!!! warning "Experimental"
    The Proactive Librarian is experimental. See [Proactive Librarian](features/librarian.md).

> **Settings:** `lango settings` → Librarian

```json
{
  "librarian": {
    "enabled": false,
    "observationThreshold": 2,
    "inquiryCooldownTurns": 3,
    "maxPendingInquiries": 2,
    "autoSaveConfidence": "high",
    "provider": "",
    "model": ""
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `librarian.enabled` | `bool` | `false` | Enable the Proactive Librarian |
| `librarian.observationThreshold` | `int` | `2` | Observations needed before triggering inquiry |
| `librarian.inquiryCooldownTurns` | `int` | `3` | Minimum turns between inquiries |
| `librarian.maxPendingInquiries` | `int` | `2` | Maximum pending inquiries at once |
| `librarian.autoSaveConfidence` | `string` | `high` | Confidence level for auto-saving: `low`, `medium`, `high` |
| `librarian.provider` | `string` | | AI provider for librarian (empty = agent default) |
| `librarian.model` | `string` | | Model for librarian (empty = agent default) |

---

## RunLedger

!!! warning "Experimental"
    The RunLedger (Task OS) is experimental. It progresses through shadow, write-through, and authoritative-read adoption phases.

> **Settings:** `lango settings` → RunLedger

```json
{
  "runLedger": {
    "enabled": false,
    "shadow": true,
    "writeThrough": false,
    "authoritativeRead": false,
    "workspaceIsolation": false,
    "staleTtl": "1h",
    "maxRunHistory": 0,
    "validatorTimeout": "2m",
    "plannerMaxRetries": 2
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `runLedger.enabled` | `bool` | `false` | Activate the RunLedger system |
| `runLedger.shadow` | `bool` | `true` | Shadow mode: journal records only, existing systems unaffected |
| `runLedger.writeThrough` | `bool` | `false` | All creates/updates go through ledger first, then mirror to legacy stores |
| `runLedger.authoritativeRead` | `bool` | `false` | State reads come from ledger snapshots only |
| `runLedger.workspaceIsolation` | `bool` | `false` | Enable runtime PEV workspace wiring for coding-step validation |
| `runLedger.staleTtl` | `duration` | `1h` | How long a paused run remains resumable |
| `runLedger.maxRunHistory` | `int` | `0` | Maximum number of runs to keep (0 = unlimited) |
| `runLedger.validatorTimeout` | `duration` | `2m` | Timeout for individual validator execution |
| `runLedger.plannerMaxRetries` | `int` | `2` | How many times a malformed planner output is retried |

---

## Provenance

!!! warning "Experimental"
    The provenance system is experimental. It provides session-level checkpoint tracking for auditability and replay.

> **Settings:** `lango settings` → Provenance

```json
{
  "provenance": {
    "enabled": false,
    "checkpoints": {
      "autoOnStepComplete": false,
      "autoOnPolicy": false,
      "maxPerSession": 0,
      "retentionDays": 0
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `provenance.enabled` | `bool` | `false` | Activate the provenance system |
| `provenance.checkpoints.autoOnStepComplete` | `bool` | `false` | Create a checkpoint when a RunLedger step passes validation |
| `provenance.checkpoints.autoOnPolicy` | `bool` | `false` | Create a checkpoint when a policy decision is applied |
| `provenance.checkpoints.maxPerSession` | `int` | `0` | Maximum checkpoints per session (0 = unlimited) |
| `provenance.checkpoints.retentionDays` | `int` | `0` | Days to keep checkpoints before pruning (0 = unlimited) |

---

## Sandbox

!!! warning "Experimental"
    The OS-level sandbox is experimental. It applies to child processes spawned by exec tools, MCP stdio servers, and skill scripts. Independent of `p2p.toolIsolation`.

> **Settings:** `lango settings` → Sandbox

```json
{
  "sandbox": {
    "enabled": false,
    "failClosed": false,
    "workspacePath": "",
    "networkMode": "deny",
    "allowedNetworkIPs": [],
    "allowedWritePaths": [],
    "timeoutPerTool": "30s",
    "os": {
      "seccompProfile": "moderate",
      "seatbeltCustomProfile": ""
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `sandbox.enabled` | `bool` | `false` | Enable OS-level sandboxing for tool-spawned child processes |
| `sandbox.failClosed` | `bool` | `false` | Reject tool execution when OS sandbox is unavailable (false = fail-open) |
| `sandbox.workspacePath` | `string` | `""` | Root directory for workspace-relative write access (empty = CWD) |
| `sandbox.networkMode` | `string` | `deny` | Network access from sandboxed processes: `deny` or `allow` |
| `sandbox.allowedNetworkIPs` | `[]string` | `[]` | IP addresses permitted for outbound connections (macOS Seatbelt only; ignored on Linux) |
| `sandbox.allowedWritePaths` | `[]string` | `[]` | Additional paths writable from the sandbox beyond `workspacePath` |
| `sandbox.timeoutPerTool` | `duration` | `30s` | Maximum duration for a single sandboxed tool execution |
| `sandbox.os.seccompProfile` | `string` | `moderate` | Seccomp filter profile on Linux: `strict`, `moderate`, or `permissive` |
| `sandbox.os.seatbeltCustomProfile` | `string` | `""` | Path to a custom `.sb` profile on macOS (overrides generated profile) |

---

## Gatekeeper

Response sanitization (output gatekeeper) settings. The gatekeeper strips internal markers, thought tags, and large raw JSON from agent responses before they reach the user.

> **Settings:** `lango settings` → Gatekeeper

```json
{
  "gatekeeper": {
    "enabled": true,
    "stripThoughtTags": true,
    "stripInternalMarkers": true,
    "stripRawJSON": true,
    "rawJsonThreshold": 500,
    "customPatterns": []
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `gatekeeper.enabled` | `bool` | `true` | Enable response sanitization |
| `gatekeeper.stripThoughtTags` | `bool` | `true` | Strip `<thought>`/`<thinking>` tags from responses |
| `gatekeeper.stripInternalMarkers` | `bool` | `true` | Strip lines starting with `[INTERNAL]`, `[DEBUG]`, `[SYSTEM]`, `[OBSERVATION]` |
| `gatekeeper.stripRawJSON` | `bool` | `true` | Replace large raw JSON code blocks with a placeholder |
| `gatekeeper.rawJsonThreshold` | `int` | `500` | Character threshold for raw JSON replacement |
| `gatekeeper.customPatterns` | `[]string` | `[]` | Additional regex patterns to strip from responses |

---

## MCP

MCP (Model Context Protocol) server integration for connecting to external tool servers.

> **Settings:** `lango settings` → MCP

```json
{
  "mcp": {
    "enabled": false,
    "defaultTimeout": "30s",
    "maxOutputTokens": 25000,
    "healthCheckInterval": "30s",
    "autoReconnect": true,
    "maxReconnectAttempts": 5,
    "servers": {
      "my-server": {
        "transport": "stdio",
        "command": "npx",
        "args": ["-y", "@modelcontextprotocol/server-filesystem"],
        "env": {},
        "enabled": true,
        "timeout": "",
        "safetyLevel": "dangerous"
      }
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mcp.enabled` | `bool` | `false` | Enable MCP server integration |
| `mcp.defaultTimeout` | `duration` | `30s` | Default timeout for MCP operations |
| `mcp.maxOutputTokens` | `int` | `25000` | Maximum output size from MCP tool calls |
| `mcp.healthCheckInterval` | `duration` | `30s` | Interval for periodic server health probes |
| `mcp.autoReconnect` | `bool` | `true` | Automatically reconnect on connection loss |
| `mcp.maxReconnectAttempts` | `int` | `5` | Maximum reconnection attempts before giving up |

Each server entry (`mcp.servers.<name>`):

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `mcp.servers.<name>.transport` | `string` | `stdio` | Transport type: `stdio`, `http`, `sse` |
| `mcp.servers.<name>.command` | `string` | | Executable for stdio transport |
| `mcp.servers.<name>.args` | `[]string` | `[]` | Command-line arguments for stdio transport |
| `mcp.servers.<name>.env` | `map[string]string` | `{}` | Environment variables for stdio transport (supports `${VAR}` expansion) |
| `mcp.servers.<name>.url` | `string` | | Endpoint URL for http/sse transport |
| `mcp.servers.<name>.headers` | `map[string]string` | `{}` | HTTP headers for http/sse transport (supports `${VAR}` expansion) |
| `mcp.servers.<name>.enabled` | `bool` | `true` | Whether this server is active |
| `mcp.servers.<name>.timeout` | `duration` | | Override the global default timeout for this server |
| `mcp.servers.<name>.safetyLevel` | `string` | `dangerous` | Tool safety level: `safe`, `moderate`, `dangerous` |

---

## Orchestration

!!! warning "Experimental"
    Structured orchestration is experimental. It wraps the agent executor with delegation guard, budget policy, and recovery policy.

> **Settings:** `lango settings` → Orchestration

```json
{
  "agent": {
    "orchestration": {
      "mode": "classic",
      "circuitBreaker": {
        "failureThreshold": 3,
        "resetTimeout": "30s"
      },
      "budget": {
        "toolCallLimit": 50,
        "delegationLimit": 15,
        "alertThreshold": 0.8
      },
      "recovery": {
        "maxRetries": 2,
        "circuitBreakerCooldown": "5m"
      }
    }
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `agent.orchestration.mode` | `string` | `classic` | Orchestration mode: `classic` (default) or `structured` |
| `agent.orchestration.circuitBreaker.failureThreshold` | `int` | `3` | Consecutive failures before circuit opens |
| `agent.orchestration.circuitBreaker.resetTimeout` | `duration` | `30s` | Time before half-open probe |
| `agent.orchestration.budget.toolCallLimit` | `int` | `50` | Maximum tool calls per agent run |
| `agent.orchestration.budget.delegationLimit` | `int` | `15` | Maximum delegations before alerting |
| `agent.orchestration.budget.alertThreshold` | `float64` | `0.8` | Budget usage percentage at which alerts fire |
| `agent.orchestration.recovery.maxRetries` | `int` | `2` | Maximum retry attempts on failure |
| `agent.orchestration.recovery.circuitBreakerCooldown` | `duration` | `5m` | Time before re-enabling a tripped agent |

---

## Environment Variable Substitution

String configuration values support `${ENV_VAR}` syntax for environment variable substitution. This is useful for sensitive values like API keys and tokens:

```json
{
  "providers": {
    "my-provider": {
      "type": "anthropic",
      "apiKey": "${ANTHROPIC_API_KEY}"
    }
  },
  "channels": {
    "telegram": {
      "botToken": "${TELEGRAM_BOT_TOKEN}"
    }
  }
}
```
