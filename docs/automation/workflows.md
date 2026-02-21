# Workflow Engine

!!! warning "Experimental"
    The workflow engine is experimental. APIs and behavior may change in future releases.

DAG-based workflow engine for orchestrating multi-step agent operations. Steps run in parallel when their dependencies allow, and results flow between steps via template variables.

## Workflow Definition

Workflows are defined in YAML files with a list of steps, each assigned to a specific agent:

```yaml
name: code-review-pipeline
description: "Automated PR code review"
deliver_to: [slack]
steps:
  - id: fetch-changes
    agent: operator
    prompt: "Get git diff main...HEAD"

  - id: security-scan
    agent: navigator
    prompt: "Analyze security in: {{fetch-changes.result}}"
    depends_on: [fetch-changes]

  - id: quality-review
    agent: navigator
    prompt: "Review code quality: {{fetch-changes.result}}"
    depends_on: [fetch-changes]

  - id: summary
    agent: planner
    prompt: |
      Security: {{security-scan.result}}
      Quality: {{quality-review.result}}
      Write a review report.
    depends_on: [security-scan, quality-review]
    deliver_to: [slack]
```

In this example, `security-scan` and `quality-review` run in parallel after `fetch-changes` completes, and `summary` waits for both to finish.

## Features

### DAG Execution

Steps are organized into a directed acyclic graph. The engine performs a topological sort to determine execution layers:

1. **Layer 0** -- steps with no dependencies (roots)
2. **Layer 1** -- steps whose dependencies are all in layer 0
3. **Layer N** -- steps whose dependencies are all in previous layers

Within each layer, steps execute in parallel up to the configured concurrency limit.

### Template Variables

Step prompts support `{{step-id.result}}` placeholders that are replaced with the output of completed steps at render time. Step IDs may contain letters, digits, hyphens, and underscores.

```yaml
prompt: "Summarize: {{fetch-data.result}}"
```

If a referenced step has no result available, the engine returns an error.

### State Persistence

Workflow runs and step statuses are persisted via Ent ORM. This enables:

- Querying run history and step-level results
- Tracking workflow progress across steps
- Resuming failed workflows from the last successful step

### Step-Level Delivery

Each step can specify its own `deliver_to` channels, independent of the workflow-level delivery. Step results are delivered with the format `[workflow-name/step-id] result`.

### Cycle Detection

The DAG constructor validates the dependency graph using topological sort. If a circular dependency is detected, the workflow is rejected before execution begins.

## Supported Agents

Workflow steps can be assigned to any of the built-in sub-agents:

| Agent | Role |
|-------|------|
| `operator` | System operations, shell execution, file management |
| `navigator` | Web browsing, research, information gathering |
| `vault` | Cryptographic operations, secrets management |
| `librarian` | Knowledge retrieval, search, learning |
| `automator` | Automation tasks, cron management |
| `planner` | Task planning, coordination, summarization |
| `chronicler` | Memory management, history, session tracking |

## CLI Commands

### Run a Workflow

```bash
lango workflow run --file review-pipeline.yaml
```

### List Runs

```bash
lango workflow list --limit 10
```

### Check Status

```bash
lango workflow status --id <run-id>
```

### Cancel a Run

```bash
lango workflow cancel --id <run-id>
```

### View History

```bash
lango workflow history --limit 20
```

## Configuration

> **Settings:** `lango settings` → Workflow Engine

```json
{
  "workflow": {
    "enabled": true,
    "maxConcurrentSteps": 4,
    "defaultTimeout": "5m",
    "stateDir": "~/.lango/state",
    "defaultDeliverTo": ["telegram"]
  }
}
```

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `workflow.enabled` | `bool` | `false` | Enable the workflow engine |
| `workflow.maxConcurrentSteps` | `int` | `4` | Maximum concurrently executing steps |
| `workflow.defaultTimeout` | `duration` | `5m` | Default timeout for a single step |
| `workflow.stateDir` | `string` | - | Directory for workflow state storage |
| `workflow.defaultDeliverTo` | `[]string` | `[]` | Default delivery channels |

## Architecture

The workflow engine consists of four main components:

- **Engine** (`internal/workflow/engine.go`) -- orchestrates DAG execution, manages concurrency, and handles delivery
- **DAG** (`internal/workflow/dag.go`) -- builds and validates the dependency graph, provides topological sort and ready-step queries
- **Template** (`internal/workflow/template.go`) -- renders `{{step-id.result}}` placeholders in prompts
- **StateStore** (`internal/workflow/state.go`) -- Ent ORM persistence for run and step records
