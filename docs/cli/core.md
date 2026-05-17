# Core Commands

## lango

Run `lango` without arguments in an interactive terminal to launch the standalone mission workbench. This is the default interactive entry point. It mounts Mission Control content directly without the full cockpit sidebar shell, while `lango chat` remains the direct focused-chat fallback and `lango cockpit` remains the explicit multi-page dashboard.
Startup notices such as the banner, log path, and initialization line are written through a seam-aware stderr path so wrappers and regressions can capture them without intercepting process-global stderr.

Interactive bare `lango` starts the mission workbench TUI. Non-interactive bare `lango` prints help to command stdout and exits successfully without starting the TUI. Unlike `lango cockpit` and `lango chat`, this bare-root fallback is not an actionable non-interactive error.

```bash
$ lango
```

TUI mode does not start the full live network and automation runtime by default. Gateway, live channels, cron, and P2P are not brought up as they are under `lango serve`, while other integrations such as configured MCP servers may still initialize through the local interactive bootstrap path. The `/status` slash command reflects that distinction: MCP is shown as active when the local interactive bootstrap initialized MCP, while configured-only MCP is shown separately from active runtime features.

If the active profile is still incomplete, the workbench empty state points you to `lango onboard`, `lango settings`, and `lango doctor` before you start chatting into a nonfunctional setup.
Once the profile is ready, the empty state switches to starter prompts and binds them to `1`, `2`, and `3`.
Those prompts are workdir-aware: they stay generic outside a repo, but become repository-aware inside a detected project, and they switch to Go-package guidance when a `go.mod` is present.
When Git metadata is available, the “next change” starter prompt also pivots around the current branch, uncommitted-change state, and the most obvious changed files or directories.
Pressing `Enter` on that empty ready-profile composer now loads the default, context-aware starter prompt automatically, and pressing `Enter` again submits it, while `1/2/3` still select the explicit starter choices. Once the starter is loaded, the body and footer copy switch to the submit step and keep `1/2/3` available for starter replacement. Once submitted, the same surface switches to a running-state hint that also explains that you can type the next prompt, use `1/2/3` to replace it, and press `Enter` to interrupt-and-run it. If you replace the staged follow-up with a starter prompt, that replacement becomes the next turn that will run. After the turn completes and the composer is empty again, the workbench treats that state as the next step rather than the original quick start: generic workspaces move to the structure-oriented starter, while detected workspaces stay on the next-change starter. Editing keys will pull focus back into the composer for that staged follow-up.
When the turn completes, the activity lane keeps both the user submission and a short, single-line assistant reply summary so the result remains visible from the workbench shell without flooding the timeline with raw multi-line output.
The composer placeholder follows the same split so the first line you can type into also reflects whether setup or immediate use is the right next action, including the `Press Enter` / `1-3` quick-start hint for ready profiles.
The header summary follows that same rule and shows `Model: Setup required` until provider and model configuration are actually usable.
The empty standalone shell also suppresses cockpit-style degraded warnings until there is real mission/control context to justify them, so first-run attention stays on setup and next actions.
After a turn finishes, the empty body also shifts from a blank no-missions message to an explicit next-step prompt.
The hint line likewise switches to "Type the next prompt here" so the next loop stays explicit.
The empty composer placeholder follows the same shift and switches to `Next step: press Enter ...` wording instead of the original first-run prompt.
The footer follows that same shift and now says `Type next prompt here` instead of generic chat wording.
The completed-turn empty body also shows the latest result as a compact preview so the next loop starts with immediate context.
If the latest result is a failed turn, the completed-turn lead now explicitly says the turn needs attention and asks you to pick the recovery step instead of implying a clean finish.
In that same failure state, the starter/body/footer wording also shifts from generic next-step language to recovery wording.
In that recovery state, pressing `Enter` now seeds a recovery-oriented starter instead of reusing the generic completed-turn default.
That includes the footer itself, which now says `Type recovery prompt here` instead of `Type next prompt here`.

---

## lango cockpit

Launch the multi-panel TUI dashboard explicitly. Mission Control still opens first inside this explicit cockpit shell, chat remains available as a cockpit detail page, and `lango chat` still bypasses the cockpit for focused chat. The cockpit provides:
Startup notices such as the banner, log path, and initialization line are written through a seam-aware stderr path so wrappers and regressions can capture them without intercepting process-global stderr.

- Mission Control as the default first surface, plus Chat, Settings, Tools, Status, Sessions, Tasks, Dead Letters, and Approvals in the sidebar
- Core cockpit pages remain routable even when some backing services are absent; those pages now surface explicit degraded/unavailable messaging instead of silently disappearing or pretending to be merely empty
- Settings keeps its own embedded inline help/footer and inline save feedback inside the page, while Status stays read-only and refreshes automatically while active
- Transcript viewport with assistant markdown reflow on resize
- Clear visual separation between user, assistant, status, and approval transcript blocks
- Dedicated turn status strip for ready/streaming/approval/failure states
- Tool lifecycle rows that show running, approval-wait, canceled, and completed states plus compact param previews for tool calls
- Approval transcript events that can include compact request-id annotations and compact request-summary previews for repeated requests
- Inline tool approval interrupts (`a` allow / `s` allow session / `d/Esc` deny)
- Slash commands (`/help`, `/clear`, `/model`, `/status`, `/mode`, `/cost`, `/exit`)
- Key bindings: `Enter` send, `Alt+Enter` newline, `Ctrl+C` cancel or double-press quit from idle/failed state, `Ctrl+D` quit

```bash
$ lango cockpit
```

---

## lango chat

Launch the plain chat TUI. A simpler, transcript-first experience without the multi-panel cockpit layout or Mission Control landing page. Suitable for quick interactions that don't require the full dashboard.
Startup notices such as the banner, log path, and initialization line are written through a seam-aware stderr path so wrappers and regressions can capture them without intercepting process-global stderr.
Focused chat uses the same setup readiness contract as the workbench: if the active profile is incomplete, normal turns show setup guidance before running and point to `lango onboard`, `lango settings`, and `lango doctor`; slash commands such as `/help` and `/status` remain available for inspection.

```bash
$ lango chat
```

---

## lango serve

Start the gateway server. This boots the full application stack including all enabled channels, tools, embedding, graph, cron, and workflow engines.
The startup banner and feature summary are written through the Cobra command output stream so wrappers and test harnesses can capture them directly.

```
lango serve
```

The server reads configuration from the active encrypted profile and starts:

- HTTP API on the configured port (default `18789`)
- WebSocket endpoint (if enabled)
- All configured channel adapters (Telegram, Discord, Slack)
- Background systems (cron scheduler, workflow engine) if enabled

Graceful shutdown is handled via `SIGINT` or `SIGTERM` with a 10-second timeout. If shutdown is already in progress, a second `Ctrl+C` forces immediate exit with code `130`.

**Example:**

```bash
$ lango serve
INFO  starting lango  {"version": "0.5.0", "profile": "default"}
INFO  server listening  {"address": ":18789"}
```

---

## lango version

Print the binary version and build timestamp.
The success output is written through the Cobra command output stream so wrappers and test harnesses can capture it directly.

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
The success output is written through the Cobra command output stream so wrappers and Docker-style callers can capture it directly.
Failure runs do not emit the `ok` payload or a duplicate Cobra-managed error body through the command output stream; callers receive the returned error instead.

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

Launch the guided 5-step setup wizard using an interactive TUI. This is the recommended way to configure Lango for the first time. Preset banners, cancel messages, and post-save guidance are written through the Cobra command output stream so wrappers and test harnesses can capture them directly.

```
lango onboard [--profile <name>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `default` | Profile name to create or edit |

The wizard walks through five steps:

1. **Provider Setup** -- Choose an AI provider (Anthropic, OpenAI, Gemini, Ollama, GitHub) and enter API credentials
2. **Agent Config** -- Select model (auto-fetched from provider), max tokens, and temperature
3. **Channel Setup** -- Configure Telegram, Discord, or Slack (or skip)
4. **Security & Auth** -- Privacy interceptor, PII redaction, approval policy
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

!!! note
    This command requires an interactive terminal. For scripted setup, use `lango config create --preset <name>` or `lango config import`.

---

## lango settings

Open the full interactive configuration editor. Provides access to all configuration options organized by category, including advanced features like embedding, graph, payment, and automation settings. Cancel and post-save guidance output are written through the Cobra command output stream so wrappers and test harnesses can capture them directly.

```
lango settings [--profile <name>]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--profile` | string | `default` | Profile name to edit |

The settings editor uses a TUI menu interface where you can navigate through categories and edit individual values. Categories are organized into sections:

- **Core:** Providers, Agent, Channels, Tools, Server, Session, Logging, Gatekeeper, Output Manager
- **AI & Knowledge:** Context Profile, Knowledge, Skill, Observational Memory, Embedding & RAG, Graph Store, Librarian, Retrieval, Auto-Adjust, Context Budget, Agent Memory, Multi-Agent, A2A Protocol, Hooks, Ontology
- **Automation:** Cron Scheduler, Background Tasks, Workflow Engine, RunLedger, Provenance
- **Payment & Account:** Payment, Smart Account, SA Session Keys, SA Paymaster, SA Modules
- **P2P & Economy:** P2P Network, P2P Workspace, P2P ZKP, P2P Pricing, P2P Owner Protection, P2P Sandbox, Economy, Risk, Negotiation, Escrow, On-Chain Escrow, Pricing
- **Integrations:** MCP Settings, MCP Server List, Observability, Alerting
- **Security:** Security, Auth, Legacy DB Encryption, Security KMS, OS Sandbox

Press `/` to search across all categories by keyword.

Changes are saved to the active encrypted profile.

!!! note
    This command requires an interactive terminal. For scripted configuration, use `lango config import` with a JSON file or `lango config set` for simple scalar values.

---

## lango doctor

Run diagnostics to check your Lango configuration and environment for common issues. Optionally attempt to fix problems automatically.
Both table and JSON modes write through the Cobra command output stream so wrappers and test harnesses can capture doctor output directly.
Doctor output also normalizes check names, messages, details, fix actions, and structured trace metadata to plain single-line text before rendering or serialization.
If bootstrap fails while loading the encrypted profile, `lango doctor` reports a dedicated failing `Bootstrap` diagnostic with the original error details, then continues running remaining checks in best-effort mode.

```
lango doctor [--fix] [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--fix` | bool | `false` | Attempt to automatically fix issues |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Checks performed include:**

- Configuration profile validity
- AI provider configuration and API keys
- API key security (env-var best practices)
- Channel token validation (Telegram, Discord, Slack)
- Session database accessibility
- Server port availability / network configuration
- Security configuration (signer, interceptor, encryption)
- Companion connectivity (WebSocket gateway status)
- Observational memory configuration
- Output scanning and interceptor settings
- Embedding / RAG provider and model setup
- Graph store configuration
- Multi-agent orchestration settings
- Recent multi-agent turn traces (`loop_detected`, `empty_after_tool_use`, `timeout`)
- Persisted isolated-turn leak detection
- A2A protocol connectivity
- RunLedger configuration invariants
- Tool hooks configuration
- Agent registry health
- Librarian status
- Retrieval coordinator and auto-adjust settings
- Context budget allocation ratio validation
- Approval system status
- Economy layer configuration
- Contract configuration
- Observability configuration

**Examples:**

```bash
# Run diagnostics
$ lango doctor

# Attempt auto-fix for known issues
$ lango doctor --fix

# Machine-readable output
$ lango doctor --output json
```

When multi-agent runtime failures exist, `--output json` now includes structured trace metadata such as `traceFailures[].traceId`, `outcome`, `errorCode`, `causeClass`, and `summary`, plus `isolationLeakCount` when applicable. The output remains valid pretty-printed JSON that wrappers can decode directly.

!!! tip
    Run `lango doctor` after `lango onboard` to verify your setup is correct. In multi-agent mode, `doctor` also reports recent failed turn traces and whether isolated specialist turns have leaked into persisted parent history.

---

## lango agent diagnostics

Agent diagnostics commands now live in the dedicated [Agent CLI Reference](agent.md).
Use that page for `lango agent trace list`, `lango agent trace show <trace-id>`, `lango agent trace metrics`, and `lango agent graph <session-key>`.

---

## lango config

Configuration profile management now lives in the dedicated [Config Management](config.md) reference.
Use that page for `lango config list/create/use/delete/import/export/get/set/keys/validate`.
