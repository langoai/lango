---
name: p2p-pricing
description: Show P2P tool pricing configuration
version: 1.0.0
type: script
status: active
---

## Usage

```bash
lango p2p pricing
```

## Description

Display the current P2P pricing configuration including whether paid invocations are enabled, the default per-query price, and tool-specific price overrides.

## Examples

```bash
# Show all pricing
lango p2p pricing

# Show pricing as JSON
lango p2p pricing --json

# Show pricing for a specific tool
lango p2p pricing --tool "knowledge_search"
```
