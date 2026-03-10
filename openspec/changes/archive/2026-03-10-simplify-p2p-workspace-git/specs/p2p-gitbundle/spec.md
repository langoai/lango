## MODIFIED Requirements

### Service
Git bundle operations wrapping BareRepoStore.
- CreateBundle: `git bundle create --all` via CLI
- ApplyBundle: `git bundle unbundle` via CLI (single step, no redundant fetch)
- Log: Commit listing across all refs via go-git, uses sentinel error for limit control
- Diff: `git diff` between two commits via CLI
- Leaves: DAG leaf detection (commits with no children)

### Protocol
libp2p stream handler for `/lango/p2p-git/1.0.0`.
- Request types: push_bundle, fetch_by_hash, list_commits, find_leaves, diff
- Session-based authentication via SessionValidator callback
- 50MB default bundle size limit
- 5-minute request timeout
- Streaming JSON decoder for memory-efficient request parsing
- Response status uses `StatusOK`/`StatusError` constants

## ADDED Requirements

### Requirement: Sentinel error for commit log limit
`Log()` SHALL use a sentinel error `errLimitReached` and `errors.Is()` for limit control flow instead of string comparison.

#### Scenario: Log with limit
- **WHEN** `Log()` reaches the specified commit limit
- **THEN** iteration SHALL stop using `errLimitReached` sentinel error
- **THEN** the sentinel SHALL be checked via `errors.Is()`, not string comparison

### Requirement: Streaming protocol request decoding
The git protocol handler SHALL use `json.NewDecoder().Decode()` instead of `io.ReadAll()` + `json.Unmarshal()` to avoid double-buffering large requests.

#### Scenario: Large bundle push
- **WHEN** a push_bundle request is received with a large bundle
- **THEN** the request SHALL be decoded via streaming decoder without allocating an intermediate buffer

### Requirement: Response status constants
Response status SHALL use named constants `StatusOK` and `StatusError` instead of string literals.

#### Scenario: Success response
- **WHEN** a git protocol request succeeds
- **THEN** response status SHALL be set to `StatusOK`

#### Scenario: Error response
- **WHEN** a git protocol request fails
- **THEN** response status SHALL be set to `StatusError`

### Requirement: No redundant subprocess in ApplyBundle
`ApplyBundle` SHALL run only `git bundle unbundle` without a subsequent dead `git fetch` subprocess.

#### Scenario: Apply bundle
- **WHEN** a bundle is applied to a workspace
- **THEN** only `git bundle unbundle` SHALL be executed
