# A2A Commands

Commands for inspecting A2A (Agent-to-Agent) protocol configuration and verifying remote agent connectivity. See the [A2A Protocol](../features/a2a-protocol.md) section for detailed documentation.

```
lango a2a <subcommand>
```

---

## lango a2a card

Show the local A2A agent card configuration, including enabled status, base URL, agent name, and configured remote agents.

```
lango a2a card [--json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango a2a card
A2A Agent Card
  Enabled:      true
  Base URL:     http://localhost:18789
  Agent Name:   lango
  Description:  AI assistant with tools

Remote Agents (2)
  NAME              AGENT CARD URL
  weather-agent     http://weather-svc:8080/.well-known/agent.json
  search-agent      http://search-svc:8080/.well-known/agent.json
```

When A2A is disabled:

```bash
$ lango a2a card
A2A Agent Card
  Enabled:      false

No remote agents configured.
```

---

## lango a2a check

Fetch and display a remote agent card from a URL. Useful for verifying that a remote A2A agent is reachable and correctly configured before adding it to your configuration.

```
lango a2a check <url> [--json]
```

| Argument | Required | Description |
|----------|----------|-------------|
| `url` | Yes | URL of the remote agent card (e.g., `http://host/.well-known/agent.json`) |

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--json` | bool | `false` | Output as JSON |

**Example:**

```bash
$ lango a2a check http://weather-svc:8080/.well-known/agent.json
Remote Agent Card
  Name:         weather-agent
  Description:  Provides weather data and forecasts
  URL:          http://weather-svc:8080
  DID:          did:lango:02abc...
  Capabilities: [weather, forecast]

Skills (2)
  ID              NAME              TAGS
  get-weather     Get Weather       [weather, location]
  forecast        5-Day Forecast    [weather, forecast]
```

!!! tip
    Use `lango a2a check` before adding a remote agent to your configuration to verify connectivity and inspect its capabilities.
