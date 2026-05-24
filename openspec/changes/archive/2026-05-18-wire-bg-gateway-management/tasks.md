## 1. Planning

- [x] 1.1 Create OpenSpec artifacts for gateway-backed background management
- [x] 1.2 Validate the active OpenSpec change
- [ ] 1.3 Commit planning artifacts as a scoped commit

## 2. Gateway API

- [x] 2.1 Add failing route tests for `/api/bg/tasks`, `/api/bg/tasks/{id}`, `/api/bg/tasks/{id}/result`, and `/api/bg/tasks/{id}/cancel`
- [x] 2.2 Implement authenticated background management routes on the app gateway router
- [x] 2.3 Return stable DTOs and proper `503`, `404`, and non-2xx error responses

## 3. CLI

- [x] 3.1 Add failing CLI tests for gateway-backed root bg management and `--addr`
- [x] 3.2 Refactor `internal/cli/bg` behind a small client interface with in-process and gateway adapters
- [x] 3.3 Add GET/POST JSON client helpers needed by bg remote management
- [x] 3.4 Wire root `lango bg` to the gateway-backed client via `cmd/lango/main.go`
- [x] 3.5 Preserve in-process manager behavior for embedded callers and existing tests

## 4. Downstream Artifacts

- [x] 4.1 Update README, CLI index, and background automation docs to describe gateway-backed bg management
- [x] 4.2 Update executable docs guards for the new gateway wording
- [x] 4.3 Sync main OpenSpec specs after implementation is verified

## 5. Review And Verification

- [x] 5.1 Run subagent spec compliance review
- [x] 5.2 Run subagent code/docs quality review
- [x] 5.3 Run `go build ./...`
- [x] 5.4 Run `go test ./...`
- [x] 5.5 Run `git diff --check`
- [x] 5.6 Run `openspec validate --all --strict`
- [x] 5.7 Archive the completed OpenSpec change
