## 1. Cron Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for missing `name`, `schedule_type`, `schedule`, and `prompt` on `cron_add`.
- [x] 1.2 Add tool-entrypoint regression coverage for missing `id` on `cron_pause`, `cron_resume`, and `cron_remove`.

## 2. Prompt And Docs Sync

- [x] 2.1 Update automator runtime and embedded prompts for the `cron_*` required-input contract.
- [x] 2.2 Update TOOL_USAGE, README, and multi-agent docs for the same contract.
- [x] 2.3 Update cron, agent-prompting, downstream-docs, and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cron -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
