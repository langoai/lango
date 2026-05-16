## MODIFIED Requirements
### Requirement: Page interface with lifecycle
The cockpit SHALL define a `Page` interface extending `tea.Model` with `Title() string`, `ShortHelp() []key.Binding`, `Activate() tea.Cmd`, and `Deactivate()`.

#### Scenario: Sessions and Tools vertical navigation hints use cockpit-standard labels
- **WHEN** the Sessions or Tools page help is rendered
- **THEN** the vertical navigation bindings SHALL use `↑/k` and `↓/j` help labels rather than textual `up/k` or `down/j` forms
