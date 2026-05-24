## 1. Wrapper Param Guards

- [x] 1.1 Enforce the required `query` parameter for `search_knowledge`.
- [x] 1.2 Add `get_knowledge_history` and `search_knowledge` wrapper coverage for missing required inputs.

## 2. Verification

- [x] 2.1 Run `go test ./internal/app -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `meta-tools` and `production-readiness` coverage for knowledge read-tool wrapper params.
- [ ] 3.2 Validate and archive the OpenSpec change.
