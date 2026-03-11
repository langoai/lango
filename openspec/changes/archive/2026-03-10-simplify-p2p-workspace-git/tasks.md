## 1. Concurrency & Safety

- [x] 1.1 Fix `Node.PubSub()` data race — use `sync.Once` for lazy GossipSub init
- [x] 1.2 Eliminate double `WorkspaceGossip` construction in `wiring_workspace.go`

## 2. Dead Code Removal

- [x] 2.1 Remove dead chronicler wiring in `app.go` (allocate+convert+discard no-op)
- [x] 2.2 Remove dead `git fetch` subprocess in `ApplyBundle`
- [x] 2.3 Remove redundant `workspaceComponents` type from `tools_workspace.go`
- [x] 2.4 Remove redundant `maxWS` default in `wiring_workspace.go` (Manager already defaults)
- [x] 2.5 Remove unused `workspace` import from `app.go`
- [x] 2.6 Remove unused `gitbundle` import from `tools_workspace.go`

## 3. Type Safety

- [x] 3.1 Add `Role` type with `RoleCreator`/`RoleMember` constants in `workspace.go`
- [x] 3.2 Update `manager.go` to use typed `Role` constants
- [x] 3.3 Add `StatusOK`/`StatusError` constants in `messages.go`
- [x] 3.4 Update `protocol.go` to use status constants
- [x] 3.5 Add `errLimitReached` sentinel error in `bundle.go`
- [x] 3.6 Replace string comparison with `errors.Is(err, errLimitReached)`

## 4. Efficiency

- [x] 4.1 Replace `io.ReadAll` + `json.Unmarshal` with `json.NewDecoder().Decode()` in protocol handler
- [x] 4.2 Replace manual byte-by-byte prefix check with `bytes.HasPrefix` in `manager.go`

## 5. Code Reuse

- [x] 5.1 Extract `errP2PDisabled` sentinel in `cli/p2p/p2p.go`
- [x] 5.2 Replace 5 error string literals in `cli/p2p/git.go` with sentinel
- [x] 5.3 Replace 5 error string literals in `cli/p2p/workspace.go` with sentinel
- [x] 5.4 Update `initP2PDeps` to use sentinel

## 6. Cleanup API

- [x] 6.1 Add `ContributionTracker.Remove(workspaceID)` method

## 7. Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 All tests pass
- [x] 7.3 `go vet` clean
