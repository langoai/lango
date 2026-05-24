## 1. Tests

- [x] 1.1 Add failing CLI tests for workspace create/list/status local persistence.
- [x] 1.2 Add failing CLI tests for workspace join/leave membership mutation.
- [x] 1.3 Add failing docs guard coverage for workspace quick-reference action wording.

## 2. Implementation

- [x] 2.1 Add a local workspace CLI manager opener with data-dir resolution and feature gates.
- [x] 2.2 Wire `workspace create/list/status` to the local manager with table and JSON output.
- [x] 2.3 Wire `workspace join/leave` to local membership mutation.
- [x] 2.4 Keep `p2p git` guidance behavior unchanged.

## 3. Documentation and Specs

- [x] 3.1 Update README and `docs/cli/index.md` workspace quick-reference descriptions.
- [x] 3.2 Update `docs/cli/p2p.md` workspace command reference.
- [x] 3.3 Sync main OpenSpec specs with the local CLI workspace behavior.

## 4. Verification

- [x] 4.1 Validate the OpenSpec change in strict mode.
- [x] 4.2 Run focused P2P workspace CLI and docs guard tests.
- [x] 4.3 Run `go build ./...` and `go test ./...`.
- [x] 4.4 Run subagent-driven review.
- [x] 4.5 Archive the OpenSpec change.
- [x] 4.6 Commit this scoped unit separately.
