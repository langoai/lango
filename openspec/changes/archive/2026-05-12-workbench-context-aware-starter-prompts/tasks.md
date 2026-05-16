## 1. Workspace-Aware Prompting

- [x] 1.1 Add a workbench helper that derives starter prompt context from the current workdir.
- [x] 1.2 Detect repository and Go-module markers without invoking external tooling.
- [x] 1.3 Feed the derived starter prompt set into the standalone workbench Mission Control page.

## 2. Verification

- [x] 2.1 Add unit coverage for generic fallback and nested Go-repo detection.
- [x] 2.2 Add workbench coverage proving the rendered empty state uses context-aware prompts.
- [x] 2.3 Run focused tests for workbench, Mission Control pages, and `cmd/lango`.
- [x] 2.4 Run `go build ./...`.
- [x] 2.5 Run `go test ./...`.

## 3. Downstream Sync

- [x] 3.1 Update README and CLI/TUI docs to mention context-aware starter prompts.
- [ ] 3.2 Validate and archive the OpenSpec change.
