---
title: Cockpit Approval Operator Guide
---

# Cockpit Approval Operator Guide

This guide covers how to interact with the approval system from within the cockpit TUI. For the full approval system reference, see [Tool Approval](../security/tool-approval.md) and [Approval System Architecture](../security/approval-cli.md).

## Overview

When an agent invokes a sensitive tool, the cockpit pauses execution and presents an interactive approval prompt. The operator reviews the request and responds with a keypress to allow, deny, or grant a session-wide permission. The cockpit renders different UI elements depending on the risk level of the tool, ensuring dangerous operations receive proportional scrutiny.

## Approval Policies

Four policies control which tools require approval:

| Policy | Behavior |
|--------|----------|
| `dangerous` | Requires approval only for dangerous-level tools (default) |
| `all` | Every tool call requires approval |
| `configured` | Only tools listed in `sensitiveTools` require approval |
| `none` | All tools execute immediately -- no approval prompts |

For full policy configuration, exemptions, and timeout settings, see [Tool Approval -- Approval Policies](../security/tool-approval.md#approval-policies).

## Safety Levels

Each tool is classified into one of three safety levels:

| Level | Description |
|-------|-------------|
| `safe` | Read-only operations with no side effects |
| `moderate` | Creates or modifies resources with limited blast radius |
| `dangerous` | Filesystem mutations, arbitrary code execution, or irreversible actions |

The safety level, combined with the tool's category and activity, determines how the approval prompt is displayed. See [Approval System](../security/approval-cli.md#approval-policies) for how policies interact with safety levels.

## Two-Tier Display

The cockpit uses two distinct rendering modes for approval prompts, selected automatically based on the tool's classification.

### Tier 1: Inline Strip

Used for moderate-risk tools and dangerous tools that do not target filesystem/automation or perform write/execute activities.

The strip is a single-line bar rendered at the bottom of the chat view:

```
 browser_search  Search for "sensitive query"  [a]llow  [s]ession  [d]eny
```

Elements shown:

- **Tool name** -- bold, highlighted
- **Summary** -- human-readable description of the action, truncated to fit
- **Keys** -- `[a]llow`, `[s]ession`, `[d]eny`

For channel-originated requests, a badge prefix appears before the summary (e.g., `[TG]`, `[DC]`, `[SL]`).

### Tier 2: Fullscreen Dialog

Used for dangerous tools that target the filesystem or automation categories, or perform write/execute activities. The dialog is a bordered overlay that shows:

- **Risk badge** -- color-coded by risk level (critical = red, high = orange, moderate = blue, low = gray)
- **Tool name** -- bold, highlighted
- **Risk label** -- human-readable description (e.g., "Modifies filesystem", "Executes arbitrary code")
- **Channel origin** -- for channel sessions, shows `[Telegram] 123456` or similar
- **Summary** -- full description of the proposed action
- **Rule explanation** -- italic text explaining why approval is required
- **Parameters** -- scrollable key-value display of tool input parameters (values truncated to 120 characters)
- **Diff preview** -- for `fs_write` and `fs_edit` tools, a unified diff showing proposed file changes with syntax coloring (`+` lines in green, `-` lines in red, `@@` headers in blue). Scrollable with up/down keys. Truncated at 500 lines.

Keys in fullscreen mode: `a` (allow), `s` (session), `d`/`esc` (deny), up/down (scroll diff), `t` (toggle split/unified diff mode).

### Classification Logic

The tier is determined by `ClassifyTier(safetyLevel, category, activity)`:

- If `safetyLevel` is not `"dangerous"` -- always Tier 1 (inline)
- If `safetyLevel` is `"dangerous"` AND (`category` is `"filesystem"` or `"automation"`, OR `activity` is `"execute"` or `"write"`) -- Tier 2 (fullscreen)
- Otherwise -- Tier 1 (inline)

## Risk Levels

Each approval request receives a computed risk indicator that determines visual styling:

| Risk Level | Criteria | Label |
|------------|----------|-------|
| **critical** | `dangerous` safety + `filesystem` category | "Modifies filesystem" |
| **critical** | `dangerous` safety + `automation` category | "Executes arbitrary code" |
| **high** | `dangerous` safety + any other category | "Dangerous operation" |
| **moderate** | `moderate` safety level | "Creates or modifies resources" |
| **low** | `safe` or unclassified safety level | "Read-only operation" |

Risk level affects the color of the risk badge in the fullscreen dialog and determines whether the double-press guardrail activates.

## Double-Press Guardrail

For **critical-risk** tools (dangerous + filesystem/automation), the cockpit requires a two-step confirmation to prevent accidental approval of destructive operations.

### How it works

1. The operator presses `a` (allow) or `s` (session allow).
2. Instead of immediately executing, the UI enters a **confirm-pending** state. The action bar changes to: `Press 'a' again to confirm (destructive operation)`. In the inline strip, the `(destructive)` label appears next to the tool name.
3. The operator must press the **same key** again within **3 seconds** to confirm.
4. If the 3-second window expires, the confirm-pending state resets and the operator must start over.
5. If the operator presses a **different key** during the confirm window, the pending state is cancelled.

### Summary

| Step | What happens |
|------|-------------|
| First press of `a` or `s` | Shows confirmation prompt, no action taken |
| Same key within 3 seconds | Executes the approval action |
| Different key pressed | Cancels confirmation, returns to initial state |
| 3-second timeout | Resets to initial state |

The double-press guardrail only applies to critical-risk tools. Non-critical tools respond immediately to a single keypress.

## Grant Management in Cockpit

### Creating Session Grants

Press `s` on any approval prompt to create a session-wide grant for that tool. All subsequent invocations of the same tool in the same session will be auto-approved without prompting.

For critical-risk tools, the `s` key also requires the double-press confirmation described above.

### Viewing Grants

Navigate to the **Approvals** page in the cockpit. The page has two sections:

- **Approval History** -- decision log (top section)
- **Active Grants** -- currently active session grants (bottom section)

Press `/` to toggle focus between the two sections. The active section's title is highlighted in the accent color.

The grants table shows three columns:

| Column | Description |
|--------|-------------|
| Session | The session key (truncated) |
| Tool | The tool name |
| Granted | Relative timestamp (e.g., "2m ago") |

### Revoking Individual Grants

1. Navigate to the **Approvals** page.
2. Press `/` to switch to the **Active Grants** section.
3. Use `up`/`down` (or `k`/`j`) to select the grant to revoke.
4. Press `r` to revoke the selected grant.

### Revoking All Session Grants

1. Navigate to the **Active Grants** section as above.
2. Select any grant belonging to the target session.
3. Press `R` (uppercase) to revoke all grants for that session.

Grant properties:

- Grants are in-memory only -- cleared on application restart
- Grants can have an optional TTL that causes automatic expiry
- Grants are scoped per session key + tool name

## History Viewing

The **Approvals** page shows a chronological decision log in the **Approval History** section.

### Columns

| Column | Description |
|--------|-------------|
| Time | Relative timestamp of the decision |
| Tool | Name of the tool |
| Summary | Human-readable description of the action |
| Outcome | Decision result (see below) |
| Provider | Which provider handled the request (e.g., `tui`, `gateway`, `headless`) |

### Outcome Values

The outcome column uses an open set of values, including:

- `bypass` -- tool was exempt or auto-approved
- `granted` -- operator approved the request
- `denied` -- operator denied the request
- `timeout` -- approval timed out before response
- `replay_blocked` -- duplicate request blocked by turn-local cache
- `unavailable` -- no approval provider was available

### Navigation

- `up`/`down` (or `k`/`j`) to scroll through entries
- `/` to toggle between History and Grants sections
- History is displayed newest-first
- The page auto-refreshes every 2 seconds while active

## Channel Approval

When tool invocations originate from messaging channels (Telegram, Discord, Slack), the cockpit displays the channel origin in the approval prompt:

- **Inline strip**: a badge prefix appears before the summary -- `[TG]` for Telegram, `[DC]` for Discord, `[SL]` for Slack
- **Fullscreen dialog**: a full origin line appears below the header -- e.g., `<- [Telegram] 123456`

The approval decision applies to the originating channel session. Session grants created via `s` are scoped to that channel's session key.

The `notifyChannel` configuration field can route approval notifications to a specific channel. See [Tool Approval -- Notification Channel](../security/tool-approval.md#notification-channel) for setup.

## Configuration Reference

All approval settings are under `security.interceptor` in the configuration. Key fields:

| Key | Default | Description |
|-----|---------|-------------|
| `approvalPolicy` | `"dangerous"` | Which tools require approval |
| `sensitiveTools` | `[]` | Tool names requiring approval (`configured` policy) |
| `exemptTools` | `[]` | Tool names exempt from approval |
| `approvalTimeoutSec` | `30` | Seconds to wait before timeout |
| `notifyChannel` | `""` | Channel for approval notifications |
| `headlessAutoApprove` | `false` | Auto-approve in headless mode |

Configure via `lango settings` or edit the config file directly. For the full reference with examples, see [Tool Approval -- Configuration Reference](../security/tool-approval.md#configuration-reference).
