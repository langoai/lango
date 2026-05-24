## ADDED Requirements

### Requirement: Extension confirmation uses shared guarded prompt helper
`lango extension install` and `lango extension remove` SHALL use the shared terminal-guarded confirmation helper for their interactive confirmation flow.

#### Scenario: Extension confirmation refuses non-TTY stdin through shared helper
- **WHEN** `lango extension install <pack>` or `lango extension remove <name>` runs without `--yes` and stdin is a non-terminal `*os.File`
- **THEN** the command SHALL fail with the existing scripted-run guidance
- **AND** the prompt SHALL be mediated by the shared guarded confirmation helper
