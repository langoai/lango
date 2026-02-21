---
title: Tool Approval
---

# Tool Approval

Lango provides an approval workflow for sensitive tool executions. When an agent attempts to call a tool that requires approval, the execution is paused until a human approves or the request times out.

## Approval Policies

The `approvalPolicy` setting determines which tools require human approval before execution:

| Policy | Behavior |
|--------|----------|
| `dangerous` | Requires approval for tools marked as dangerous-level (default) |
| `all` | Requires approval for every tool call |
| `configured` | Requires approval only for tools listed in `sensitiveTools` |
| `none` | Disables approval entirely -- all tools execute immediately |

!!! danger "Policy: none"

    Setting `approvalPolicy: none` disables all safety checks for tool execution. Only use this in fully trusted, isolated environments.

```yaml
security:
  interceptor:
    enabled: true
    approvalPolicy: dangerous    # dangerous | all | configured | none
```

## Sensitive Tools

When using `approvalPolicy: configured`, you must explicitly list which tools require approval:

```yaml
security:
  interceptor:
    approvalPolicy: configured
    sensitiveTools:
      - exec            # Shell command execution
      - browser         # Browser automation
      - filesystem      # File system operations
      - wallet_send     # Blockchain transactions
```

## Exempt Tools

Tools listed in `exemptTools` are exempt from approval regardless of the active policy. This is useful when a broad policy like `all` is active but certain safe tools should always execute immediately:

```yaml
security:
  interceptor:
    approvalPolicy: all
    exemptTools:
      - knowledge_search    # Always allow knowledge lookups
      - memory_recall       # Always allow memory queries
```

!!! warning "Exempt Overrides Policy"

    `exemptTools` takes precedence over both the approval policy and `sensitiveTools`. A tool listed in both `sensitiveTools` and `exemptTools` will be exempt.

## Approval Timeout

The `approvalTimeoutSec` setting controls how long the system waits for human approval before the tool call is rejected:

```yaml
security:
  interceptor:
    approvalTimeoutSec: 30    # seconds (default: 30)
```

If the timeout expires without approval, the tool call is denied and the agent receives an error.

## Notification Channel

Configure which messaging channel receives approval notifications. When a tool requires approval, a notification is sent to the specified channel with details about the pending tool call:

```yaml
security:
  interceptor:
    notifyChannel: telegram    # telegram | discord | slack
```

The notification includes:

- Tool name
- Input parameters (with secrets masked)
- Requesting session ID
- Approve/deny action buttons (channel-dependent)

## Headless Auto-Approve

For CI/CD or automated deployments where no human is available to approve, enable headless auto-approve:

```yaml
security:
  interceptor:
    headlessAutoApprove: true
```

!!! warning "Security Risk"

    Headless auto-approve bypasses the approval workflow entirely. Use only in controlled environments where the agent's tool access is already restricted by other means.

## Configuration Reference

```yaml
security:
  interceptor:
    enabled: true
    approvalPolicy: dangerous        # dangerous | all | configured | none
    sensitiveTools:                   # Tools requiring approval (configured policy)
      - exec
      - browser
    exemptTools:                      # Tools exempt from approval
      - knowledge_search
    approvalTimeoutSec: 30           # Timeout in seconds
    notifyChannel: telegram          # Channel for approval notifications
    headlessAutoApprove: false       # Auto-approve in headless mode
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `approvalPolicy` | string | `"dangerous"` | Which tools require approval |
| `sensitiveTools` | list | `[]` | Tool names requiring approval (`configured` policy) |
| `exemptTools` | list | `[]` | Tool names exempt from approval |
| `approvalTimeoutSec` | int | `30` | Seconds to wait for approval |
| `notifyChannel` | string | `""` | Channel for approval notifications |
| `headlessAutoApprove` | bool | `false` | Auto-approve all tools in headless mode |
