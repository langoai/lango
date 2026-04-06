## MODIFIED Requirements

### Requirement: Channel status section in context panel
The context panel SHALL display a "Channels" section showing each channel's connection status (connected/disconnected indicator), name, and message count.

#### Scenario: Connected channel displayed
- **WHEN** a channel with Connected=true and MessageCount=5 is set
- **THEN** the context panel renders a green "●" indicator, the channel name, and "5 msgs"

#### Scenario: Disconnected channel displayed
- **WHEN** a channel with Connected=false is set
- **THEN** the context panel renders a red "○" indicator

#### Scenario: No channels configured
- **WHEN** no channel statuses are set
- **THEN** the "Channels" section is not rendered (graceful degradation)

#### Scenario: Channel statuses updated on tick
- **WHEN** the context panel tick fires and a ChannelTracker is available
- **THEN** the cockpit calls `tracker.Snapshot()` and pushes results to `SetChannelStatuses`
