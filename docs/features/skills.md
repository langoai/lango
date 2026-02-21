---
title: Skill System
---

# Skill System

The Skill System extends Lango's capabilities through file-based skill definitions. Skills are Markdown files (`SKILL.md`) that define reusable behaviors, scripts, templates, and multi-step workflows -- without writing Go code.

## Skill Types

| Type | Description | Definition |
|------|-------------|------------|
| **instruction** | Reference documents that guide agent reasoning | Markdown content in the body |
| **composite** | Multi-step tool chains executed sequentially | JSON step definitions |
| **script** | Shell scripts executed in a sandboxed environment | `sh`/`bash` code block |
| **template** | Go templates rendered with parameters | `template` code block |

## SKILL.md Format

Every skill is defined in a `SKILL.md` file with YAML frontmatter and a Markdown body:

```markdown
---
name: my-skill
description: What this skill does
type: instruction
status: active
created_by: user
requires_approval: false
allowed-tools: exec file_read
---

The skill content goes here. For instruction skills,
this entire body is the reference document.
```

### Frontmatter Fields

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| `name` | Yes | -- | Unique skill identifier |
| `description` | No | -- | Description shown to the agent |
| `type` | No | `instruction` | Skill type: `instruction`, `composite`, `script`, `template` |
| `status` | No | `active` | Lifecycle status: `draft`, `active`, `disabled` |
| `created_by` | No | -- | Creator attribution |
| `requires_approval` | No | `false` | Whether execution requires user approval |
| `allowed-tools` | No | -- | Space-separated list of pre-approved tools |

### Instruction Skills

The Markdown body is stored as the skill content. When invoked, the agent receives this as a reference document:

```markdown
---
name: coding-guidelines
description: Project coding standards and conventions
type: instruction
---

## Variable Naming
- Use camelCase for local variables
- Use PascalCase for exported functions
...
```

### Script Skills

Shell scripts are defined in an `sh` or `bash` code block:

```markdown
---
name: system-info
description: Gather system information
type: script
---

```sh
echo "OS: $(uname -s)"
echo "CPU: $(nproc) cores"
echo "Memory: $(free -h | awk '/Mem:/{print $2}')"
```
```

!!! warning "Script Safety"

    Scripts are validated against dangerous patterns before execution. The following patterns are blocked:

    | Pattern | Threat |
    |---|---|
    | `rm -rf /` | Filesystem destruction |
    | Fork bomb (`:(){ ...\|...& };:`) | Resource exhaustion |
    | `curl\|sh` or `wget\|sh` | Remote code execution |
    | `> /dev/sd*` | Disk overwrite |
    | `mkfs.*` | Filesystem formatting |
    | `dd if=` | Raw disk write |

    Pattern blocking provides a basic safety net but is not a comprehensive sandbox. Review imported skills before enabling them, especially scripts from untrusted sources.

### Composite Skills

Multi-step tool chains are defined as JSON blocks:

```markdown
---
name: deploy
description: Deploy the application
type: composite
requires_approval: true
---

### Step 1

```json
{
  "tool": "exec",
  "params": {"command": "make build"}
}
```

### Step 2

```json
{
  "tool": "exec",
  "params": {"command": "make deploy"}
}
```
```

### Template Skills

Go templates with parameter substitution:

```markdown
---
name: greeting
description: Generate a greeting message
type: template
---

```template
Hello, {{.name}}! Welcome to {{.project}}.
```

## Parameters

```json
{
  "type": "object",
  "properties": {
    "name": {"type": "string", "description": "User's name"},
    "project": {"type": "string", "description": "Project name"}
  }
}
```
```

## Skill Lifecycle

Skills progress through three statuses:

```
draft → active → disabled
```

- **Draft** -- Created but not loaded into the agent
- **Active** -- Loaded and available as an agent tool (prefixed `skill_`)
- **Disabled** -- Retained in storage but not loaded

## Importing Skills

Skills can be imported from external sources when `skill.allowImport` is enabled.

### From URLs

Import a single skill from a raw URL:

```
Use the skill_import tool to import from https://example.com/skills/my-skill/SKILL.md
```

### From GitHub Repositories

Import skills from GitHub repositories. Each subdirectory containing a `SKILL.md` file is treated as a skill:

```
Import skills from https://github.com/owner/repo
```

The importer supports:

- **Git clone** (preferred) -- Shallow clone with resource file support (`scripts/`, `references/`, `assets/`)
- **GitHub API fallback** -- When `git` is not available, uses the GitHub Contents API

### Bulk Import

Import all skills from a repository directory at once:

```
Import all skills from https://github.com/owner/repo/tree/main/skills
```

Bulk import respects the `maxBulkImport` limit (default: 50) and runs with configurable concurrency (default: 5).

## Default Skills

Lango ships with **30 embedded default skills** that are deployed to `~/.lango/skills/` on first run. Existing skills are never overwritten, so user customizations are preserved.

| Category | Skills |
|---|---|
| **Agent** | `agent-list`, `agent-status` |
| **Config** | `config-create`, `config-delete`, `config-list`, `config-use`, `config-validate` |
| **Cron** | `cron-add`, `cron-delete`, `cron-history`, `cron-list`, `cron-pause`, `cron-resume` |
| **Graph** | `graph-clear`, `graph-query`, `graph-stats`, `graph-status` |
| **Memory** | `memory-clear`, `memory-list`, `memory-status` |
| **Security** | `secrets-list`, `security-status` |
| **Server** | `serve`, `version`, `doctor` |
| **Workflow** | `workflow-cancel`, `workflow-history`, `workflow-list`, `workflow-run`, `workflow-status` |

## Storage

Skills are stored as Markdown files in the skills directory (default: `~/.lango/skills/`). Each skill gets its own subdirectory:

```
~/.lango/skills/
  my-skill/
    SKILL.md
    scripts/
    references/
    assets/
```

Resource directories (`scripts/`, `references/`, `assets/`) are automatically imported alongside the skill definition.

## Configuration

> **Settings:** `lango settings` → Skill

```json
{
  "skill": {
    "enabled": true,
    "skillsDir": "~/.lango/skills",
    "allowImport": true,
    "maxBulkImport": 50,
    "importConcurrency": 5,
    "importTimeout": "2m"
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `enabled` | `bool` | `false` | Enable the skill system |
| `skillsDir` | `string` | `~/.lango/skills` | Directory containing skill files |
| `allowImport` | `bool` | `false` | Enable importing skills from URLs and GitHub |
| `maxBulkImport` | `int` | `50` | Maximum skills in a single bulk import |
| `importConcurrency` | `int` | `5` | Concurrent HTTP requests during import |
| `importTimeout` | `duration` | `2m` | Overall timeout for import operations |

## Agent Integration

Active skills are registered as agent tools with the `skill_` prefix. For example, a skill named `deploy` becomes the tool `skill_deploy`.

The agent uses the skill description to decide when to invoke it. For instruction skills, the agent receives the full reference document when the tool is called.

Skills also feed into the [Knowledge System](knowledge.md) as **Layer 3 (Skill Patterns)**, making them available during context retrieval.

## Related

- [Knowledge System](knowledge.md) -- Skills appear in context layer 3
- [Tool Approval](../security/tool-approval.md) -- Skills with `requires_approval: true`
