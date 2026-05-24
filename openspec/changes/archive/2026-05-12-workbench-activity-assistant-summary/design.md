## Overview

This change makes the workbench result loop more visible without adding a full transcript pane.

## Design Decisions

### Activity keeps short assistant summaries

The workbench now appends a short assistant summary after `DoneMsg` using the existing `turnrunner.Result` summary fields. Successes are labeled as assistant replies; non-success outcomes carry an outcome-prefixed summary.

### Stay lightweight

This keeps the workbench result trail lightweight and mission-first: enough to show what happened, without trying to replicate the full chat transcript inside Mission Control.
