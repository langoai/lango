## Why

The current approval system has solid infrastructure (per-channel providers, fail-closed design) but defaults to **disabled** — users must explicitly set `approvalRequired: true` plus a `sensitiveTools` list, otherwise all tools execute without approval. Additionally, tools lack intrinsic risk classification, and approval messages show no detail about what will be executed. This change makes the system default-safe and information-rich.

## What Changes

- Add a `SafetyLevel` enum (Safe/Moderate/Dangerous) to `agent.Tool`, with zero-value treated as Dangerous (fail-safe)
- Assign SafetyLevel to every built-in tool based on its risk profile
- Introduce `ApprovalPolicy` config field (`dangerous`/`all`/`configured`/`none`) replacing the boolean `approvalRequired`
- Default new installations to `"dangerous"` policy — Dangerous tools require approval out of the box
- Add `ExemptTools` config list for opt-out overrides
- Auto-migrate legacy `approvalRequired` + `sensitiveTools` configs to the new policy model
- Add `Summary` field to `ApprovalRequest` with human-readable execution descriptions
- Render Summary in all approval channels (Gateway, TTY, Headless, Telegram, Discord, Slack)
- **BREAKING**: `approvalRequired` field is deprecated in favor of `approvalPolicy`

## Capabilities

### New Capabilities
- `tool-safety-level`: Intrinsic risk classification system for agent tools with SafetyLevel enum and fail-safe zero-value behavior
- `approval-policy`: Policy-based approval system with four modes (dangerous/all/configured/none), exempt tools, and automatic legacy config migration

### Modified Capabilities
- `channel-approval`: ApprovalRequest gains a Summary field; all providers render execution details in approval messages
- `ai-privacy-interceptor`: Approval workflow updated from boolean gate to policy-based gate with SafetyLevel awareness

## Impact

- `internal/agent/runtime.go` — SafetyLevel type and Tool struct field
- `internal/config/types.go` — ApprovalPolicy type, InterceptorConfig fields
- `internal/config/loader.go` — Migration logic, default config, viper defaults
- `internal/app/tools.go` — SafetyLevel assignments, needsApproval, buildApprovalSummary, wrapWithApproval refactor
- `internal/app/app.go` — Policy-based approval gate
- `internal/approval/` — Summary field in ApprovalRequest, rendering in all providers
- `internal/channels/{telegram,discord,slack}/approval.go` — Summary rendering in channel-specific messages
