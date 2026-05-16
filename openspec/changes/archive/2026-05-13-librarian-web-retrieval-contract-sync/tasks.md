## 1. Web Retrieval Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for missing `query` on `web_search`.
- [x] 1.2 Add tool-entrypoint regression coverage for missing `url` on `web_fetch`.

## 2. Prompt And Docs Sync

- [x] 2.1 Update librarian runtime and embedded prompts for lightweight web retrieval.
- [x] 2.2 Update TOOL_USAGE, README, and multi-agent docs for librarian `web_*` routing and required inputs.
- [x] 2.3 Update web tool, agent-prompting, downstream-docs, and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/tools/websearch ./internal/tools/webfetch -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
