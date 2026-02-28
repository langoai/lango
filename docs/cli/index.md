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
| `lango security keyring store` | Store passphrase in hardware keyring (Touch ID / TPM) |
| `lango security keyring clear` | Remove passphrase from keyring |
| `lango security keyring status` | Show hardware keyring status |
| `lango security db-migrate` | Encrypt database with SQLCipher |
| `lango security db-decrypt` | Decrypt database to plaintext |
| `lango security kms status` | Show KMS provider status |
| `lango security kms test` | Test KMS encrypt/decrypt roundtrip |
| `lango security kms keys` | List KMS keys in registry |

### Payment

| Command | Description |
|---------|-------------|
| `lango payment balance` | Show USDC wallet balance |
| `lango payment history` | Show payment transaction history |
| `lango payment limits` | Show spending limits and daily usage |
| `lango payment info` | Show wallet and payment system info |
| `lango payment send` | Send a USDC payment |

### P2P Network

| Command | Description |
|---------|-------------|
| `lango p2p status` | Show P2P node status |
| `lango p2p peers` | List connected peers |
| `lango p2p connect <multiaddr>` | Connect to a peer by multiaddr |
| `lango p2p disconnect <peer-id>` | Disconnect from a peer |
| `lango p2p firewall list` | List firewall ACL rules |
| `lango p2p firewall add` | Add a firewall ACL rule |
| `lango p2p firewall remove` | Remove firewall rules for a peer |
| `lango p2p discover` | Discover agents by capability |
| `lango p2p identity` | Show local DID and peer identity |
| `lango p2p reputation` | Query peer trust score |
| `lango p2p pricing` | Show tool pricing |
| `lango p2p session list` | List active peer sessions |
| `lango p2p session revoke` | Revoke a peer session |
| `lango p2p session revoke-all` | Revoke all active peer sessions |
| `lango p2p sandbox status` | Show sandbox runtime status |
| `lango p2p sandbox test` | Run sandbox smoke test |
| `lango p2p sandbox cleanup` | Remove orphaned sandbox containers |

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
| `lango bg list` | List background tasks |
| `lango bg status <id>` | Show background task status |
| `lango bg cancel <id>` | Cancel a running background task |
| `lango bg result <id>` | Show completed task result |

## Global Behavior

All commands read configuration from the active encrypted profile stored in `~/.lango/lango.db`. On first run, Lango prompts for a passphrase to initialize encryption.

Commands that need a running server (like `lango health`) connect to `localhost` on the configured port (default: `18789`).

!!! tip "Getting Started"
    If you're new to Lango, start with `lango onboard` to walk through the initial setup, then use `lango doctor` to verify everything is configured correctly.
