## Overview

This change keeps the `Enter` quick-start path simple while making its default choice smarter.

## Design Decisions

### Dirty repos prefer the review prompt

When the shared prompt helper detects a dirty Git workspace, `Enter` now seeds the changed-review prompt instead of the summary prompt. That reduces the chance that the default quick-start path ignores the operator's actual active work.

### Clean and non-repo flows stay conservative

For clean repositories and non-repository workspaces, `Enter` still seeds the summary-oriented default. This avoids overfitting the quick-start behavior to Git-only workflows.
