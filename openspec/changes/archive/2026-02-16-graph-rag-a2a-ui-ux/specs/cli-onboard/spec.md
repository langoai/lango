## ADDED Requirements

### Requirement: Graph store wizard screen
The onboard wizard SHALL include a "Graph Store" menu item that opens a form with fields: graph_enabled (bool), graph_backend (select: bolt), graph_db_path (text), graph_max_depth (int), graph_max_expand (int). Form values SHALL be written back to the config.

#### Scenario: Configure graph via wizard
- **WHEN** user selects "Graph Store" from onboard menu and fills the form
- **THEN** config.Graph fields are updated with form values

### Requirement: Multi-agent wizard screen
The onboard wizard SHALL include a "Multi-Agent" menu item that opens a form with a single multi_agent (bool) toggle. Form values SHALL be written back to config.Agent.MultiAgent.

#### Scenario: Enable multi-agent via wizard
- **WHEN** user selects "Multi-Agent" and toggles enabled
- **THEN** config.Agent.MultiAgent is set to true

### Requirement: A2A protocol wizard screen
The onboard wizard SHALL include an "A2A Protocol" menu item that opens a form with fields: a2a_enabled (bool), a2a_base_url (text), a2a_agent_name (text), a2a_agent_desc (text). Form values SHALL be written back to the config.

#### Scenario: Configure A2A via wizard
- **WHEN** user selects "A2A Protocol" and fills the form
- **THEN** config.A2A fields are updated with form values
