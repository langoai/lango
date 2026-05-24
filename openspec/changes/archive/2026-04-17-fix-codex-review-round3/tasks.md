# Tasks: fix-codex-review-round3

- [x] Fix 1 (P1): Add SaveToolResult method to knowledge.Store
- [x] Fix 1 (P1): Wire iv.KC.store as KnowledgeSaver in app.go line 180
- [x] Fix 2 (P2): Resolve symlinks via EvalSymlinks before copyFile in copyTree
- [x] Fix 2 (P2): Add TestCopyTreeAcceptsInRootSymlink test
- [x] Fix 3 (P2): Add cliboot.Version to settings.go bootstrap.Run
- [x] Fix 3 (P2): Add cliboot.Version to doctor.go bootstrap.Run
- [x] Fix 3 (P2): Add cliboot.Version to onboard.go bootstrap.Run
- [x] Verify: go build ./... — PASS
- [x] Verify: go test ./... — ALL PASS
