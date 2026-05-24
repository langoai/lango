## Overview

`lango contract read` is a local ABI-validation command that emits two kinds of output:

- a validation payload intended for stdout
- an informational runtime note intended for stderr

The implementation change is only about routing those payloads through Cobra-managed streams.

## Decisions

### Split stdout and stderr intentionally

- Validation summary or JSON payload uses `cmd.OutOrStdout()`
- Informational runtime note uses `cmd.ErrOrStderr()`

This preserves the semantic separation between operator payload and informational guidance.

### Cover both text and JSON paths with split capture tests

Tests use temporary ABI fixtures and assert:

- stdout contains the validated payload
- stderr contains the runtime note

## Non-Goals

- No change to ABI parsing or method validation behavior
- No change to the note wording
