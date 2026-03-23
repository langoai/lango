---
title: Approval System
---

# Approval System

The approval system provides a unified interface for tool execution authorization across multiple channels. When sensitive tools are invoked, the system routes approval requests to the appropriate provider based on the session context.

## Architecture

```
Tool Invocation ──► CompositeProvider ──► Channel Routing
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
         Gateway      Channel     TTY Fallback
        (WebSocket)   (Telegram/   (Terminal)
                      Discord/
                      Slack)
```

## Providers

### CompositeProvider

The `CompositeProvider` is the central router. It evaluates registered providers in order and routes to the first one whose `CanHandle()` returns true for the given session key.

**Routing rules:**
- P2P sessions (`p2p:...` keys) use a dedicated P2P fallback — **never** the headless provider
- Non-P2P sessions fall back to the TTY provider when no other provider matches
- If no provider matches, the request is denied (fail-closed)

### GatewayProvider

Routes approval requests to connected companion apps via WebSocket. Active when at least one companion is connected to the gateway.

### TTYProvider

Prompts the terminal user via stdin/stderr. Supports three responses:

| Input | Behavior |
|-------|----------|
| `y` / `yes` | Approve this single invocation |
| `a` / `always` | Approve and grant persistent access for this tool in this session |
| `N` / anything else | Deny |

TTY approval is unavailable when stdin is not a terminal (e.g., Docker containers, background processes).

### One-shot Approve vs Always Allow

Lango distinguishes between two approval scopes:

| Action | Scope |
|--------|-------|
| `Approve` / `yes` | Current request only |
| `Always Allow` / `always` | Session-wide persistent grant |

A normal one-shot approval now also acts as a **turn-local replay grant** for the current request. If the agent retries the same canonical approval action later in the same turn, Lango reuses the approval result instead of opening a second identical prompt.

Canonicalization can ignore approval-neutral params. For example, `browser_search` shares the same turn-local approval state across `limit`-only variants of the same query.

Likewise, if the same canonical action was denied in the current request, Lango blocks the duplicate retry immediately. Timeouts are retryable only within a bounded per-turn budget; once that budget is exhausted, later identical retries are blocked until the next user turn.

### HeadlessProvider

Auto-approves all requests with WARN-level audit logging. Intended for headless environments (Docker, CI) where no interactive approval is possible.

**Security**: HeadlessProvider is **never** used for P2P sessions. Remote peers cannot trigger auto-approval.

## Grant Store

The `GrantStore` tracks per-session, per-tool "always allow" grants in memory. When a user selects "always" on a TTY prompt, subsequent invocations of that tool in the same session are auto-approved.

**Properties:**
- In-memory only — grants are cleared on application restart
- Optional TTL — grants can expire automatically via `SetTTL()`
- Scoped — grants are per session key + tool name
- Revocable — individual grants or entire sessions can be revoked

### Grant Lifecycle

| Method | Description |
|--------|-------------|
| `Grant(sessionKey, toolName)` | Record an approval |
| `IsGranted(sessionKey, toolName)` | Check if a valid grant exists |
| `Revoke(sessionKey, toolName)` | Remove a single grant |
| `RevokeSession(sessionKey)` | Remove all grants for a session |
| `CleanExpired()` | Remove expired grants (when TTL is set) |

This store is separate from the request-scoped turn-local approval cache used to suppress duplicate prompts within a single agent turn.

## Approval Request

Each approval request contains:

| Field | Type | Description |
|-------|------|-------------|
| `ID` | string | Unique request identifier |
| `ToolName` | string | Name of the tool requiring approval |
| `SessionKey` | string | Session key (determines routing) |
| `Params` | map | Tool invocation parameters |
| `Summary` | string | Human-readable description of the action |
| `CreatedAt` | time | Request timestamp |

## Approval Policies

The approval policy controls which tools require approval:

| Policy | Behavior |
|--------|----------|
| `dangerous` | Only tools marked as dangerous require approval |
| `all` | All tool invocations require approval |
| `configured` | Only tools explicitly listed in the policy require approval |
| `none` | No approval required (all tools auto-approved) |

Configure via `security.interceptor.approvalPolicy`:

```json
{
  "security": {
    "interceptor": {
      "enabled": true,
      "approvalPolicy": "dangerous"
    }
  }
}
```

## P2P Approval Pipeline

Inbound P2P tool invocations pass through a three-stage approval pipeline:

1. **Firewall ACL** — Static allow/deny rules by peer DID and tool pattern
2. **Reputation Check** — Peer trust score must exceed `minTrustScore`
3. **Owner Approval** — Interactive approval via the composite provider

Small paid tool invocations can be auto-approved when the amount is below `payment.limits.autoApproveBelow`.

## CLI Commands

```bash
lango approval status        # Show approval system configuration
```

## Configuration

| Setting | Default | Description |
|---------|---------|-------------|
| `security.interceptor.enabled` | `false` | Enable the security interceptor |
| `security.interceptor.approvalPolicy` | `"dangerous"` | Approval policy for tool invocations |
| `security.interceptor.redactPII` | `false` | Redact PII in tool inputs/outputs |
