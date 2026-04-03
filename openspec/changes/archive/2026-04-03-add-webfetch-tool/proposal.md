## Why

Agents need a lightweight way to fetch and extract content from web pages without requiring a full browser session. The existing `browser_navigate` tool launches a headless browser which is heavyweight for simple content retrieval tasks. A dedicated HTTP fetch tool provides a fast, resource-efficient alternative for reading web content.

## What Changes

- Add new `internal/tools/webfetch` package with HTTP fetch + content extraction
- Implement readability extraction using `golang.org/x/net/html` (strip nav/footer/aside/script/style, find article/main/largest div)
- Support three output modes: `text` (clean extracted text), `html` (raw HTML), `markdown` (simplified markdown with headings, links, lists)
- P2P URL validation to block internal/private network access in peer contexts
- Content truncation with configurable `max_length` parameter

## Capabilities

### New Capabilities
- `tool-webfetch`: HTTP web page fetch and content extraction tool with text, HTML, and markdown output modes

### Modified Capabilities
<!-- No existing capabilities are modified -->

## Impact

- New package: `internal/tools/webfetch/` (fetch.go, readability.go, tools.go)
- Uses existing dependency `golang.org/x/net/html` (already in go.mod)
- Uses existing `internal/agent` (Tool, Schema, SafetyLevel, ToolCapability)
- Uses existing `internal/ctxkeys` (P2P request detection)
- Uses existing `internal/toolparam` (parameter extraction helpers)
- No breaking changes to existing code
