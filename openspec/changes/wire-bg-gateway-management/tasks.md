## 1. Planning

- [x] 1.1 Create OpenSpec artifacts for gateway-backed background management
- [x] 1.2 Validate the active OpenSpec change
- [ ] 1.3 Commit planning artifacts as a scoped commit

## 2. Gateway API

- [ ] 2.1 Add failing route tests for `/api/bg/tasks`, `/api/bg/tasks/{id}`, `/api/bg/tasks/{id}/result`, and `/api/bg/tasks/{id}/cancel`
- [ ] 2.2 Implement authenticated background management routes on the app gateway router
- [ ] 2.3 Return stable DTOs and proper `503`, `404`, and non-2xx error responses

## 3. CLI

- [ ] 3.1 Add failing CLI tests for gateway-backed root bg management and `--addr`
- [ ] 3.2 Refactor `internal/cli/bg` behind a small client interface with in-process and gateway adapters
- [ ] 3.3 Add GET/POST JSON client helpers needed by bg remote management
- [ ] 3.4 Wire root `lango bg` to the gateway-backed client via `cmd/lango/main.go`
- [ ] 3.5 Preserve in-process manager behavior for embedded callers and existing tests

## 4. Downstream Artifacts

- [ ] 4.1 Update README, CLI index, and background automation docs to describe gateway-backed bg management
- [ ] 4.2 Update executable docs guards for the new gateway wording
- [ ] 4.3 Sync main OpenSpec specs after implementation is verified

## 5. Review And Verification

- [ ] 5.1 Run subagent spec compliance review
- [ ] 5.2 Run subagent code/docs quality review
- [ ] 5.3 Run `go build ./...`
- [ ] 5.4 Run `go test ./...`
- [ ] 5.5 Run `git diff --check`
- [ ] 5.6 Run `openspec validate --all --strict`
- [ ] 5.7 Archive the completed OpenSpec change
