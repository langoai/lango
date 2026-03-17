# Tasks: P2P Workspace & Git Bundle Integration

## Wave 1: Foundation (no dependencies)

- [x] **WU-1**: Add `WorkspaceConfig` struct to `internal/config/types_p2p.go`
  - Added 7-field config struct and `Workspace` field to `P2PConfig`

- [x] **WU-2**: Create `internal/eventbus/workspace_events.go`
  - 6 event types: Created, MemberJoined, MemberLeft, CommitReceived, MessagePosted, Archived

- [x] **WU-6**: Create `internal/p2p/gitbundle/store.go`
  - `BareRepoStore` with Init, Repo, RepoPath, List, Remove methods
  - Added `go-git/go-git/v5` dependency

## Wave 2: Core packages

- [x] **WU-3**: PubSub sharing refactor
  - Added `ps *pubsub.PubSub` field and `PubSub()` method to `p2p.Node`
  - Added optional `PubSub` field to `discovery.GossipConfig`
  - Backward-compatible: nil PubSub creates a new instance

- [x] **WU-4**: Create workspace core package
  - `internal/p2p/workspace/workspace.go` — types (Workspace, Member, Status, CreateRequest, ReadOptions)
  - `internal/p2p/workspace/message.go` — Message, MessageType (6 types)
  - `internal/p2p/workspace/manager.go` — BoltDB-backed Manager with full CRUD + messaging

- [x] **WU-7**: Create `internal/p2p/gitbundle/bundle.go`
  - `Service` with Init, CreateBundle, ApplyBundle, Log, Diff, Leaves
  - Uses git CLI for bundle ops, go-git for programmatic access

- [x] **WU-8**: Create git protocol messages + handler
  - `internal/p2p/gitbundle/messages.go` — Protocol ID, 5 request types, payload structs
  - `internal/p2p/gitbundle/protocol.go` — Stream handler with session validation

## Wave 3: Integration layer

- [x] **WU-5**: Create `internal/p2p/workspace/gossip.go`
  - `WorkspaceGossip` with Subscribe, Unsubscribe, Publish, Stop
  - Per-workspace GossipSub topics via shared PubSub instance

- [x] **WU-9**: Create `internal/app/tools_workspace.go`
  - 7 workspace tools + 5 git tools (12 total)
  - `workspaceComponents` struct + `buildWorkspaceTools` + `buildGitTools`

- [x] **WU-10**: Create CLI commands
  - `internal/cli/p2p/workspace.go` — 5 commands (create, list, status, join, leave)
  - `internal/cli/p2p/git.go` — 5 commands (init, log, diff, push, fetch)
  - Modified `internal/cli/p2p/p2p.go` to register workspace + git subcommands

- [x] **WU-12**: Create chronicler + contribution tracking
  - `internal/p2p/workspace/chronicler.go` — Message → graph triple conversion
  - `internal/p2p/workspace/contribution.go` — Per-member contribution tracking

## Wave 4: Final integration

- [x] **WU-11**: App wiring
  - Created `internal/app/wiring_workspace.go` — `initWorkspace()` function
  - Modified `internal/app/app.go` — workspace initialization, tool registration, lifecycle
  - Updated `internal/cli/p2p/p2p_test.go` — subcommand count 13 → 15

## Verification

- [x] `go build ./...` — passes
- [x] `go test ./...` — 0 failures
- [x] `go vet ./...` — clean
