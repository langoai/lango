---
title: AI Providers
---

# AI Providers

Lango supports multiple AI providers through a unified interface. Switch between providers without changing application code.

## Supported Providers

| Provider | Config ID | Models | Notes |
|----------|-----------|--------|-------|
| **OpenAI** | `openai` | GPT-5.2, GPT-5.3 Codex | Also supports OpenAI-compatible APIs via `baseUrl` |
| **Anthropic** | `anthropic` | Claude Opus, Sonnet, Haiku | Full tool-use support |
| **Gemini** | `gemini` | Gemini Pro, Flash | Google Generative AI |
| **Ollama** | `ollama` | Any local model | Default endpoint: `http://localhost:11434/v1` |

## Provider Aliases

For convenience, Lango resolves common aliases to their canonical provider IDs:

| Alias | Resolves To |
|-------|-------------|
| `gpt`, `chatgpt` | `openai` |
| `claude` | `anthropic` |
| `llama` | `ollama` |
| `bard` | `gemini` |

You can use aliases anywhere a provider ID is accepted (CLI flags, config files, etc.).

## Configuration

### Interactive Setup

The easiest way to configure providers is through the onboarding wizard:

```bash
lango onboard
```

Or update provider settings at any time:

```bash
lango settings
```

### Config File

Providers are defined in the `providers` map in your config file (`~/.lango/config.yaml`):

```yaml
providers:
  my-openai:
    type: openai
    apiKey: ${OPENAI_API_KEY}
  my-anthropic:
    type: anthropic
    apiKey: ${ANTHROPIC_API_KEY}
  my-gemini:
    type: gemini
    apiKey: ${GEMINI_API_KEY}
  local-ollama:
    type: ollama
```

Then reference a provider in the agent config:

```yaml
agent:
  provider: openai
  model: gpt-5.2
```

!!! tip "Environment Variable Substitution"

    API keys support `${ENV_VAR}` syntax. Store sensitive keys in environment variables rather than plain text in the config file.

### OpenAI-Compatible APIs

Any OpenAI-compatible API can be used by setting a custom `baseUrl`:

```yaml
providers:
  my-custom-llm:
    type: openai
    apiKey: ${CUSTOM_API_KEY}
    baseUrl: https://api.custom-provider.com/v1
```

### Ollama (Local Models)

Ollama requires no API key. Just ensure the Ollama server is running:

```bash
# Start Ollama
ollama serve

# Pull a model
ollama pull llama3.2
```

```yaml
providers:
  local:
    type: ollama
```

The default endpoint is `http://localhost:11434/v1`.

## Fallback Configuration

Configure a fallback provider to handle failures gracefully:

```yaml
agent:
  provider: anthropic
  model: claude-sonnet-4-6
  fallbackProvider: openai
  fallbackModel: gpt-5.2
```

When the primary provider fails, Lango automatically retries with the fallback provider and model.

!!! tip "Recommended Setup"

    Use a reasoning model (e.g., Claude Opus, GPT-5.3 Codex) as your primary provider for complex tasks, and a faster model as the fallback for reliability.

## Provider Selection

The agent resolves providers in this order:

1. **Explicit provider** -- `agent.provider` + `agent.model` in config
2. **Alias resolution** -- common names mapped to canonical IDs
3. **Provider map lookup** -- matches against entries in `providers`

## Related

- [Configuration Basics](../getting-started/configuration.md) -- Full config file reference
- [Embedding & RAG](embedding-rag.md) -- Embedding providers share the same provider map
