## REMOVED Requirements

### Requirement: Remote browser WebSocket connection
**Reason**: Single Docker image always includes Chromium. Remote browser WebSocket connections are no longer needed since there is no sidecar deployment pattern.
**Migration**: Remove `tools.browser.remoteBrowserUrl` from configuration. Remove `ROD_BROWSER_WS` environment variable. Use the unified Docker image with local Chromium.
