## MODIFIED Requirements

### Requirement: Tool catalog browser with categories
ToolsPage SHALL display categories from `ToolCatalog.ListCategories()` with cursor navigation. The currently highlighted category SHALL immediately drive the tool details shown in the right panel.

#### Scenario: Rendered tools-page text stays plain and single-line
- **WHEN** category names, category descriptions, tool names, tool descriptions, or safety labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Tools page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
- **AND** known safety-level colors SHALL still be selected from the sanitized safety label
