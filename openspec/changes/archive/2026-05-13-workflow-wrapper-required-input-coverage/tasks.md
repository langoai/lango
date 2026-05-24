## 1. Workflow Guard Coverage

- [x] 1.1 Add tool-entrypoint regression coverage for missing `run_id` on `workflow_status` and `workflow_cancel`.
- [x] 1.2 Add tool-entrypoint regression coverage for missing `name` and `yaml_content` on `workflow_save`.

## 2. Prompt And Docs Sync

- [x] 2.1 Update automator runtime and embedded prompts for the `workflow_*` required-input contract.
- [x] 2.2 Update TOOL_USAGE, README, and multi-agent docs for the same contract.
- [x] 2.3 Update workflow, agent-prompting, downstream-docs, and production-readiness specs for the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/workflow -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
