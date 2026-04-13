---
title: Provenance CLI
---

# lango provenance

Inspect checkpoints, session lineage, attribution, and signed provenance bundles.

## Synopsis

```bash
lango provenance <command>
```

## Commands

```bash
lango provenance status
lango provenance checkpoint list --run <id>
lango provenance checkpoint create <label> --run <id>
lango provenance checkpoint show <id>
lango provenance session tree <session-key> --depth 10
lango provenance session list --limit 50 --status active
lango provenance attribution show <session-key>
lango provenance attribution report <session-key>
lango provenance bundle export <session-key> --redaction content --out bundle.json
lango provenance bundle import bundle.json
```

## Bundle Semantics

- Export signs the canonical provenance payload with the local wallet identity and embeds the signer DID.
- Import verifies the signer DID and signature before storing the bundle contents.
- Redaction levels:
  - `none`: include full provenance data
  - `content`: strip content-bearing fields such as goals, git refs, and raw file paths
  - `full`: keep only the aggregate report envelope
