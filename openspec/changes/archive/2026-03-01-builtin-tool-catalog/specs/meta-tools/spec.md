## MODIFIED Requirements

### Requirement: Skill import access control
The `import_skill` tool handler SHALL check `SkillConfig.AllowImport` before processing any import request. When `AllowImport` is false, the handler SHALL return an error indicating skill import is disabled.

#### Scenario: Import blocked when AllowImport is false
- **WHEN** `import_skill` is invoked and `SkillConfig.AllowImport` is `false`
- **THEN** the handler SHALL return error "skill import disabled (skill.allowImport=false)"
- **AND** no import processing SHALL occur

#### Scenario: Import proceeds when AllowImport is true
- **WHEN** `import_skill` is invoked and `SkillConfig.AllowImport` is `true`
- **THEN** the handler SHALL proceed with normal import logic
