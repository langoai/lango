## Why

The doctor JSON renderer currently uses `json.MarshalIndent` and converts the resulting bytes to a string. The behavior works, but it leaves the decodable pretty-JSON contract less explicit than other output paths we have recently tightened.

## What Changes

- Encode doctor JSON output through a `json.Encoder` backed by a buffer
- Preserve the existing no-trailing-newline string behavior
- Strengthen regression coverage by decoding the rendered JSON output
- Update docs and OpenSpec to state the decode-safe contract explicitly

## Impact

- Keeps doctor JSON behavior stable while tightening the rendering path
- Improves confidence for wrappers that consume `lango doctor --json`
