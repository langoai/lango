## ADDED Requirements

### Requirement: P2P embedded skills
The system SHALL provide 8 embedded skills for P2P operations, each using `type: script` with `status: active` and mapping to a `lango p2p` CLI command.

#### Scenario: All P2P skills present
- **WHEN** the skills directory is scanned
- **THEN** the following skill directories exist with valid SKILL.md files: p2p-status, p2p-peers, p2p-connect, p2p-disconnect, p2p-discover, p2p-identity, p2p-firewall-list, p2p-firewall-add

### Requirement: Skill format consistency
Each P2P skill SKILL.md SHALL follow the existing skill format with YAML frontmatter (name, description, type, status) and a shell code block with the corresponding CLI command.

#### Scenario: Skill file structure
- **WHEN** any P2P SKILL.md file is parsed
- **THEN** it contains valid YAML frontmatter with `type: script` and `status: active`, and a shell code block executing `lango p2p <subcommand>`
