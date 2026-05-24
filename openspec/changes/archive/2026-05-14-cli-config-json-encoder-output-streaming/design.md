## Overview

The config CLI already routes output through command writers. This change tightens the JSON path by writing encoded JSON directly to those writers instead of converting JSON bytes into intermediate strings first.

## Decisions

### Use `json.Encoder` for config JSON paths

`config export` and `printValue(..., "json")` now use `json.NewEncoder(...)` with indentation enabled. This preserves pretty-printed output while avoiding extra string conversion.

### Validate JSON output by decoding captured command output

Tests now parse the captured JSON output to ensure the contract is stronger than simple substring checks.

## Non-Goals

- No change to config key semantics
- No change to warning routing on stderr for plaintext export
