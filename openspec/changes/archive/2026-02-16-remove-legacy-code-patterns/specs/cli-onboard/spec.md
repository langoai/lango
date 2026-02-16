## REMOVED Requirements

### Requirement: SystemPromptPath onboard form field
**Reason**: The `SystemPromptPath` field and its "legacy" label in the Agent configuration form are removed. The `PromptsDir` directory-based approach is the only supported method.
**Migration**: Use `agent.promptsDir` to specify a directory of .md prompt files.

### Requirement: Embedding provider backward-compatibility auto-resolve
**Reason**: The onboard state update no longer auto-resolves `Provider` type from `ProviderID` for backward compatibility. When a non-local provider is selected, only `ProviderID` is set; `Provider` is cleared.
**Migration**: No action needed; the onboard TUI handles this automatically.
