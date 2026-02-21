---
title: Quick Start
---

# Quick Start

Get your agent running in three steps.

## Step 1: Install

```bash
git clone https://github.com/langowarny/lango.git
cd lango
make build
```

See [Installation](installation.md) for detailed instructions and prerequisites.

## Step 2: Onboard

Run the guided setup wizard:

```bash
./bin/lango onboard
```

The wizard walks you through five steps:

### 1. Provider Setup

Choose an AI provider and enter your API credentials.

| Provider | What You Need |
|---|---|
| OpenAI | API key |
| Anthropic | API key |
| Gemini | API key |
| Ollama | Local Ollama server URL |

### 2. Agent Config

Configure your agent's model settings:

- **Model** -- Select which model to use (e.g., `gpt-4o`, `claude-sonnet-4-20250514`)
- **Max Tokens** -- Maximum response length
- **Temperature** -- Creativity level (0.0 = deterministic, 1.0 = creative)

### 3. Channel Setup

Connect at least one messaging channel:

- **Telegram** -- Bot token from [@BotFather](https://t.me/BotFather)
- **Discord** -- Bot token from the Discord Developer Portal
- **Slack** -- Bot token and App token from the Slack API dashboard

### 4. Security & Auth

Configure security features:

- **Privacy Interceptor** -- Filter sensitive content from AI requests
- **PII Protection** -- Automatically detect and redact personally identifiable information

### 5. Test Config

The wizard validates your configuration by checking:

- Provider connectivity
- Channel authentication
- Configuration consistency

!!! tip

    You can re-run `lango onboard` at any time to update your configuration. It loads your existing settings as defaults.

## Step 3: Serve

Start the agent server:

```bash
./bin/lango serve
```

Your agent is now running and connected to your configured channels. Send a message through Telegram, Discord, or Slack to start a conversation.

!!! info "Health Check"

    Verify the server is running:

    ```bash
    lango health
    ```

    Or with curl:

    ```bash
    curl http://localhost:18789/health
    ```

## Beyond the Wizard

### Full Configuration Editor

For access to all configuration options beyond what the onboard wizard covers:

```bash
lango settings
```

This opens an interactive TUI editor with every available setting.

### Validate Configuration

Check your configuration for errors without starting the server:

```bash
lango config validate
```

## Next Steps

- [Configuration Basics](configuration.md) -- Learn about profiles and config management
- [CLI Reference](../cli/index.md) -- Full command documentation
- [Features](../features/index.md) -- Explore AI providers, channels, knowledge system, and more
