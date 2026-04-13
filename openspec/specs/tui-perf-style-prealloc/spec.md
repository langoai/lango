## Purpose

Capability spec for tui-perf-style-prealloc. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Module-level style pre-allocation for render functions
All render helper functions in the chat package SHALL use module-level pre-allocated lipgloss.Style variables for fixed style properties (Bold, Border, PaddingLeft, etc.). Per-call style chains SHALL only set variable properties (Foreground color).

#### Scenario: renderTranscriptBlock uses pre-allocated styles
- **WHEN** renderTranscriptBlock is called
- **THEN** it SHALL use module-level transcriptLabelStyle and transcriptBorderStyle instead of lipgloss.NewStyle()

#### Scenario: render_tool.go uses pre-allocated styles
- **WHEN** renderToolBlock is called
- **THEN** it SHALL use module-level toolLabelStyle, toolDetailStyle, and toolOutputStyle

#### Scenario: render_thinking.go uses pre-allocated styles
- **WHEN** renderThinkingBlock or renderPendingIndicator is called
- **THEN** they SHALL use module-level pre-allocated styles

#### Scenario: render_channel.go uses pre-allocated styles
- **WHEN** renderChannelBlock is called
- **THEN** it SHALL use module-level channelBadgeStyle, channelSenderStyle, channelTextStyle

#### Scenario: render_delegation.go and render_recovery.go use pre-allocated styles
- **WHEN** renderDelegationBlock or renderRecoveryBlock is called
- **THEN** they SHALL use module-level pre-allocated styles

### Requirement: Terminal task FIFO cap
The BackgroundManager SHALL enforce a maximum of 500 terminal (Done/Failed/Cancelled) tasks. When exceeded, the oldest terminal task (by CompletedAt, then StartedAt) SHALL be evicted from the map.

#### Scenario: Oldest terminal task evicted on overflow
- **WHEN** a task transitions to terminal state and terminal count exceeds 500
- **THEN** the oldest terminal task SHALL be removed from the tasks map

#### Scenario: Active tasks never evicted
- **WHEN** eviction runs
- **THEN** only terminal tasks SHALL be candidates for eviction; Running/Pending tasks SHALL be preserved

### Requirement: Grant store lazy cleanup
The GrantStore SHALL automatically clean expired grants during List() calls via a cleanExpiredLocked() internal helper under a write lock.

#### Scenario: Expired grants removed on List
- **WHEN** List() is called and expired grants exist
- **THEN** expired grants SHALL be removed from the map before listing

#### Scenario: List uses write lock for cleanup
- **WHEN** List() executes
- **THEN** it SHALL acquire a write Lock (not RLock) to enable cleanup
