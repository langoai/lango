## Overview

The workflow run command has three major non-error branches after validation: schedule-not-implemented, runtime unavailable, and workflow engine disabled. Only the first one was covered. This change adds command-level tests for the other two branches and documents the fallback guidance contract.

## Decisions

### Reuse the existing workflow fixture helper

The tests use the same YAML fixture pattern as the existing schedule-path regression to keep the coverage focused on command behavior rather than parser setup.

### Document fallback guidance as part of the workflow run contract

The operator-facing messages for unavailable runtime and disabled engine are treated as supported behavior and captured in docs/specs.

## Non-Goals

- No change to workflow run semantics
- No implementation of direct execution fallback beyond current guidance
