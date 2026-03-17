## Why

Code review of the P2P workspace & git integration revealed concurrency bugs, dead code, redundant types, stringly-typed fields, and inefficient patterns that need cleanup before the feature stabilizes.

## What Changes

- Fix `Node.PubSub()` data race with `sync.Once` (was creating duplicate GossipSub instances)
- Remove dead chronicler wiring in `app.go` (allocated/converted triples then discarded them)
- Eliminate double `WorkspaceGossip` construction in `wiring_workspace.go`
- Replace manual byte-by-byte prefix check with `bytes.HasPrefix` in `manager.go`
- Use `errLimitReached` sentinel error instead of string comparison in `bundle.go`
- Remove dead `git fetch` subprocess in `ApplyBundle`
- Switch to streaming JSON decoder in git protocol handler (halves memory for large bundles)
- Add `StatusOK`/`StatusError` constants for git protocol responses
- Add `Role` type with `RoleCreator`/`RoleMember` constants (was raw strings)
- Merge redundant `workspaceComponents` struct into `wsComponents`
- Add `ContributionTracker.Remove()` for workspace cleanup
- Extract `errP2PDisabled` sentinel to eliminate 10+ duplicate error strings in CLI

## Capabilities

### New Capabilities

(none)

### Modified Capabilities
- `p2p-workspace`: Fix concurrency bug in PubSub init, eliminate dead code, add typed Role, add contribution cleanup
- `p2p-gitbundle`: Fix error handling, remove dead subprocess, improve protocol efficiency

## Impact

- `internal/p2p/node.go` — sync.Once for PubSub
- `internal/p2p/gitbundle/bundle.go` — sentinel error, dead code removal
- `internal/p2p/gitbundle/protocol.go` — streaming decoder
- `internal/p2p/gitbundle/messages.go` — status constants
- `internal/p2p/workspace/workspace.go` — Role type
- `internal/p2p/workspace/manager.go` — bytes.HasPrefix, typed roles
- `internal/p2p/workspace/contribution.go` — Remove method
- `internal/app/app.go` — dead chronicler removal, import cleanup
- `internal/app/tools_workspace.go` — redundant type removal
- `internal/app/wiring_workspace.go` — single gossip construction, remove redundant default
- `internal/cli/p2p/p2p.go`, `git.go`, `workspace.go` — sentinel error
