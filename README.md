# Lango 🚀

A high-performance AI agent built with Go, supporting multiple AI providers, channels (Telegram, Discord, Slack), and a self-learning knowledge system.

## Features

- 🔥 **Fast** - Single binary, <100ms startup, <100MB memory
- 🤖 **Multi-Provider AI** - OpenAI, Anthropic, Gemini, Ollama with unified interface
- 🔌 **Multi-Channel** - Telegram, Discord, Slack support
- 🛠️ **Rich Tools** - Shell execution, file system operations
- 🧠 **Self-Learning** - Knowledge store, learning engine, skill system
- 🔒 **Secure** - AES-256-GCM encryption, key registry, companion app support
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
│   ├── agent/              # Agent types, PII redactor
│   ├── app/                # Application bootstrap, wiring, tool registration
│   ├── channels/           # Telegram, Discord, Slack integrations
│   ├── cli/                # CLI commands
│   │   ├── auth/           #   lango login (OIDC/OAuth)
│   │   ├── doctor/         #   lango doctor (diagnostics)
│   │   ├── onboard/        #   lango onboard (TUI wizard)
│   │   └── security/       #   lango security migrate-passphrase
│   ├── companion/          # mDNS companion app discovery
│   ├── config/             # Config loading, env var substitution, validation
│   ├── ent/                # Ent ORM schemas and generated code
│   ├── gateway/            # WebSocket/HTTP server, OIDC auth
│   ├── knowledge/          # Knowledge store, context retriever
│   ├── learning/           # Learning engine, error pattern analyzer
│   ├── logging/            # Zap structured logger
│   ├── provider/           # AI provider interface and implementations
│   │   ├── anthropic/      #   Claude models
│   │   ├── gemini/         #   Google Gemini models
│   │   └── openai/         #   OpenAI-compatible (GPT, Ollama, etc.)
│   ├── security/           # Crypto providers, key registry, secrets store
│   ├── session/            # Ent-based SQLite session store
│   ├── skill/              # Skill registry, executor, builder
│   ├── supervisor/         # Provider proxy, privileged tool execution
│   └── tools/              # exec, filesystem
└── openspec/               # Specifications (OpenSpec workflow)
```

## AI Providers

Lango supports multiple AI providers with a unified interface. Provider aliases are resolved automatically (e.g., `gpt` -> `openai`, `claude` -> `anthropic`, `llama` -> `ollama`).

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
| `server.host` | string | `localhost` | Bind address |
| `server.port` | int | `18789` | Listen port |
| `agent.provider` | string | `gemini` | Primary AI provider ID |
| `agent.model` | string | `gemini-2.0-flash-exp` | Primary model ID |
| `agent.fallbackProvider` | string | - | Fallback provider ID |
| `agent.fallbackModel` | string | - | Fallback model ID |
| `agent.maxTokens` | int | `4096` | Max tokens |
| `agent.temperature` | float | `0` | Generation temperature |
| `providers.<id>.type` | string | - | Provider type (openai, anthropic, gemini) |
| `providers.<id>.apiKey` | string | - | Provider API key (supports `${ENV_VAR}`) |
| `providers.<id>.baseUrl` | string | - | Custom base URL (e.g. for Ollama) |
| `logging.level` | string | `info` | Log level |
| `logging.format` | string | `console` | `json` or `console` |
| `session.databasePath` | string | `~/.lango/sessions.db` | SQLite path |
| `security.signer.provider` | string | `local` | `local` or `rpc` |
| `security.passphrase` | string | - | **DEPRECATED** Use `LANGO_PASSPHRASE` env var |
| `knowledge.enabled` | bool | `false` | Enable self-learning knowledge system |
| `knowledge.maxLearnings` | int | - | Max learning entries per session |
| `knowledge.maxKnowledge` | int | - | Max knowledge entries per session |
| `knowledge.autoApproveSkills` | bool | `false` | Auto-approve new skills |
| `knowledge.maxSkillsPerDay` | int | - | Rate limit for skill creation |

## Self-Learning System

Lango includes a self-learning knowledge system that improves agent performance over time.

- **Knowledge Store** - Persistent storage for facts, patterns, and external references
- **Learning Engine** - Observes tool execution results, extracts error patterns, boosts successful strategies
- **Skill System** - Agents can create reusable composite/script/template skills with safety validation
- **Context Retriever** - 6-layer context architecture that assembles relevant knowledge into prompts

Configure in `lango.json`:

```json
{
  "knowledge": {
    "enabled": true,
    "maxLearnings": 100,
    "maxKnowledge": 500,
    "autoApproveSkills": false,
    "maxSkillsPerDay": 10
  }
}
```

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

### Key Registry

Lango manages cryptographic keys via an Ent-backed key registry. Keys are used for secret encryption, signing, and companion app integration.

### Companion App Discovery (RPC Mode)

Lango supports optional companion apps for hardware-backed security:

- **mDNS Discovery** - Auto-discovers companion apps on the local network via `_lango-companion._tcp`
- **Manual Config** - Set a fixed companion address

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
