## 1. Turn Budget And Response Boundaries

- [x] 1.1 Raise single-agent and multi-agent effective turn defaults in runtime/config-facing output
- [x] 1.2 Filter isolated sub-agent text out of user-visible collection/streaming paths
- [x] 1.3 Stop exposing raw partial drafts in channel and gateway user responses
- [x] 1.4 Stop persisting raw partial drafts in timeout session annotations
- [x] 1.5 Update/extend tests for turn defaults and safe partial handling

## 2. Browser Tooling Upgrade

- [x] 2.1 Add structured browser snapshot types and richer `browser_navigate` output
- [x] 2.2 Add `browser_search` high-level tool and result extraction heuristics
- [x] 2.3 Add `browser_observe` and `browser_extract` high-level tools
- [x] 2.4 Add/extend browser tests for structured extraction helpers and tool registration

## 3. Downstream Sync

- [x] 3.1 Update approval summaries, prompts, and navigator agent definitions for the new browser tools and response boundaries
- [x] 3.2 Update CLI/docs/README references for new defaults and browser capabilities
- [x] 3.3 Run `go build ./...` and `go test ./...`
