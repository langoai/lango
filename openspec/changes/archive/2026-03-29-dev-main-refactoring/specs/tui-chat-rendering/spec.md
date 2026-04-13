## MODIFIED Requirements

### Requirement: Markdown rendering performance
The chat TUI SHALL cache the glamour `TermRenderer` at module level, keyed by terminal width. The renderer SHALL be reused across `renderMarkdown()` calls at the same width. A new renderer SHALL only be created when width changes.

#### Scenario: Renderer reused on cursor tick
- **WHEN** `renderMarkdown` is called multiple times at the same width (e.g., cursor blink every 400ms)
- **THEN** the same cached renderer SHALL be reused without creating a new one

#### Scenario: Renderer recreated on width change
- **WHEN** the terminal width changes
- **THEN** a new renderer SHALL be created and cached for the new width

### Requirement: Transcript render optimization
The chat `render()` method SHALL use the pre-rendered `content` field for finalized assistant entries. It SHALL NOT re-invoke `renderMarkdown()` on every render pass. Re-rendering of assistant entries SHALL only occur in `setSize()` when the width actually changes.

#### Scenario: Cursor tick does not re-render finalized entries
- **WHEN** a cursor blink tick fires
- **THEN** `render()` SHALL use cached `entry.content` for all finalized assistant entries

#### Scenario: Width change triggers assistant re-render
- **WHEN** `setSize()` is called and width differs from previous
- **THEN** all assistant entries with `rawContent` SHALL have their `content` field re-rendered

#### Scenario: Height-only change skips re-render
- **WHEN** `setSize()` is called with the same width but different height
- **THEN** assistant entries SHALL NOT be re-rendered
