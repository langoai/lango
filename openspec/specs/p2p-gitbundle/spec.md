# P2P Git Bundle

## Overview

P2P git bundle exchange for workspace code sharing. Enables agents to share code changes without a central git server using the git bundle format.

## Package

`internal/p2p/gitbundle/`

## Components

### BareRepoStore
Manages bare git repositories per workspace under `~/.lango/workspaces/{id}/repo.git`.
- Init, Repo, RepoPath, List, Remove
- Thread-safe with RWMutex-protected cache
- Dependency: go-git/go-git/v5

### Service
Git bundle operations wrapping BareRepoStore.
- CreateBundle: `git bundle create --all` via CLI
- ApplyBundle: `git bundle unbundle` + fetch via CLI
- Log: Commit listing across all refs via go-git
- Diff: `git diff` between two commits via CLI
- Leaves: DAG leaf detection (commits with no children)

### Protocol
libp2p stream handler for `/lango/p2p-git/1.0.0`.
- Request types: push_bundle, fetch_by_hash, list_commits, find_leaves, diff
- Session-based authentication via SessionValidator callback
- 50MB default bundle size limit
- 5-minute request timeout
- JSON protocol with base64-encoded bundle data

## Agent Tools

- p2p_git_init, push, log, diff, leaves

## CLI Commands

- `lango p2p git init/log/diff/push/fetch`

## Design Decisions

- Git bundle approach (not smart protocol) for simplicity
- Sprawling DAG model — no branches, navigate by commit hash
- go-git for programmatic access, git CLI for bundle operations
- Base64 JSON transport consistent with A2A protocol
