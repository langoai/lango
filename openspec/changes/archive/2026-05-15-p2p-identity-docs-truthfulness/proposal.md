## Why

The public P2P feature docs still said that `lango p2p identity` focused on peer identity and listen addresses rather than printing the DID directly, but the current CLI already prints the active DID when available. That mismatch makes the documentation less trustworthy for operators.

## What Changes

- sync the P2P feature docs to the current `lango p2p identity` output contract
- add an executable docs guard so the stale wording cannot silently return

## Impact

- public docs better match the actual CLI experience
- reduced confusion for operators checking P2P identity behavior
- stronger regression protection for feature-doc drift
