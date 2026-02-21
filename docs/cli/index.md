# CLI Reference

Lango provides a comprehensive command-line interface built with [Cobra](https://github.com/spf13/cobra). Every command supports `--help` for detailed usage information.

## Quick Reference

| Command | Description |
|---------|-------------|
| `lango serve` | Start the gateway server |
| `lango version` | Print version and build info |
| `lango health` | Check gateway health |
| `lango onboard` | Guided 5-step setup wizard |
| `lango settings` | Full interactive configuration editor |
| `lango doctor` | Diagnostics and health checks |

### Config Management

| Command | Description |
|---------|-------------|
| `lango config list` | List all configuration profiles |
| `lango config create <name>` | Create a new profile with defaults |
| `lango config use <name>` | Switch to a different profile |
| `lango config delete <name>` | Delete a configuration profile |
| `lango config import <file>` | Import and encrypt a JSON config |
| `lango config export <name>` | Export a profile as plaintext JSON |
| `lango config validate` | Validate the active profile |

### Agent & Memory

| Command | Description |
|---------|-------------|
| `lango agent status` | Show agent mode and configuration |
| `lango agent list` | List local and remote agents |
| `lango memory list` | List observational memory entries |
| `lango memory status` | Show memory system status |
| `lango memory clear` | Clear all memory entries for a session |
| `lango graph status` | Show graph store status |
| `lango graph query` | Query graph triples |
| `lango graph stats` | Show graph statistics |
| `lango graph clear` | Clear all graph data |

### Security

| Command | Description |
|---------|-------------|
| `lango security status` | Show security configuration status |
| `lango security migrate-passphrase` | Rotate encryption passphrase |
| `lango security secrets list` | List stored secrets (values hidden) |
| `lango security secrets set <name>` | Store an encrypted secret |
| `lango security secrets delete <name>` | Delete a stored secret |

### Payment

| Command | Description |
|---------|-------------|
| `lango payment balance` | Show USDC wallet balance |
| `lango payment history` | Show payment transaction history |
| `lango payment limits` | Show spending limits and daily usage |
| `lango payment info` | Show wallet and payment system info |
| `lango payment send` | Send a USDC payment |

### Automation

| Command | Description |
|---------|-------------|
| `lango cron add` | Add a new cron job |
| `lango cron list` | List all cron jobs |
| `lango cron delete <id-or-name>` | Delete a cron job |
| `lango cron pause <id-or-name>` | Pause a cron job |
| `lango cron resume <id-or-name>` | Resume a paused cron job |
| `lango cron history` | Show cron execution history |
| `lango workflow run <file>` | Execute a workflow YAML file |
| `lango workflow list` | List workflow runs |
| `lango workflow status <run-id>` | Show workflow run status |
| `lango workflow cancel <run-id>` | Cancel a running workflow |
| `lango workflow history` | Show workflow execution history |

## Global Behavior

All commands read configuration from the active encrypted profile stored in `~/.lango/lango.db`. On first run, Lango prompts for a passphrase to initialize encryption.

Commands that need a running server (like `lango health`) connect to `localhost` on the configured port (default: `18789`).

!!! tip "Getting Started"
    If you're new to Lango, start with `lango onboard` to walk through the initial setup, then use `lango doctor` to verify everything is configured correctly.
