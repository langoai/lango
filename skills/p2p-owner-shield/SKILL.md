---
name: p2p-owner-shield
description: Show Owner Shield protection status
version: 1.0.0
type: script
status: active
---

## Usage

```bash
lango p2p status --json | jq '.ownerShield'
```

## Description

Check the Owner Shield configuration that prevents owner PII (name, email, phone) from leaking through P2P responses. The Owner Shield sanitizes outgoing P2P responses to remove personally identifiable information.

## Examples

```bash
# Check owner shield status via P2P status
lango p2p status --json | jq '.ownerShield'

# Full P2P status including owner shield
lango p2p status
```
