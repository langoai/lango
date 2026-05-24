## Why

The `lango p2p git` guidance commands still tell operators to use a generic "runtime API" for `init`, `log`, `diff`, and `push`, but the public P2P gateway surface only exposes read-only identity/reputation/pricing endpoints plus provenance push/fetch. There is no public workspace/git control API matching that guidance. The docs and CLI should point users to the server-backed runtime and the actual `p2p_git_*` tools instead of inventing a nonexistent public API path.

## What Changes

- Replace stale "runtime API" wording in `lango p2p git init/log/diff/push` with truthful server-backed runtime guidance.
- Add CLI regressions locking the new guidance strings.
- Sync the public P2P CLI docs and CLI P2P management spec to the same contract.

## Impact

- `cli-p2p-management`: git guidance becomes consistent with the actual gateway/runtime surface.
- Operator UX: users are no longer sent toward a nonexistent public git/workspace API.
