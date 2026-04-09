## Purpose

Capability spec for p2p-gitbundle. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: p2p-gitbundle capability documented
The p2p-gitbundle capability SHALL be documented through the sections in this spec. This requirement is a structural placeholder that satisfies the canonical openspec format; detailed behavior contracts live in the descriptive sections of this file.

#### Scenario: Spec file is readable
- **WHEN** the p2p-gitbundle spec.md file is read
- **THEN** it SHALL describe the capability's behavior in sections below

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
- ApplyBundle: `git bundle unbundle` via CLI (single step, no redundant fetch)
- Log: Commit listing across all refs via go-git, uses sentinel error for limit control
- Diff: `git diff` between two commits via CLI
- Leaves: DAG leaf detection (commits with no children)

### Protocol
libp2p stream handler for `/lango/p2p-git/1.0.0`.
- Request types: push_bundle, fetch_by_hash, list_commits, find_leaves, diff, push_incremental_bundle, fetch_incremental, verify_bundle, has_commit
- Session-based authentication via SessionValidator callback
- 50MB default bundle size limit
- 5-minute request timeout
- Streaming JSON decoder for memory-efficient request parsing
- Response status uses `StatusOK`/`StatusError` constants

## Agent Tools

- p2p_git_init, push, log, diff, leaves

## CLI Commands

- `lango p2p git init/log/diff/push/fetch`

## Design Decisions

- Git bundle approach (not smart protocol) for simplicity
- Sprawling DAG model — no branches, navigate by commit hash
- go-git for programmatic access, git CLI for bundle operations
- Base64 JSON transport consistent with A2A protocol

## Incremental Bundle Protocol

### Requirement: Incremental bundle protocol messages
The git bundle protocol SHALL support four new request types: push_incremental_bundle, fetch_incremental, verify_bundle, and has_commit.

#### Scenario: Push incremental bundle
- **WHEN** a push_incremental_bundle request is received with a valid bundle
- **THEN** the handler calls SafeApplyBundle and returns PushBundleResponse with Applied=true

#### Scenario: Fetch incremental bundle
- **WHEN** a fetch_incremental request is received with a base commit hash
- **THEN** the handler calls CreateIncrementalBundle and returns FetchIncrementalResponse with the bundle and HEAD hash

#### Scenario: Verify bundle
- **WHEN** a verify_bundle request is received
- **THEN** the handler calls VerifyBundle and returns VerifyBundleResponse with Valid=true or Valid=false with message

#### Scenario: Has commit check
- **WHEN** a has_commit request is received
- **THEN** the handler calls HasCommit and returns HasCommitResponse with Exists boolean
