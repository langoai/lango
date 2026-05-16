## Why

The workbench Mission Control empty state can replay the latest assistant activity summary as `Last result:`. That path currently strips only a fixed prefix and does not sanitize the summary, so malformed assistant activity text can still leak control sequences or embedded newlines into the default empty-state surface.

## What Changes

- Sanitize the workbench empty-state `Last result:` summary before rendering it.
- Add regression coverage for malformed assistant activity summaries in the empty state.
- Record the plain-text empty-state result contract in OpenSpec and downstream docs.

## Impact

- Closes the remaining raw assistant-summary replay path inside Mission Control after lane/header hardening.
- Prevents malformed assistant activity text from destabilizing the empty-state workbench surface.
