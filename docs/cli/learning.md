# Learning Commands

Commands for inspecting the learning and knowledge system configuration. See the [Learning System](../features/learning.md) section for detailed documentation.

```
lango learning <subcommand>
```

---

## lango learning status

Show the learning and knowledge system configuration, including knowledge store settings, error correction, graph learning, and embedding/RAG status.
The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango learning status [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango learning status
Learning Status
  Knowledge Enabled:       true
  Error Correction:        true
  Confidence Threshold:    0.7
  Max Context/Layer:       3
  Analysis Turn Threshold: 5
  Analysis Token Threshold:2000

Graph Learning
  Graph Enabled:           true
  Graph Backend:           bolt

Embedding & RAG
  Embedding Provider:      openai
  Embedding Model:         text-embedding-3-small
  RAG Enabled:             true
```

### Output Fields

| Field | Description |
|-------|-------------|
| Knowledge Enabled | Whether the knowledge store is active |
| Error Correction | Whether learned fixes are auto-applied on tool errors |
| Confidence Threshold | Minimum confidence (0.7) for auto-applying a learned fix |
| Max Context/Layer | Maximum context entries retrieved per knowledge layer |
| Analysis Turn Threshold | Number of turns before triggering conversation analysis |
| Analysis Token Threshold | Token count before triggering conversation analysis |
| Graph Enabled | Whether the knowledge graph is active for relationship tracking |
| Graph Backend | Graph store backend (e.g., `bolt`) |
| Embedding Provider | Provider used for text embeddings |
| Embedding Model | Embedding model identifier |
| RAG Enabled | Whether retrieval-augmented generation is active |

---

## lango learning history

Show recent learning entries stored by the learning engine.
The command writes through the Cobra command output stream so wrappers and test harnesses can capture table or JSON output directly.

```
lango learning history [--limit N] [--output table|json]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | `20` | Maximum number of entries to show |
| `--output` | string | `table` | Output format (`table` or `json`) |

**Example:**

```bash
$ lango learning history
ID        CATEGORY         TRIGGER                    CONFIDENCE  CREATED
a1b2c3d4  tool_error       tool:exec_shell            0.85        2026-05-14 10:00:00
e5f6g7h8  timeout          tool:browser_navigate      0.72        2026-05-14 09:55:00
i9j0k1l2  user_correction  conversation:go-style      0.90        2026-05-14 09:50:00
```
