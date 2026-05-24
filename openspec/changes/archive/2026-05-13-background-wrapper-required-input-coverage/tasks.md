## 1. Background Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for missing `prompt` on `bg_submit`.
- [x] 1.2 Add tool-entrypoint regression coverage for missing `task_id` on `bg_status`, `bg_result`, and `bg_cancel`.

## 2. Prompt And Docs Sync

- [x] 2.1 Update automator runtime and embedded prompts for the `bg_*` required-input contract.
- [x] 2.2 Update TOOL_USAGE, README, and multi-agent docs for the same contract.
- [x] 2.3 Update background, agent-prompting, downstream-docs, and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/background -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
