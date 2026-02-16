## Why

When `approvalPolicy: "all"` is configured and a Telegram/Discord/Slack user requests a tool action, the bot responds with "tool execution denied" instead of sending an approval request (inline keyboard/button). This is because `runAgent()` passes a bare `context.Background()` without the session key, so `CompositeProvider` cannot match any channel provider and falls back to silent denial.

## What Changes

- Inject session key into context in `runAgent()` so approval providers can route requests to the correct channel
- Improve `wrapWithApproval` error messages to distinguish "no channel available" from "user denied"
- Change `CompositeProvider` fail-closed path to return a sentinel error instead of silent `(false, nil)`
- Add "Tool Approval" guidance to the system prompt so the AI correctly interprets denial vs. configuration errors

## Capabilities

### New Capabilities

### Modified Capabilities
- `channel-approval`: Session key must be present in context for channel-based approval routing to work; fail-closed path now returns an error
- `approval-policy`: `wrapWithApproval` error messages now differentiate between missing session key and user denial

## Impact

- `internal/app/channels.go`: 1-line context injection in `runAgent()`
- `internal/app/tools.go`: Error message branching in `wrapWithApproval` handler
- `internal/approval/composite.go`: Sentinel error on no-provider-matched path
- `internal/approval/approval_test.go`: Updated test expectation for fail-closed error
- `prompts/TOOL_USAGE.md`: New "Tool Approval" section for AI guidance
