---
name: p2p-reputation
description: Show peer reputation and trust score details
version: 1.0.0
type: script
status: active
---

## Usage

```bash
lango p2p reputation --peer-did "$PEER_DID"
```

## Description

Query the reputation system for a specific peer's trust score, exchange history, and interaction timeline. Shows successful exchanges, failed exchanges, timeout count, and trust score.

## Arguments

- `PEER_DID` — The DID of the peer to query (e.g. `did:lango:abc123...`)

## Examples

```bash
# Show reputation details for a peer
lango p2p reputation --peer-did "did:lango:abc123"

# Output as JSON
lango p2p reputation --peer-did "did:lango:abc123" --json
```
