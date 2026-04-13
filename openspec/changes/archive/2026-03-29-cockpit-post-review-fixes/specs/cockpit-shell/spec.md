## MODIFIED Requirements

### Requirement: Context panel toggle with synthetic resize
When Ctrl+P toggles the context panel, cockpit SHALL send synthetic WindowSizeMsg to all components with updated effective widths. Additionally, the context panel itself SHALL receive its correct width on toggle-on, even if it previously received width=0 while hidden.

#### Scenario: First toggle after hidden initial state
- **WHEN** context panel was hidden during initial WindowSizeMsg (received width=0) and user presses Ctrl+P
- **THEN** the context panel SHALL receive WindowSizeMsg with Width=ContextPanelWidth before rendering
