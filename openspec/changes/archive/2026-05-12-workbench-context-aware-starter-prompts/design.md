## Overview

This change keeps the workbench startup path lightweight while letting it acknowledge the operator's current workspace.

## Design Decisions

### Context detection stays local and dependency-free

Starter prompt context is derived from:

- the configured exec workdir when present
- otherwise the process current directory
- upward marker discovery for `.git` and `go.mod`

This avoids shelling out to `git` or adding a heavy repository dependency just to improve first-screen guidance.

### Prompts remain stable in count and hotkey mapping

The workbench still exposes exactly three starter prompts mapped to `1`, `2`, and `3`. Only the text adapts based on workspace context, preserving the quick-start interaction model already taught in the UI.

### Repository awareness is additive

If no repository markers are found, the workbench falls back to the existing generic prompts. Context detection improves the startup path when possible but never makes bare `lango` less usable outside a repo.
