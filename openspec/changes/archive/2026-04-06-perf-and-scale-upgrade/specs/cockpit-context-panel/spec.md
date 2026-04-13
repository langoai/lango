## MODIFIED Requirements

### Requirement: Context panel optimized snapshot handling
The ContextPanel SHALL reuse existing slice capacity in SetChannelStatuses() instead of allocating a new slice on every call. Style variables for render methods SHALL be pre-allocated at module level. The toolCountSum SHALL be cached alongside the sortedTools dirty flag.

#### Scenario: SetChannelStatuses reuses slice capacity
- **WHEN** SetChannelStatuses is called with a status list of equal or smaller length than existing capacity
- **THEN** the existing slice SHALL be resliced and copied without new allocation

#### Scenario: Render styles pre-allocated
- **WHEN** renderRuntimeStatus or renderChannelStatus renders status items
- **THEN** they SHALL use module-level pre-allocated style variables instead of inline lipgloss.NewStyle()

#### Scenario: Tool count sum cached with dirty flag
- **WHEN** the sortedTools dirty flag is false and toolCountSum is needed
- **THEN** the cached sum SHALL be returned without iterating the tool breakdown map
