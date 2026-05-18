## Why

The Slack SDK defaults to an `http.Client` without a timeout. Lango's Slack channel uses that default when no test client is injected, so Slack REST calls such as auth checks, message posts, updates, and Socket Mode connection setup can rely on an unbounded HTTP client.

## What Changes

- Provide a finite default HTTP client timeout for Slack channel REST/API calls.
- Preserve injected HTTP clients for tests and custom callers.
- Add regression coverage for the default and injected-client paths.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `channel-slack`: Slack channel API clients use a bounded default HTTP timeout when no custom client is supplied.

## Impact

- Affected code: `internal/channels/slack/slack.go`.
- Affected tests: `internal/channels/slack`.
- Affected specs: `openspec/specs/channel-slack/spec.md`.
