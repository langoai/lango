## 1. Changed-Target Awareness

- [x] 1.1 Parse changed targets from lightweight Git status output.
- [x] 1.2 Include the changed-target summary in the dirty-repository starter prompt.
- [x] 1.3 Preserve existing Git-aware fallback wording when changed targets are unavailable.

## 2. Verification

- [x] 2.1 Add unit coverage for single-target and multi-target dirty repository prompts.
- [x] 2.2 Run focused workbench tests.
- [x] 2.3 Run `go build ./...`.
- [x] 2.4 Run `go test ./...`.

## 3. Downstream Sync

- [x] 3.1 Update README and CLI/TUI docs to mention changed-target-aware starter prompts.
- [ ] 3.2 Validate and archive the OpenSpec change.
