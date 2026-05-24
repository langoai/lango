## 1. Post-Turn Default Starter

- [x] 1.1 Make the empty ready-profile workbench use the next-step starter as the default `Enter` seed after a completed turn.
- [x] 1.2 Add regressions for the post-turn empty-state copy and `Enter` behavior.
- [x] 1.3 Update public workbench docs for the new post-turn default starter behavior.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/workbench -count=1`.
- [ ] 2.2 Run `go build ./...`.
- [ ] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `mission-workbench-tui` and `downstream-docs-sync` coverage for the post-turn default starter.
- [ ] 3.2 Validate and archive the OpenSpec change.
