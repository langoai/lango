## Overview

The command already implements cleanup by removing the just-created file when the exporter fails. This change only makes that behavior executable and explicit.

## Decision

- Reuse a stub exporter that writes partial output and then returns an error
- Verify both the direct single-circuit path and the `--all` orchestration path

## Consequences

- Partial verifier artifacts can no longer silently regress into the filesystem on failed export paths
