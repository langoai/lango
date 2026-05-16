## Why

The public P2P feature page already needed a truthfulness fix for `lango p2p identity`, and its CLI command summary still described the command as showing only peer identity and listen addresses. The current CLI prints the active DID when available, so the summary line should match that behavior too.

## What Changes

- sync the P2P feature-page command summary to the current `lango p2p identity` output contract
- extend the existing P2P identity docs guard so the stale summary wording cannot silently return

## Impact

- public P2P feature docs stay internally consistent
- reduced operator confusion when scanning the command summary
- stronger regression protection for identity output documentation
