# Lango 🚀

A high-performance AI agent built with Go, supporting multiple AI providers, channels (Telegram, Discord, Slack), and a self-learning knowledge system.

## Features

- 🔥 **Fast** - Single binary, <100ms startup, <100MB memory
- 🤖 **Multi-Provider AI** - OpenAI, Anthropic, Gemini, Ollama with unified interface
- 🔌 **Multi-Channel** - Telegram, Discord, Slack support
- 🛠️ **Rich Tools** - Shell execution, file system operations, browser automation, crypto & secrets tools
- 🧠 **Self-Learning** - Knowledge store, learning engine, skill system, observational memory
- 🔒 **Secure** - AES-256-GCM encryption, key registry, secret management, output scanning
- 💾 **Persistent** - Ent ORM with SQLite session storage
- 🌐 **Gateway** - WebSocket/HTTP server for control plane
- 🔑 **Auth** - OIDC authentication, OAuth login flow

## Quick Start

### Installation

```bash
# Build from source
git clone https://github.com/langowarny/lango.git
cd lango
make build

# Or install directly
go install github.com/langowarny/lango/cmd/lango@latest
```

### Configuration

Create `lango.json`:

```json
{
  "server": {
    "host": "localhost",
    "port": 18789
  },
  "agent": {
    "provider": "gemini",
    "model": "gemini-2.0-flash-exp"
  },
  "providers": {
    "gemini": {
      "apiKey": "${GOOGLE_API_KEY}"
    }
  },
  "channels": {
    "telegram": {
      "enabled": true,
      "botToken": "${TELEGRAM_BOT_TOKEN}"
    }
  },
  "logging": {
    "level": "info",
    "format": "console"
  }
}
```

### Run

```bash
# Start the server (ensure GOOGLE_API_KEY is set)
export GOOGLE_API_KEY=your_key_here
lango serve

# Or with custom config
lango serve --config /path/to/lango.json

# Validate configuration
lango config validate
```

### Getting Started

Use the interactive onboard wizard for first-time setup:

```bash
lango onboard
```

This guides you through:
1. AI provider configuration (API keys, models)
2. Server and channel setup (Telegram, Discord, Slack)
3. Security settings (encryption, signer mode)
4. Tool configuration

### Diagnostics

Run the doctor command to check your setup:

```bash
# Check configuration and environment
lango doctor

# Auto-fix common issues
lango doctor --fix

# JSON output for scripting
lango doctor --json
```

## Architecture

```
lango/
├── cmd/lango/              # CLI entry point (cobra)
├── internal/
│   ├── adk/                # Google ADK agent wrapper, session/state adapters
│   ├── agent/              # Agent types, PII redactor, secret scanner
│   ├── app/                # Application bootstrap, wiring, tool registration
│   ├── channels/           # Telegram, Discord, Slack integrations
│   ├── cli/                # CLI commands
│   │   ├── auth/           #   lango login (OIDC/OAuth)
│   │   ├── doctor/         #   lango doctor (diagnostics)
│   │   ├── onboard/        #   lango onboard (TUI wizard)
│   │   └── security/       #   lango security migrate-passphrase
│   ├── config/             # Config loading, env var substitution, validation
│   ├── ent/                # Ent ORM schemas and generated code
│   ├── gateway/            # WebSocket/HTTP server, OIDC auth
│   ├── knowledge/          # Knowledge store, 8-layer context retriever
│   ├── learning/           # Learning engine, error pattern analyzer
│   ├── logging/            # Zap structured logger
│   ├── memory/             # Observational memory (observer, reflector, token counter)
│   ├── provider/           # AI provider interface and implementations
│   │   ├── anthropic/      #   Claude models
│   │   ├── gemini/         #   Google Gemini models
│   │   └── openai/         #   OpenAI-compatible (GPT, Ollama, etc.)
│   ├── security/           # Crypto providers, key registry, secrets store, companion discovery
│   ├── session/            # Ent-based SQLite session store
│   ├── skill/              # Skill registry, executor, builder
│   ├── supervisor/         # Provider proxy, privileged tool execution
│   └── tools/              # exec, filesystem, browser
└── openspec/               # Specifications (OpenSpec workflow)
```

## AI Providers

Lango supports multiple AI providers with a unified interface. Provider aliases are resolved automatically (e.g., `gpt`/`chatgpt` -> `openai`, `claude` -> `anthropic`, `llama` -> `ollama`, `bard` -> `gemini`).

### Supported Providers
- **OpenAI** (`openai`): GPT-4o, GPT-4, and OpenAI-compatible APIs
- **Anthropic** (`anthropic`): Claude 3, Claude 3.5
- **Gemini** (`gemini`): Google Gemini models
- **Ollama** (`ollama`): Local models via Ollama (default: `http://localhost:11434/v1`)

### Configuration Example

```json
{
  "agent": {
    "provider": "openai",
    "model": "gpt-4o",
    "fallbackProvider": "anthropic",
    "fallbackModel": "claude-3-5-sonnet-20241022"
  },
  "providers": {
    "openai": {
      "apiKey": "${OPENAI_API_KEY}"
    },
    "anthropic": {
      "apiKey": "${ANTHROPIC_API_KEY}"
    },
    "ollama": {
      "baseUrl": "http://localhost:11434/v1"
    }
  },
  "security": {
    "interceptor": {
      "enabled": true,
      "redactPii": true
    },
    "signer": {
      "provider": "local"
    }
  }
}
```

### Onboarding TUI
Use `lango onboard` to interactively configure providers, models, and security settings. The TUI allows you to manage multiple providers and set up local encryption.

## Configuration Reference

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| **Server** | | | |
| `server.host` | string | `localhost` | Bind address |
| `server.port` | int | `18789` | Listen port |
| `server.httpEnabled` | bool | `false` | Enable HTTP API endpoints |
| `server.wsEnabled` | bool | `false` | Enable WebSocket server |
| **Agent** | | | |
| `agent.provider` | string | `gemini` | Primary AI provider ID |
| `agent.model` | string | `gemini-2.0-flash-exp` | Primary model ID |
| `agent.fallbackProvider` | string | - | Fallback provider ID |
| `agent.fallbackModel` | string | - | Fallback model ID |
| `agent.maxTokens` | int | `4096` | Max tokens |
| `agent.temperature` | float | `0` | Generation temperature |
| `agent.systemPromptPath` | string | - | Custom system prompt template path |
| **Providers** | | | |
| `providers.<id>.type` | string | - | Provider type (openai, anthropic, gemini) |
| `providers.<id>.apiKey` | string | - | Provider API key (supports `${ENV_VAR}`) |
| `providers.<id>.baseUrl` | string | - | Custom base URL (e.g. for Ollama) |
| **Logging** | | | |
| `logging.level` | string | `info` | Log level |
| `logging.format` | string | `console` | `json` or `console` |
| **Session** | | | |
| `session.databasePath` | string | `~/.lango/sessions.db` | SQLite path |
| `session.ttl` | duration | - | Session TTL before expiration |
| `session.maxHistoryTurns` | int | - | Maximum history turns per session |
| **Security** | | | |
| `security.signer.provider` | string | `local` | `local`, `rpc`, or `enclave` |
| `security.passphrase` | string | - | **DEPRECATED** Use `LANGO_PASSPHRASE` env var |
| `security.interceptor.enabled` | bool | `false` | Enable AI Privacy Interceptor |
| `security.interceptor.redactPii` | bool | `false` | Redact PII from AI interactions |
| `security.interceptor.approvalRequired` | bool | `false` | Require approval for sensitive tool use |
| **Tools** | | | |
| `tools.exec.defaultTimeout` | duration | - | Default timeout for shell commands |
| `tools.exec.allowBackground` | bool | `false` | Allow background processes |
| `tools.exec.workDir` | string | - | Working directory (empty = current) |
| `tools.filesystem.maxReadSize` | int | - | Maximum file size to read |
| `tools.filesystem.allowedPaths` | []string | - | Allowed paths (empty = allow all) |
| `tools.browser.enabled` | bool | `false` | Enable browser automation tools (requires Chromium) |
| `tools.browser.headless` | bool | `true` | Run browser in headless mode |
| `tools.browser.sessionTimeout` | duration | `5m` | Browser session timeout |
| **Knowledge** | | | |
| `knowledge.enabled` | bool | `false` | Enable self-learning knowledge system |
| `knowledge.maxLearnings` | int | - | Max learning entries per session |
| `knowledge.maxKnowledge` | int | - | Max knowledge entries per session |
| `knowledge.maxContextPerLayer` | int | - | Max context items per layer in retrieval |
| `knowledge.autoApproveSkills` | bool | `false` | Auto-approve new skills |
| `knowledge.maxSkillsPerDay` | int | - | Rate limit for skill creation |
| **Observational Memory** | | | |
| `observationalMemory.enabled` | bool | `false` | Enable observational memory system |
| `observationalMemory.provider` | string | - | LLM provider for observer/reflector (empty = agent default) |
| `observationalMemory.model` | string | - | Model for observer/reflector (empty = agent default) |
| `observationalMemory.messageTokenThreshold` | int | `1000` | Token threshold to trigger observation |
| `observationalMemory.observationTokenThreshold` | int | `2000` | Token threshold to trigger reflection |
| `observationalMemory.maxMessageTokenBudget` | int | `8000` | Max token budget for recent messages in context |

## Self-Learning System

Lango includes a self-learning knowledge system that improves agent performance over time.

- **Knowledge Store** - Persistent storage for facts, patterns, and external references
- **Learning Engine** - Observes tool execution results, extracts error patterns, boosts successful strategies
- **Skill System** - Agents can create reusable composite/script/template skills with safety validation
- **Context Retriever** - 8-layer context architecture that assembles relevant knowledge into prompts:
  1. Tool Registry — available tools and capabilities
  2. User Knowledge — rules, preferences, definitions, facts
  3. Skill Patterns — known working tool chains and workflows
  4. External Knowledge — docs, wiki, MCP integration
  5. Agent Learnings — error patterns, discovered fixes
  6. Runtime Context — session history, tool results, env state
  7. Observations — compressed conversation observations
  8. Reflections — condensed observation reflections

### Observational Memory

Observational Memory is an async subsystem that compresses long conversations into durable observations and reflections, keeping context relevant without exceeding token budgets.

- **Observer** — monitors conversation token count and produces compressed observations when the message token threshold is reached
- **Reflector** — condenses accumulated observations into higher-level reflections when the observation token threshold is reached
- **Async Buffer** — queues observation/reflection tasks for background processing
- **Token Counter** — tracks token usage to determine when compression should trigger

Configure in `lango.json`:

```json
{
  "knowledge": {
    "enabled": true,
    "maxLearnings": 100,
    "maxKnowledge": 500,
    "maxContextPerLayer": 10,
    "autoApproveSkills": false,
    "maxSkillsPerDay": 10
  },
  "observationalMemory": {
    "enabled": true,
    "messageTokenThreshold": 1000,
    "observationTokenThreshold": 2000,
    "maxMessageTokenBudget": 8000
  }
}
```

> **Note**: Observational Memory is currently configured via `lango.json` only. CLI/TUI commands for managing observations are planned for a future release.

## Security

Lango includes built-in security features for AI agents:

### Security Configuration

Lango supports two security modes:

1. **Local Mode** (Default)
   - Encrypts secrets using AES-256-GCM derived from a passphrase (PBKDF2).
   - **Interactive**: Prompts for passphrase on startup (Recommended).
   - **Headless**: Set `LANGO_PASSPHRASE` environment variable.
   - **Migration**: Rotate your passphrase using:
     ```bash
     lango security migrate-passphrase
     ```
   > **⚠️ Warning**: Losing your passphrase results in permanent loss of all encrypted secrets. Lango does not store your passphrase.

2. **RPC Mode** (Production)
   - Offloads cryptographic operations to a hardware-backed companion app or external signer.
   - Keys never leave the secure hardware.

Configure mode in `lango.json`:

```json
{
  "security": {
    "signer": {
      "provider": "local" // or "rpc"
    }
  }
}
```

### AI Privacy Interceptor

Lango includes a privacy interceptor that sits between the agent and AI providers:

- **PII Redaction** — automatically detects and redacts personally identifiable information before sending to AI providers
- **Approval Workflows** — optionally require human approval before executing sensitive tools
- **Custom PII Patterns** — extend detection with custom regex patterns via `security.interceptor.piiRegexPatterns`

### Secret Management

Agents can manage encrypted secrets as part of their tool workflows. Secrets are stored using AES-256-GCM encryption and referenced by name, preventing plaintext values from appearing in logs or conversation history.

### Output Scanning

The built-in secret scanner monitors agent output for accidental secret leakage. Registered secret values are automatically replaced with `[SECRET:name]` placeholders before being displayed or logged.

### Key Registry

Lango manages cryptographic keys via an Ent-backed key registry. Keys are used for secret encryption, signing, and companion app integration.

### Companion App Discovery (RPC Mode)

Lango supports optional companion apps for hardware-backed security. Companion discovery is handled within the `internal/security` module:

- **mDNS Discovery** — auto-discovers companion apps on the local network via `_lango-companion._tcp`
- **Manual Config** — set a fixed companion address

### Authentication

Lango supports OIDC authentication for the gateway:

```bash
# Login via OIDC provider
lango login google
```

Configure OIDC providers in `lango.json`:

```json
{
  "auth": {
    "providers": {
      "google": {
        "issuerUrl": "https://accounts.google.com",
        "clientId": "${GOOGLE_CLIENT_ID}",
        "clientSecret": "${GOOGLE_CLIENT_SECRET}",
        "redirectUrl": "http://localhost:18789/auth/callback/google",
        "scopes": ["openid", "email", "profile"]
      }
    }
  }
}
```

## Docker

```bash
# Build Docker image
make docker-build

# Run with docker-compose
docker-compose up -d
```

## Development

```bash
# Run tests with race detector
make test

# Run linter
make lint

# Build for all platforms
make build-all

# Run locally (build + serve)
make dev

# Generate Ent code
make generate

# Download and tidy dependencies
make deps
```

## License

MIT
