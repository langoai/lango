## Overview

The doctor JSON renderer already returns a string, so the safest improvement is to use an encoder over a local buffer and then normalize the trailing newline back to the historical shape.

## Decisions

### Use `json.Encoder` with indentation

The renderer now writes into a `bytes.Buffer` using `json.NewEncoder(...)` with indentation enabled, then trims the encoder's trailing newline.

### Keep the historical string shape

Consumers continue to receive pretty-printed JSON without an extra newline suffix, so no downstream formatting changes are required.

## Non-Goals

- No change to doctor result fields
- No change to command-level writer routing
