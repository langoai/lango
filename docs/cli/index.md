# CLI Reference

Lango provides a comprehensive command-line interface built with [Cobra](https://github.com/spf13/cobra). Every command supports `--help` for detailed usage information.

## Quick Reference

| Command | Description |
|---------|-------------|
| `lango` | Launch multi-panel TUI cockpit (default entry point) |
| `lango cockpit` | Launch multi-panel TUI (same as bare `lango`) |
| `lango chat` | Launch plain chat TUI |
| `lango serve` | Start the gateway server |
| `lango version` | Print version and build info |
| `lango health` | Check gateway health |
| `lango status` | [Unified status dashboard](status.md) (health, config, features) |
| `lango status dead-letter-summary` | Show overview counts for the current dead-letter backlog |
| `lango status dead-letters` | List current dead-lettered post-adjudication executions |
| `lango status dead-letter <transaction-receipt-id>` | Show detailed dead-letter execution status for one transaction |
| `lango status dead-letter retry <transaction-receipt-id>` | Request retry for one dead-lettered post-adjudication execution |
| `lango onboard` | Guided 5-step setup wizard |
| `lango settings` | Full interactive configuration editor |
| `lango doctor` | Diagnostics and health checks |

### Agent Diagnostics

| Command | Description |
|---------|-------------|
| `lango agent trace list` | List recent turn traces with outcomes |
| `lango agent trace <id>` | Show detailed event timeline for a trace |
| `lango agent graph <session>` | Show delegation graph for a session |
| `lango agent trace metrics` | Per-agent trace-derived performance metrics |

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
| `lango agent tools` | Show tool-to-agent assignments |
| `lango agent hooks` | Show registered tool hooks |
| `lango memory list` | List observational memory entries |
| `lango memory status` | Show memory system status |
| `lango memory clear` | Clear all memory entries for a session |
| `lango memory agents` | List agents with persistent memory |
| `lango memory agent <name>` | Show memory entries for a specific agent |
| `lango graph status` | Show graph store status |
| `lango graph query` | Query graph triples |
| `lango graph stats` | Show graph statistics |
| `lango graph clear` | Clear all graph data |
| `lango graph add` | Add a triple to the knowledge graph |
| `lango graph export` | Export graph data to a file |
| `lango graph import` | Import graph data from a file |

### A2A Protocol

| Command | Description |
|---------|-------------|
| `lango a2a card` | Show local A2A agent card configuration |
| `lango a2a check <url>` | Fetch and display a remote agent card |

### Learning

| Command | Description |
|---------|-------------|
| `lango learning status` | Show learning system configuration |
| `lango learning history` | Show recent learning entries |

### Librarian

| Command | Description |
|---------|-------------|
| `lango librarian status` | Show librarian configuration and inquiry stats |
| `lango librarian inquiries` | List pending knowledge inquiries |

### Approval

| Command | Description |
|---------|-------------|
| `lango approval status` | Show approval system configuration |

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
| `lango security db-migrate` | Legacy SQLCipher migration command (unsupported in current runtime) |
| `lango security db-decrypt` | Legacy SQLCipher decrypt command (unsupported in current runtime) |
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
| `lango payment x402` | Show X402 auto-pay configuration |

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
| `lango p2p team list` | List active P2P teams |
| `lango p2p team status <id>` | Show team details and member status |
| `lango p2p team disband <id>` | Disband an active team |
| `lango p2p zkp status` | Show ZKP configuration |
| `lango p2p zkp circuits` | List compiled ZKP circuits |

### Economy

| Command | Description |
|---------|-------------|
| `lango economy budget status` | Show budget allocation status |
| `lango economy risk status` | Show risk assessment configuration |
| `lango economy pricing status` | Show dynamic pricing configuration |
| `lango economy negotiate status` | Show negotiation protocol status |
| `lango economy escrow status` | Show escrow service status |

### Smart Account

| Command | Description |
|---------|-------------|
| `lango account info` | Show smart account configuration and status |
| `lango account deploy` | Deploy a new Safe smart account with ERC-7579 adapter |
| `lango account session list` | List active session keys |
| `lango account session create` | Create a new session key |
| `lango account session revoke` | Revoke a session key or all session keys |
| `lango account module list` | List registered ERC-7579 modules |
| `lango account module install` | Install an ERC-7579 module |
| `lango account policy show` | Show current harness policy configuration |
| `lango account policy set` | Set harness policy limits |
| `lango account paymaster status` | Show paymaster configuration and approval status |
| `lango account paymaster approve` | Approve USDC spending for the paymaster |

### Contract

| Command | Description |
|---------|-------------|
| `lango contract read` | Call a view/pure smart contract method |
| `lango contract call` | Execute a state-changing contract method |
| `lango contract abi load` | Load and cache a contract ABI |

### Metrics

| Command | Description |
|---------|-------------|
| `lango metrics` | Show system metrics snapshot |
| `lango metrics sessions` | Show per-session token usage |
| `lango metrics tools` | Show per-tool metrics |
| `lango metrics agents` | Show per-agent metrics |
| `lango metrics history` | Show historical metrics |

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
| `lango workflow validate <file>` | Validate a workflow YAML file |
| `lango bg list` | List background tasks |
| `lango bg status <id>` | Show background task status |
| `lango bg cancel <id>` | Cancel a running background task |
| `lango bg result <id>` | Show completed task result |

### MCP Servers

| Command | Description |
|---------|-------------|
| `lango mcp list` | List all configured MCP servers |
| `lango mcp add <name>` | Add a new MCP server |
| `lango mcp remove <name>` | Remove an MCP server configuration |
| `lango mcp get <name>` | Show server details and discovered tools |
| `lango mcp test <name>` | Test server connectivity |
| `lango mcp enable <name>` | Enable an MCP server |
| `lango mcp disable <name>` | Disable an MCP server |

### RunLedger (Task OS)

!!! warning "Experimental"
    The RunLedger is experimental. See [RunLedger](../features/run-ledger.md).

| Command | Description |
|---------|-------------|
| `lango run list` | List recent runs |
| `lango run status` | Show RunLedger configuration |
| `lango run journal <run-id>` | View run journal events |

### Session Provenance

!!! warning "Experimental"
    Session provenance is experimental. See [Session Provenance](../features/provenance.md).

| Command | Description |
|---------|-------------|
| `lango provenance status` | Show provenance configuration and state |
| `lango provenance checkpoint list` | List checkpoints |
| `lango provenance checkpoint create` | Create a manual checkpoint |
| `lango provenance checkpoint show <id>` | Show checkpoint details |
| `lango provenance session tree` | Show session hierarchy tree |
| `lango provenance session list` | List persisted session nodes |
| `lango provenance attribution show <session>` | Show attribution data for a session |
| `lango provenance attribution report` | Generate attribution report |
| `lango provenance bundle export` | Export a signed provenance bundle |
| `lango provenance bundle import` | Import a signed provenance bundle |

### Sandbox (OS-level)

!!! warning "Experimental"
    The OS-level sandbox is experimental. This is distinct from `lango p2p sandbox` which manages P2P remote execution isolation.

| Command | Description |
|---------|-------------|
| `lango sandbox status` | Show sandbox configuration and platform capabilities |
| `lango sandbox test` | Run OS sandbox smoke tests |

## Global Behavior

All commands read configuration from the active encrypted profile stored in `~/.lango/lango.db`. On first run, Lango prompts for a passphrase to initialize encryption.

Commands that need a running server (like `lango health`) connect to `localhost` on the configured port (default: `18789`).

!!! tip "Getting Started"
    If you're new to Lango, start with `lango onboard` to walk through the initial setup, then use `lango doctor` to verify everything is configured correctly.
