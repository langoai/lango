## Why

`lango config export` and `config get --format json` currently materialize JSON with `MarshalIndent` and then print the resulting string. The behavior works, but it duplicates buffering and leaves the JSON output contract less explicit than other command-writer paths.

## What Changes

- Encode config export JSON directly to the command writer with `json.Encoder`
- Encode `config get --format json` output directly to the target writer with `json.Encoder`
- Strengthen regression coverage by decoding the captured JSON output
- Update docs and OpenSpec to make the decodable command-writer contract explicit

## Impact

- Keeps config JSON output behavior stable while reducing unnecessary buffering
- Improves confidence that wrappers can consume the JSON output directly
