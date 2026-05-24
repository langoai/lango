## 1. Onboard Next-Step Guidance

- [x] 1.1 Add a regression test that captures onboard post-save stdout and asserts `lango`, `lango serve`, `lango doctor`, and `lango settings` are all mentioned.
- [x] 1.2 Update onboard post-save messaging so the next-step block matches the real workbench, live runtime, verification, and full-editor entry points.

## 2. Documentation Truth Alignment

- [x] 2.1 Update `README.md` so prompts, embedding, graph, multi-agent, A2A, security, and OIDC guidance no longer reference nonexistent onboard submenu paths.
- [x] 2.2 Update advanced feature docs to point users at `lango settings` or config import/export instead of false onboard submenu flows.

## 3. Verification

- [x] 3.1 Run `gofmt -w` on modified Go files.
- [x] 3.2 Run `go test ./internal/cli/onboard -count=1`.
- [ ] 3.3 Run `go build ./...`.
- [ ] 3.4 Run `go test ./...`.
- [ ] 3.5 Validate and archive the OpenSpec change.
