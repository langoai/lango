## Why

The `p2p-skills` main spec still claimed that a bundle of embedded P2P script skills shipped in the repository, but the current runtime embeds only the placeholder skill scaffold. That mismatch makes the spec materially false and can mislead maintainers about what the binary actually contains.

## What Changes

- sync the `p2p-skills` main spec to the current placeholder-only embedded state
- add an executable guard so stale embedded-P2P-skill claims cannot silently return

## Impact

- main specs better reflect the actual shipped runtime
- reduced confusion around embedded skills versus imported/user-created skills
- stronger regression protection for future spec drift
