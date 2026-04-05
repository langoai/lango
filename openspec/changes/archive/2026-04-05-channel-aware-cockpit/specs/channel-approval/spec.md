## MODIFIED Requirements

### Requirement: Channel origin display on approval requests
Approval rendering surfaces SHALL display channel origin information when the approval request's SessionKey contains a recognized channel prefix.

#### Scenario: Telegram origin on approval banner
- **WHEN** an approval request has SessionKey="telegram:123:456"
- **THEN** the approval banner renders an origin line containing "[Telegram]"

#### Scenario: Channel badge on approval strip
- **WHEN** a Tier 1 approval has SessionKey="telegram:123:456"
- **THEN** the approval strip summary is prefixed with "[TG]"

#### Scenario: Channel origin on approval dialog
- **WHEN** a Tier 2 approval has SessionKey="discord:ch1:user1"
- **THEN** the approval dialog header includes an origin line containing "[Discord]"

#### Scenario: No origin for local session
- **WHEN** an approval request has SessionKey="tui-12345"
- **THEN** no channel origin info is displayed
