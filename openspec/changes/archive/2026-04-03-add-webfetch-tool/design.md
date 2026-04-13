## Context

The agent tool system provides domain-specific packages under `internal/tools/` (browser, exec, filesystem, crypto, secrets). The browser tools require a headless browser session (via go-rod) which is heavyweight for simple page content retrieval. A lightweight HTTP-only fetch tool fills this gap.

The project already depends on `golang.org/x/net/html` (used by browser package), so no new dependencies are needed.

## Goals / Non-Goals

**Goals:**
- Provide a single `web_fetch` tool that retrieves and extracts web page content
- Support three output modes: plain text, raw HTML, markdown
- Implement basic readability extraction (strip navigation/footer/scripts, find article/main content)
- Enforce P2P URL safety (block private/loopback addresses)
- Follow existing tool patterns (Schema builder, toolparam, SafetyLevel, ToolCapability)

**Non-Goals:**
- JavaScript rendering (use browser tools for that)
- Cookie/session management
- POST requests or form submission
- PDF or binary content extraction
- Full readability algorithm (Mozilla's readability.js equivalent)

## Decisions

### Package structure: three files
`fetch.go` (HTTP client + orchestration), `readability.go` (HTML parsing + extraction), `tools.go` (agent tool registration). This mirrors the separation of concerns in other tool packages.

### Readability approach: simple DOM traversal
Use `golang.org/x/net/html` to find `<article>`, `<main>`, or largest `<div>` by text length. Strip `<nav>`, `<footer>`, `<aside>`, `<script>`, `<style>`, `<noscript>`, `<iframe>`, `<svg>`. This is simpler than a full readability algorithm but sufficient for most content pages.

Alternative considered: importing a Go readability library. Rejected to avoid new dependencies for a first iteration.

### P2P validation: own copy of URL validator
Rather than importing from `internal/tools/browser`, the webfetch package has its own `ValidateURLForP2P` with the same logic. This avoids coupling webfetch to the browser package and follows the existing pattern where each tool package is self-contained.

### Safety level: Moderate
The tool performs network access (read-only HTTP GET) but cannot execute code or modify data. `SafetyLevelModerate` is appropriate per the safety level spec.

## Risks / Trade-offs

- [Content extraction quality] Simple DOM heuristics may miss content on heavily JavaScript-rendered pages → Agents can fall back to browser tools for complex pages
- [Memory usage on large pages] Body limited to 5MB read cap → Sufficient for text content; binary/media pages will be truncated harmlessly
- [Duplicated P2P validation] Same logic exists in browser package → Acceptable for package isolation; could be extracted to a shared package later if more tools need it
