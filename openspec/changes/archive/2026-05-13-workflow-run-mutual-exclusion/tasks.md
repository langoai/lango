## 1. Contract Sync

- [x] 1.1 Reject dual `file_path` + `yaml_content` input in `workflow_run`.
- [x] 1.2 Add regression coverage for the mutual-exclusion path.
- [x] 1.3 Update workflow tool spec wording for the mutual-exclusion contract.

## 2. Verification

- [x] 2.1 Run `go test ./internal/workflow -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `automation-agent-tools` coverage for `workflow_run` mutual exclusion.
- [x] 3.2 Validate and archive the OpenSpec change.
