# Core Commands

## lango serve

Start the gateway server. This boots the full application stack including all enabled channels, tools, embedding, graph, cron, and workflow engines.

```
lango serve
```

The server reads configuration from the active encrypted profile and starts:

- HTTP API on the configured port (default `18789`)
- WebSocket endpoint (if enabled)
- All configured channel adapters (Telegram, Discord, Slack)
- Background systems (cron scheduler, workflow engine) if enabled

Graceful shutdown is handled via `SIGINT` or `SIGTERM` with a 10-second timeout.

**Example:**

```bash
$ lango serve
INFO  starting lango  {"version": "0.5.0", "profile": "default"}
INFO  server listening  {"address": ":18789"}
```

---

## lango version

Print the binary version and build timestamp.

```
lango version
```

**Example:**

```bash
$ lango version
lango 0.5.0 (built 2026-02-20T12:00:00Z)
```

---

## lango health

Check whether the gateway server is running and healthy. Sends an HTTP GET to the `/health` endpoint.

```
lango health [--port N]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | int | `18789` | Gateway port to check |

**Examples:**

```bash
# Check default port
$ lango health
ok

# Check custom port
$ lango health --port 9090
ok
```

!!! info
    This command is designed to work as a Docker `HEALTHCHECK`. It exits with code 0 on success and non-zero on failure.

---

## lango onboard

Launch the guided 5-step setup wizard using an interactive TUI. This is the recommended way to configure Lango for the first time.

```
lango onboard [--profile <name>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `default` | Profile name to create or edit |

The wizard walks through five steps:

1. **Provider Setup** -- Choose an AI provider and enter API credentials
2. **Agent Config** -- Select model, max tokens, and temperature
3. **Channel Setup** -- Configure Telegram, Discord, or Slack
4. **Security & Auth** -- Enable privacy interceptor and PII protection
5. **Test Config** -- Validate your configuration

All settings are saved to an encrypted profile in `~/.lango/lango.db`.

**Example:**

```bash
# Run with default profile
$ lango onboard

# Create a separate "staging" profile
$ lango onboard --profile staging
```

!!! tip
    For full control over every configuration option, use `lango settings` instead.

---

## lango settings

Open the full interactive configuration editor. Provides access to all configuration options organized by category, including advanced features like embedding, graph, payment, and automation settings.

```
lango settings
```

The settings editor uses a TUI menu interface where you can navigate through categories and edit individual values. Changes are saved to the active encrypted profile.

!!! note
    This command requires an interactive terminal. For scripted configuration, use `lango config import` with a JSON file.

---

## lango doctor

Run diagnostics to check your Lango configuration and environment for common issues. Optionally attempt to fix problems automatically.

```
lango doctor [--fix] [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--fix` | bool | `false` | Attempt to automatically fix issues |
| `--json` | bool | `false` | Output results as JSON |

**Checks performed:**

- Encrypted configuration profile validity
- API key and provider configuration
- Channel token validation
- Session database accessibility
- Server port availability
- Security configuration (signer, interceptor, passphrase)
- Embedding provider connectivity
- Graph store status
- Multi-agent configuration
- A2A remote agent connectivity
- Output scanning and PII detection settings

**Examples:**

```bash
# Run diagnostics
$ lango doctor

# Attempt auto-fix for known issues
$ lango doctor --fix

# Machine-readable output
$ lango doctor --json
```

!!! tip
    Run `lango doctor` after `lango onboard` to verify your setup is correct. If issues are found, the `--fix` flag can resolve common problems automatically.
