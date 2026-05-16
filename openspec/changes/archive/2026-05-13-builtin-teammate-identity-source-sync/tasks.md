## 1. Source Of Truth Sync

- [x] 1.1 Embed per-agent prompt files in `prompts.FS`.
- [x] 1.2 Make built-in agent specs prefer embedded `IDENTITY.md` content.
- [x] 1.3 Add regression coverage that embedded prompt files and runtime agent specs stay aligned.

## 2. Spec Sync

- [x] 2.1 Update `sub-agent-default-prompts` to describe embedded identity files as the preferred source for built-in instructions.

## 3. Verification

- [x] 3.1 Run `go test ./internal/orchestration -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
