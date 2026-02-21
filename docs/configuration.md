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
    "multiAgent": false
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
| `session.maxHistoryTurns` | `int` | | Maximum conversation turns to retain per session |

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
| `security.interceptor.approvalPolicy` | `string` | `dangerous` | Tool approval policy: `always`, `dangerous`, `never` |
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
| `security.interceptor.presidio.url` | `string` | | Presidio analyzer service URL |
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
| `tools.exec.defaultTimeout` | `duration` | | Default timeout for shell command execution |
| `tools.exec.allowBackground` | `bool` | `true` | Allow background command execution |
| `tools.exec.workDir` | `string` | | Working directory for command execution |

### Filesystem Tool

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `tools.filesystem.maxReadSize` | `int` | | Maximum file read size in bytes |
| `tools.filesystem.allowedPaths` | `[]string` | | Allowed filesystem paths (empty = all) |

### Browser Tool

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `tools.browser.enabled` | `bool` | `false` | Enable browser automation tool |
| `tools.browser.headless` | `bool` | `true` | Run browser in headless mode |
| `tools.browser.sessionTimeout` | `duration` | `5m` | Browser session timeout |

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
