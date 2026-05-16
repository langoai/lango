---
title: Cockpit TUI
---

# Cockpit TUI

## Overview

The cockpit is the explicit multi-panel terminal dashboard launched via `lango cockpit`. Bare `lango` now launches the standalone mission workbench instead, while `lango chat` remains the focused single-panel chat fallback. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), the cockpit still opens on Mission Control first, keeps direct chat available inside the cockpit, and adds sidebar navigation, multiple detail pages, and a live context panel around that shared session.

On the standalone workbench surface, an incomplete profile now triggers setup guidance in the Mission Control empty state so the operator is pointed at `lango onboard`, `lango settings`, and `lango doctor` immediately.
For a ready profile, that same empty state now proposes a few starter prompts instead of generic dead-air copy and binds them to `1`, `2`, and `3`.
Those starter prompts are context-aware: they stay generic outside a detected repo, but become repository-aware inside a project and can pivot to Go-package guidance when the workdir contains a `go.mod`.
When Git metadata is available, the workbench also reshapes the change-review prompt around the current branch, dirty state, and the most obvious changed files or directories.
Pressing `Enter` on the empty ready-profile workbench now seeds the default, context-aware starter prompt automatically, and pressing `Enter` again submits it, while `1/2/3` still select the explicit starter choices. Once seeded, the empty-state body and footer move from the seed hint to the submit hint and keep `1/2/3` available for starter replacement. Once submitted, they switch again to a running-state hint that also explains that you can type the next prompt, use `1/2/3` to replace it, and press `Enter` to interrupt-and-run it. If you replace the staged follow-up with a starter prompt, that replacement becomes the next turn that will run. After the turn completes and the composer is empty again, the workbench treats that state as the next step rather than the original quick start: generic workspaces move to the structure-oriented starter, while detected workspaces stay on the next-change starter. Editing keys will pull focus back into the composer for that staged follow-up.
Once the turn completes, the shared activity lane keeps a short, single-line assistant reply summary as well as the user submission and turn summary, so the result remains visible from the workbench-first shell without dropping the full raw reply into the timeline.
The workbench composer placeholder mirrors that same state split, reinforcing setup recovery for incomplete profiles and exposing the `Press Enter` / `1-3` quick-start hint for ready ones.
The workbench header summary also reflects readiness now, showing `Model: Setup required` until the active profile has a usable provider/model path.
On that same empty standalone shell, cockpit-style degraded warnings are suppressed until there is real mission/control context to justify them, so the first screen stays focused on setup and next actions.
After a turn finishes, the empty body also shifts from a blank no-missions message to an explicit next-step prompt.
The hint line likewise switches to "Type the next prompt here" so the next loop stays explicit.
The empty composer placeholder follows the same shift and switches to `Next step: press Enter ...` wording instead of the original first-run prompt.
The footer follows that same shift and now says `Type next prompt here` instead of generic chat wording.
The completed-turn empty body also shows the latest result as a compact preview so the next loop starts with immediate context.
If the latest result is a failed turn, the completed-turn lead now explicitly says the turn needs attention and asks you to pick the recovery step instead of implying a clean finish.
In that same failure state, the starter/body/footer wording also shifts from generic next-step language to recovery wording.
In that recovery state, pressing `Enter` now seeds a recovery-oriented starter instead of reusing the generic completed-turn default.
That includes the footer itself, which now says `Type recovery prompt here` instead of `Type next prompt here`.

## Launch

```bash
lango            # launch standalone mission workbench
lango cockpit    # explicit cockpit launch
lango chat       # focused single-panel chat fallback (no sidebar, no pages)
```

The cockpit requires an interactive terminal with TTY support.

## Layout

```
┌──────────┬─────────────────────────┬──────────────┐
│          │                         │              │
│ Sidebar  │     Main Content        │   Context    │
│ (pages)  │     (active page)       │   Panel      │
│          │                         │  (metrics)   │
│          │                         │              │
└──────────┴─────────────────────────┴──────────────┘
```

- **Sidebar** -- page navigation list, toggled with `Ctrl+B`
- **Main content** -- active page rendering (Mission Control, chat, settings, tools, status, sessions, tasks, dead letters, or approvals)
- **Context panel** -- live system metrics, toggled with `Ctrl+P`

## Pages

| Page | Description | Notes |
|------|-------------|-------|
| **Mission Control** | Default first screen with durable missions first, one live pending decision, a compact agenda/loops layer, recent activity, unmatched runtime overlays, and an inline composer | Chat remains available by typing here or via `lango chat` |
| **Chat** | Focused AI conversation page inside cockpit | `Ctrl+1`; `lango chat` bypasses cockpit entirely |
| **Settings** | Interactive configuration viewer | Always available; persists through the profile store when configured and otherwise degrades to a read-write editor with explicit save-unavailable messaging |
| **Tools** | Read-only tool catalog browser with category navigation and per-category tool details | Always available; degrades to explicit unavailable messaging when the tool catalog is absent |
| **Status** | System status dashboard (health, features, agent state) | Always available; degrades gracefully when feature-status or metrics providers are absent |
| **Sessions** | Session history and management | Always available; degrades to explicit unavailable messaging when the session list source is absent |
| **Tasks** | Background task status and management | Always available; degrades to explicit unavailable messaging when the background task manager is absent |
| **Dead Letters** | Dead-letter backlog and retry surface | Always available; degrades to explicit unavailable messaging until the dead-letter bridge is ready |
| **Approvals** | Approval history and grant management | Always available; degrades to explicit unavailable messaging when approval stores are absent |

The current sidebar order is Mission Control, Chat, Settings, Tools, Status, Sessions, Tasks, Dead Letters, and Approvals. Mission Control is the default active page on startup. Core cockpit pages stay routable even when some backing services are unavailable; degraded pages explain the missing dependency in-page instead of disappearing from the navigation surface.
Mission Control follows the same rule: if its backing projector is unavailable, the page stays open and surfaces an explicit degraded note instead of silently pretending the dashboard is simply empty.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `Ctrl+B` | Toggle sidebar visibility |
| `Ctrl+P` | Toggle context panel |
| `Tab` | Switch focus between sidebar and main content |
| `Ctrl+Y` | Copy to clipboard |
| `Ctrl+1` | Switch to Chat page |
| `Ctrl+2` | Switch to Settings page |
| `Ctrl+3` | Switch to Tools page |
| `Ctrl+4` | Switch to Status page |
| `Ctrl+5` | Switch to Tasks page |
| `Ctrl+6` | Switch to Approvals page |

Mission Control, Sessions, and Dead Letters are reached from the sidebar. Wave 1 intentionally keeps the existing `Ctrl+1` through `Ctrl+6` mappings for the detail pages instead of remapping them around Mission Control.

## Settings Page

The Settings page embeds the interactive configuration editor directly inside cockpit. It stays routable even when the profile store is absent, so the operator can still inspect and edit configuration structure without leaving the dashboard.

When the config profile store is configured, saves persist through the embedded editor's normal save path. When the store is absent, the page shows an explicit warning that settings changes cannot be saved and the save callback fails closed with a `config profile store is not configured` error instead of pretending persistence succeeded.
Embedded save attempts also surface inline feedback at the top of the Settings menu: a success banner after a save completes, or a failure banner with the save error when persistence fails.

Unlike the read-only cockpit pages, Settings does not use the cockpit help bar for discoverability. The embedded editor renders its own inline help footer that changes by mode:

- menu screens advertise `↑/↓`, `Enter`, `/`, `Tab`, and `Esc`
- forms advertise field navigation and editing keys such as `Tab`, `Shift+Tab`, `Space`, `←/→`, `Enter`, and `Esc`
- provider/auth/MCP list screens advertise `↑/↓`, `Enter`, `d`, and `Esc`

## Status Page

The Status page groups five read-only sections:

- **Feature Status** — feature flags and reasons for disabled entries
- **Token Usage** — input, output, and total token counters
- **Tool Execution** — executed tool counts and last-tool summary
- **Graph Admission** — graph-related admission or extraction counters
- **System** — high-level runtime/environment summary

The page stays available even when its backing providers are missing. If no feature-status provider is configured, the Feature Status section shows explicit unavailable messaging. If no metrics collector is configured, the Token Usage, Tool Execution, and Graph Admission sections show explicit unavailable messaging instead of looking like zero-valued runtime data. Runtime-fed labels such as feature names, disable reasons, tool names, graph-admission identifiers, and config-fed provider/model labels are normalized to plain single-line text before rendering.
Status is read-only: it does not expose cockpit help-bar bindings, and the page refreshes automatically while it stays active.

## Tools Page

The Tools page is a read-only catalog browser for the currently registered tool surface. It renders category rows on the left and the selected category's tool table on the right.

### Tools Page Keys

| Key | Action |
|-----|--------|
| `↑`/`k` | Move category selection up |
| `↓`/`j` | Move category selection down |

Changing the category cursor updates the right-side detail pane immediately; there is no separate Enter-to-select step. Each detail table row shows the tool name, summary description, and safety level for the selected category. Category labels, descriptions, tool labels, and safety text are normalized to plain single-line text before rendering, and known safety colors are selected from that sanitized label.

Navigation hints appear only when there is another category to move to. When the tool catalog is unavailable, both panes stay routable and surface the same explicit unavailable message: `Tool catalog is not available.` If the catalog exists but no categories are registered, both panes explain `No categories registered.` instead of leaving the category browser looking blank.

## Mission Control

Mission Control is the default cockpit landing surface. Inside the explicit cockpit shell it remains a sidebar-first page rather than a separate top-level CLI command.

- **Missions lane** — durable mission rows render first for the current cockpit session. Linked RunLedger and AgentRun data enrich those rows when available.
- **Collaboration context** — durable mission rows can now show compact local coworking context such as participants, active owner, recent handoff, blocked state, budget pressure, recovery hint, and linked review state when attribution is provable.
- **Runtime overlays** — background/runtime work that is still unlinked to a durable mission remains visible as overlay instead of disappearing. Foreign-session background work is filtered out when it carries a different `OriginSession`.
- **Proactive proposals** — transient, session-scoped proposal records now render as first-class proposed rows. In the current slice, only learning suggestions actively create proposals.
- **Learning-buffer fallback** — when the older learning-buffer compatibility path is used, buffered pattern/rule/rationale text is normalized to plain single-line text before replay as well.
- **Decisions lane** — the latest pending approval is shown as one live decision with action, reason, effect, and risk, while durable mission rows can separately show coarse `waiting_decision` state.
- **Agenda lane** — a compact loop surface renders operator loops in deterministic priority order without replacing the mission board.
- **Activity lane** — recent deterministic activity is listed as compact timeline entries.
- **Composer** — the first-screen hint is: "Type a request here, or use `lango chat` for focused chat."
- **Help bar** — in the true empty state, Mission Control drops inert `↑/↓` navigation hints and keeps only the actionable focus/submit keys.
- **Header summaries** — active-agent, model/provider, context, metrics, and degraded-note summaries are normalized to plain single-line text before rendering.
- Header snapshot storage also follows that rule before replay, so downstream consumers do not retain raw control-sequence payloads in `HeaderView`.
- Active-agent header aggregation follows the same rule even before snapshot assembly, so malformed owner-agent labels do not survive the summary builder.
- **Lane text** — mission titles/details, collaboration labels, proposal source labels, decision text, activity summaries, loop summaries, and overflow summaries are also normalized to plain single-line text before rendering.
Activity summaries are compacted into plain single-line text at buffer time as well, so replay paths do not retain raw control-sequence payloads.
The exported assistant-activity helper also returns that normalized single-line form directly, so workbench reuse does not rely on a later buffer pass for safety.
Projected mission/proposal/decision/collaboration text is also normalized before it enters the Mission Control snapshot model, so downstream replay paths inherit the same plain-text baseline.
Projected loop rows now follow that same rule before they enter the Mission Control snapshot model.

### Mission Control Keys

| Key | Action |
|-----|--------|
| `Tab` | Cycle focus between missions, decisions, and composer |
| `Enter` | When actionable, submit from the composer, accept a focused proposed mission, seed the default starter prompt in an empty ready workbench, or run that starter once it is staged |
| `↑`/`k` | Move mission or activity selection when the currently focused lane has another row |
| `↓`/`j` | Move mission or activity selection when the currently focused lane has another row |
| `d` | Dismiss the selected proposed mission when the missions lane is focused and proposal dismissal is available |

In the true empty state, Mission Control reduces that help surface to the keys that still do real work. On the ordinary cockpit surface that means focus cycling only, while the standalone workbench can additionally expose `Enter` for seeding or running the starter prompt. Outside the empty state, the `↑/↓` navigation hints still appear only when the currently focused lane has another row to move through, and composer/activity focus uses those keys to move the activity cursor rather than swallowing them into the composer. When a proposed mission is focused and another mission row still exists, those navigation hints remain visible alongside `Enter` accept and `d` dismiss. The footer hint now follows that same proposal-focused state instead of falling back to generic request-entry copy. When focus is on the decisions lane, the help bar drops the generic `Enter` submit hint because decision handling does not use `Enter`.
When the workbench empty state replays the latest assistant outcome as `Last result:`, that summary is normalized to plain single-line text before rendering.

Current landed behavior:

- submitting a top-level request from the Mission Control composer creates a durable mission row before the shared turn is dispatched
- those durable mission labels are normalized to plain single-line text before they are persisted
- when the mission service is unavailable, ordinary Mission Control composer submits fail closed with an explicit system message instead of silently falling back to generic chat execution
- learning suggestions are the only active proactive producer in this slice; librarian-gap and runtime-failure producers remain disabled
- proposal preparation is deterministic and source-native: the prepared brief is derived from stable learning-suggestion fields and does not launch generic proposal-owned background work or `RunLedger` work
- accepting a prepared proposal creates the first durable mission row and preserves the prepared brief context on the mission path before the transient proposal leaves the active set
- that accepted proposal description is normalized to plain single-line text before it is persisted into the durable mission row
- prepared brief segments are collapsed into one single-line durable description instead of being persisted as multiline note blocks
- if the services behind proposed-mission accept or dismiss actions are unavailable, Mission Control returns an explicit system message instead of silently doing nothing
- `waiting_decision` is a coarse durable mission state, not a durable approval queue
- Mission Control can stay visible while a pending approval is active, and resolving that decision uses the same shared approval path as the chat surface
- The shared pending approval owner also normalizes tool name, summary, rule explanation, and risk labels before replay so Mission Control and cockpit chat do not retain raw control-sequence payloads in their shared approval state.
- loops are projected only from real existing sources already wired into the app: durable missions, pending inquiries, dead-letter backlog, cron jobs, and deterministic follow-up signals
- agenda ordering is deterministic: waiting-user, blocked, active, scheduled, needs-review, then resolved; ties break by newer updates first
- collaboration is mission-linked local coworking only in this slice; it remains attached to durable mission rows/details instead of becoming a separate dashboard or durable model
- participants, handoffs, blocked state, budget pressure, and recovery hints appear only when attribution is provable from linked local execution data
- reviewing appears only when linked local `RunLedger` execution state is review-needed, currently represented by `verify_pending`

Current limits:

- proposal state is transient and session-scoped; no durable `proposed` mission row exists before acceptance
- the proposal registry is the primary truth for rendered proposals; the older learning-buffer path remains only as compatibility fallback when the registry is unavailable
- the activity lane remains deterministic runtime rendering, not LLM-generated humanized narration
- task tracking stays lightweight and separate from mission truth; the Tasks page and `TaskEntry` tooling are not the authoritative durable mission checklist model
- RunLedger and AgentRun remain enrichments for durable or unmatched runtime rows. When those readers or mission details are unavailable, Mission Control shows a degraded note instead of inventing values
- collaboration context is not shown from generic session noise alone; unattributable delegation, budget, or recovery events stay out of mission rows
- scheduled automation loops mean cron jobs only in this slice; workflow runs are not projected as scheduled loops yet
- cron-job names and last-run status labels are normalized to plain single-line text before they are projected into those scheduled loops
- dead-letter loops and cron loops are operator-global in the current slice, while mission loops, inquiry loops, and follow-up loops are scoped to the current cockpit session
- no calendar, inbox, or external task-system loop integrations exist yet; the loop slice only renders sources that already exist in code
- external P2P team UX remains secondary and is not part of the main Mission Control collaboration surface in this slice

## Context Panel

The context panel displays live system metrics in a right-side panel. It refreshes every 5 seconds via tick messages. Displayed tool names, runtime active-agent labels, and channel names are normalized to plain single-line text before rendering, and the cached/shared snapshot models for those labels are sanitized both before replay and at the context-panel setter boundary.

### Token Usage

Displays cumulative token counts for the session:

| Row | Description |
|-----|-------------|
| Input | Input tokens consumed |
| Output | Output tokens generated |
| Total | Combined total |
| Cache | Tokens served from cache |

### Tool Stats

Shows the top 5 tools by invocation count, sorted by frequency. Displays "No tool executions" when no tools have been called.

### Runtime Status (conditional)

Shown only while a turn is actively running (`IsRunning = true`). Hidden when idle.

- **Active agent indicator** -- green dot (●) with the currently active agent name
- **Delegation count** -- number of agent-to-agent delegations in the current turn (shown only when > 0)
- **Turn token count** -- token usage for the current turn (shown only when > 0)

### Channels (conditional)

Shown only when channel connections exist. Each channel displays:

- **Connection status** -- green dot (●) for connected, hollow circle (○) for disconnected
- **Channel name**
- **Message count** -- total messages received

### System

- **Uptime** -- process uptime since launch

When the observability collector is unavailable, the context panel's Token Usage, Tool Stats, and System sections each render explicit unavailable messaging instead of silently looking like zero-valued data.

## Cockpit vs Chat Mode

| Feature | Cockpit (`lango`) | Chat (`lango chat`) |
|---------|-------------------|---------------------|
| Sidebar navigation | Yes | No |
| Multiple pages | Yes (Mission Control + 8 detail/sidebar pages) | No (chat only) |
| Context panel | Yes | No |
| Keyboard shortcuts | Full set | Chat-only |
| Terminal width | Recommended 120+ cols | Any width |

## Chat Page

The Chat page is the focused transcript-first conversation surface inside the cockpit shell. It shares the same approval owner and runtime event stream as Mission Control, but keeps the classic chat interaction model visible on one page.

The shell header and turn-status strip stay single-line and clamp on narrow terminals so the chat layout does not gain surprise extra rows when space gets tight.
Assistant markdown content is sanitized for terminal control sequences before rendering, and the plain-text fallback path uses that same sanitized content if markdown rendering fails.
User transcript blocks likewise strip terminal control sequences from displayed prompt text, so pasted terminal escapes do not leak back into the `You` lane.

### Chat Page Keys

| Key | Action |
|-----|--------|
| `Enter` | Send the current composer contents |
| `Alt+Enter` | Insert a newline without sending |
| `Ctrl+C` | Cancel generation, or quit from idle/failed state on double press |
| `Ctrl+D` | Quit immediately |
| `PgUp` / `PgDn` | Scroll transcript |

The page also exposes built-in slash commands such as `/help`, `/clear`, `/model`, `/status`, `/mode`, `/cost`, and `/exit`. Inline approval interrupts use the shared approval owner and stay actionable from the Chat page with `a` (allow), `s` (allow session), and `d` or `Esc` (deny), while `Ctrl+D` remains the immediate quit path even during approval. The approval-state turn strip now follows that same contract instead of shortening the deny path to `d` alone. Compact status/approval rows strip control sequences before collapsing into single-line text, and approval request params render in stable key order across the banner and fullscreen dialog with both parameter keys and values sanitized to plain single-line text. Nested structured payloads still render deterministically instead of following raw map iteration order.

## Tool Lifecycle Visibility

During streaming, each tool invocation appears as a distinct transcript item with lifecycle state:

- **Running** (⚙) — tool is executing, with a compact param preview when request params exist
- **Success** (✓) — tool completed with duration
- **Error** (✗) — tool failed with error preview
- **Canceled** (⊘) — tool was canceled
- **Awaiting Approval** (🔒) — tool requires user approval

When a tool row started with request params, that compact param preview now persists through approval waits, cancellation, success, and error so the execution context remains visible beside each lifecycle transition. Tool names and preview/output detail lines strip control sequences, render as plain single-line text, and each visible detail line clamps on narrow terminals instead of letting indented rows spill past the transcript width. Approval transcript events also carry compact request-id annotations and compact request-summary previews when available, sanitize that traceability text into plain single-line content before compaction, and therefore remain readable even when malformed metadata appears. The live in-flight streaming row applies that same control-sequence stripping before showing partial assistant output.

## Thinking Indicators

When the model uses extended thinking (via `genai.Part.Thought`), thinking phases appear as collapsible transcript items showing duration. Active thinking previews stay single-line, strip control sequences, and truncate on narrow terminals instead of spilling past the transcript width. The pending indicator (`⏳ Working...`) that covers the submit-to-first-event gap now follows that same narrow-terminal width clamp instead of overflowing.

## Two-Tier Approval

Approval requests are classified into two tiers based on tool safety level and capability:

- **Tier 1 (Inline Strip)** — compact single-line prompt for safe/moderate tools (e.g., browser_search, browser_observe)
- **Tier 2 (Fullscreen Dialog)** — overlay with risk badge, parameters, diff preview, a `t` split-mode toggle, and overflow-aware scroll controls for dangerous filesystem/exec tools (e.g., exec, fs_write, fs_edit)

Both tiers support the same actions: `a` (allow), `s` (allow session), `d`/`Esc` (deny).

The Tier 1 inline strip, fallback approval banner, and fullscreen approval dialog all sanitize displayed tool names and summaries into plain single-line text before truncation or layout, and any displayed channel-origin text is sanitized the same way. Channel-origin and badge extraction also sanitize the session-key prefix before deciding whether the request came from Telegram, Discord, or Slack. The fullscreen dialog applies that same plain-text sanitization to visible risk metadata such as the risk badge, risk label, and `Why:` explanation, uses the sanitized risk level for badge color selection, and strips control sequences from diff preview lines before styling them, so malformed metadata cannot break the approval surfaces.

### Double-Press Guardrail

Critical-risk tools (dangerous + filesystem or dangerous + automation) require pressing `a` or `s` twice within 3 seconds to confirm:

1. First press shows a warning telling you to press the same pending action key again (for example `a` or `s`) to confirm the destructive operation
2. Second press of that same key within 3 seconds executes the action
3. Pressing a different key or waiting longer than 3 seconds resets the pending state
4. Pressing `d` or `Esc` still denies the request immediately while the confirm prompt is visible

In the inline strip, critical-risk tools are labeled with **(destructive)** in red next to the tool name.

### Rule Explanation

The fullscreen approval dialog includes a "Why: ..." explanation between the summary and parameters sections. This explanation is derived from the tool's SafetyLevel and Category combination:

| SafetyLevel + Category | Explanation |
|------------------------|-------------|
| dangerous + filesystem | "This tool modifies the filesystem and is classified as dangerous." |
| dangerous + automation | "This tool executes arbitrary code and is classified as dangerous." |
| moderate (any) | "This tool creates or modifies resources (moderate risk)." |
| other | "This tool requires approval under the current approval policy." |

## Runtime Visibility

During active turns, runtime events appear as inline items in the chat transcript:

### Delegation Events

Agent-to-agent delegation events display in the transcript as:

```
 🔀 from → to  reason
```

Where `from` and `to` are the agent names (highlighted, sanitized to plain single-line text), and the reason is shown in italics when provided.

The delegation row stays a compact one-line event and truncates long actor or reason text on narrow terminals instead of spilling past the transcript width.

### Budget Warnings

When delegation budget usage is reported, a warning appears:

```
⚠ Delegation budget: used/max (percentage%)
```

### Recovery Events

Recovery decisions during structured orchestration appear as:

```
 🔄 Action #attempt  (causeClass) backoff
```

The recovery row stays a compact one-line event, strips control sequences from recovery metadata before both action-label mapping and visible text rendering, normalizes that metadata to a single line, and truncates on narrow terminals instead of spilling past the transcript width.

The action label maps as follows:
- `retry` → "Retry"
- `retry_with_hint` → "Reroute"
- `direct_answer` → "Direct Answer"
- `escalate` → "Escalate"

### Turn Token Summary

After each assistant response completes, a token usage summary is appended to the transcript:

```
📊 Token usage: Xinput, Youtput, Ztotal (Wcached)
```

The cached portion is omitted when cache tokens are zero. Large numbers are formatted with k/M suffixes (e.g., "1.5k").

## Approvals Page

The Approvals page (`Ctrl+6`) provides a dedicated view for approval history and active session grants. Data refreshes automatically every 2 seconds.

### History Section

Displays a table of past approval decisions with the following columns:

| Column | Description |
|--------|-------------|
| Time | Relative timestamp (e.g., "2m ago") |
| Tool | Tool name that was evaluated |
| Summary | Brief description of the request |
| Outcome | Decision result (e.g., "granted", "denied", "bypass", "timeout") |
| Provider | Which approval provider handled the request |

### Grants Section

Shows currently active session-level grants:

| Column | Description |
|--------|-------------|
| Session | Session key identifier |
| Tool | Granted tool name |
| Granted | Relative time when the grant was created |

### Approvals Page Keys

| Key | Action |
|-----|--------|
| `Tab /` | Toggle between history and grants sections |
| `↑`/`k` | Move cursor up when the active section has rows |
| `↓`/`j` | Move cursor down when the active section has rows |
| `r` | Revoke the selected grant (grants section with rows only) |
| `R` | Revoke the selected session's grants (grants section with rows only) |

When both approval stores are absent, the page displays an unavailable message instead of pretending the system merely has no entries yet. If only one store is absent, the corresponding section shows an unavailable message while the other section stays live. When the history store is configured but empty, the history section uses the same `No approval history yet.` wording that the fully empty page uses. History and grant metadata such as session keys, tool names, summaries, outcomes, and providers are normalized to plain single-line text before rendering. Both the help bar and the footer hint strip use the shared `tab /` section-toggle label, navigation hints only appear when the currently active section has another row to move to, and revoke actions only appear when the grants section actually contains grant rows.

## Sessions Page

The Sessions page shows lightweight session summaries using session key plus relative last-update time. The page sorts loaded sessions by most recent update first and supports simple cursor navigation with `↑`/`k` and `↓`/`j` only when there is another session row to move to. Displayed session keys and configured-source load errors are normalized to plain single-line text before rendering.

When the session list source is absent, the page shows an unavailable message instead of pretending the session history is simply empty. When the source is configured but returns no sessions, the page displays "No sessions found."
When the source is configured but the load fails, the page shows an explicit session-list failure message together with the underlying error text.

## Background Tasks Page

The Tasks page (`Ctrl+5`) shows background tasks in a table view with columns for ID, Prompt, Status, and Elapsed time (elapsed is hidden on narrow terminals below 50 columns).
When the background task manager is absent, the page shows an unavailable message instead of pretending there are simply no tasks.
List-mode `Enter` appears only when a task row exists.
List-mode `↑/↓` navigation hints appear only when another task row exists.
If the list refreshes to empty while the detail panel was open, the page exits detail mode and drops the detail-only help surface instead of leaving stale close-detail actions behind.
Displayed task IDs, prompts, statuses, origin labels, results, errors, and transient task-action messages are normalized to plain single-line text before rendering, and the detail panel wraps that sanitized text instead of replaying raw control sequences or embedded newlines.

### Task Detail View

- **Enter** -- toggle the detail panel for the selected task
- **Esc** -- close the detail panel
- **↑/↓** -- scroll within the detail content when the detail panel has overflow

The detail panel shows:
- Status with elapsed time
- Origin channel (e.g., "telegram", "slack")
- Token usage
- Full prompt text (word-wrapped)
- Result text
- Error message (if any)

### Task Actions

When a `TaskActioner` is available:

| Key | Action | Applies to |
|-----|--------|------------|
| `c` | Cancel task | Running or pending tasks |
| `r` | Retry task | Failed or cancelled tasks |

The help bar only shows the action keys that are valid for the currently selected task state.

Action results appear as a transient status message that auto-clears after 3 seconds. When cancel or retry fails, the message uses explicit task-action failure wording together with the underlying error text.

## Background Task Strip

When a BackgroundManager is available, a compact task strip appears above the footer showing active task count and the most recent task's status. The full Tasks page (Ctrl+5) provides a detailed table view.

## Approval Operations

Approval handling uses the shared cockpit approval owner. When a tool call needs approval, the live request is always forwarded into the shared chat model. If the operator is already on Mission Control, that page can stay visible and render the live decision there; other pages still switch to Chat so the approval remains actionable. Operators respond with `a` to allow, `s` to allow for the session, or `d`/`Esc` to deny.

Critical-risk filesystem or automation tools still require the existing double-press confirmation before `a` or `s` takes effect. For the full policy model and approval-provider behavior, see [Tool Approval](../security/tool-approval.md) and [Approval CLI](../security/approval-cli.md).

## Channel Operations

With `--with-channels`, the cockpit acts as a live operator console for Telegram, Discord, and Slack. Channel messages flow into the Chat transcript through the EventBus, and approval requests from those sessions surface in the cockpit even when the operator is on another page.

Channel transcript rows stay compact one-line events: displayed channel names, remote sender names, and message text have terminal control sequences stripped, are normalized to a single line, and narrow terminals truncate the final row instead of letting channel noise spill past the transcript width. Badge color selection follows that same sanitized channel name so known channels do not silently downgrade to the unknown-color path under malformed input.

Channel approvals apply to the originating channel session, so session grants and denials stay scoped to that remote conversation. Do not run `lango cockpit --with-channels` and `lango serve` against the same channel credentials at the same time. For setup details, see [Channels](channels.md).

## Background Task Operations

The Tasks page and the chat footer strip expose background task progress. Operators can inspect a task's detail panel, cancel running or pending work, and retry failed or cancelled tasks from within the cockpit. The page refreshes automatically, so the current task state stays visible without manual reloads.

For the system-level task model, CLI commands, and configuration reference, see [Background Tasks](../automation/background.md).

## Dead Letters Page

The Dead Letters page provides a post-adjudication backlog and retry surface for dead-lettered transaction flows. It combines a filter bar, backlog table, selected-transaction detail pane, and retry request flow in one operator page.

### Dead Letters Filter Controls

The filter strip supports direct text entry plus compact family/subtype selectors. The currently focused field is highlighted in-page, and pressing `Enter` applies the draft filters to the backlog query.

| Key | Action |
|-----|--------|
| `↑`/`k` | Move backlog selection up when rows exist |
| `↓`/`j` | Move backlog selection down when rows exist |
| `Tab` | Move to the next filter field |
| `←`/`→` | Cycle adjudication |
| `[` / `]` | Cycle latest subtype |
| `,` / `.` | Cycle latest family |
| `;` / `/` | Cycle any-match family |
| `Backspace` | Edit the active text filter |
| `Enter` | Apply the current draft filters |
| `Ctrl+R` | Reset all filters |

When filters are active but no rows match, the page displays `No dead-letter backlog matches the current filters.` instead of the unfiltered empty-state message. Backlog row-navigation hints appear only when there is another backlog row to move through.

When retry confirmation is pending for the selected row, the filter hint line switches to confirm-state guidance instead of the generic apply hint: `Enter` confirms the retry request, `Esc` cancels it, and `Ctrl+R` still resets filters.
While a retry request is actively running, both the help bar and the filter hint line drop `Ctrl+R` so neither surface advertises a reset action that the runtime is currently holding back.

### Dead Letters Detail And Retry Flow

Selecting a backlog row shows canonical transaction detail, latest retry metadata, background task status when present, and the current retry action state. Retry uses a two-step request flow:

1. Press `r` once to enter confirm state for the selected retryable row.
2. Press `r` again or `Enter` to submit the retry request.

While confirm state is pending, `Esc` cancels the retry request without leaving the current row.

The detail pane then reports either:

- `Retry request accepted ...` followed by refresh/follow-up notes when the request was submitted
- `Retry request failed ...` when the retry request itself fails
- `Retry action: disabled` when the selected row is not retryable

### Dead Letters Degraded And Failure States

Dead Letters remains routable even before the bridge is ready, but the page surfaces explicit degraded states instead of pretending the backlog is merely empty:

- no list callback: `Dead-letter backlog is not configured.`
- configured list callback failure: `Failed to load dead letters: ...`
- no rows with no active filters: `No current dead-letter backlog.`
- selected-detail load failure: `Failed to load detail: ...`

Displayed dead-letter transaction identifiers, reasons, adjudication labels, dispatch references, actor labels, summary-strip fields, detail values, and retry status messages are normalized to plain single-line text before rendering.

## Troubleshooting

Start with `lango doctor` when cockpit behavior looks wrong. The most common issues are:

- startup problems caused by a non-interactive terminal or unsupported alt-screen behavior
- empty context or runtime panels when the metrics collector has no data yet or no active turn is running
- missing channel messages or approval prompts when `--with-channels` is not enabled or channel credentials are invalid
- task actions that appear to do nothing because the selected task state does not allow retry or cancel

Check `<DataRoot>/cockpit.log` for the underlying error details when the TUI does not show enough context; the default path is `~/.lango/cockpit.log`. Use a modern terminal with TTY and alt-screen support, and verify the channel/task wiring before assuming the cockpit UI is broken.
