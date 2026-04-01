---
title: Knowledge Ontology
---

# Knowledge Ontology

!!! warning "Experimental"
    The knowledge ontology subsystem is experimental. Schema lifecycle governance, P2P exchange, and ACL enforcement may change in future releases.

## Overview

The ontology subsystem provides a typed knowledge framework for structured entity management. It layers on top of the [Knowledge Graph](knowledge-graph.md), adding:

- **Schema governance** -- lifecycle stages for types and predicates (proposed → shadow → active)
- **Truth maintenance** -- temporal facts with source precedence and conflict detection
- **Entity resolution** -- alias-based deduplication with merge/split operations
- **Property store** -- typed entity properties with query support
- **Access control** -- role-based permissions (read/write/admin) per principal
- **P2P exchange** -- cross-network schema and fact synchronization

## Architecture

```
OntologyService (facade)
├── Registry            — type + predicate definitions with schema versioning
├── TruthMaintainer     — temporal fact assertion, conflict detection, auto-resolution
├── EntityResolver      — alias management, merge/split, canonical ID resolution
├── PropertyStore       — typed entity properties, query by type + filters
├── ACLPolicy           — role-based operation gating (read/write/admin)
├── GovernanceEngine    — schema lifecycle FSM, rate limiting, explosion budget
├── ActionExecutor      — custom registered ontology actions
└── Graph Store         — underlying BoltDB triple store
```

All consumers interact through the `OntologyService` interface -- never reference internal components directly.

## Schema Lifecycle

When governance is enabled, types and predicates follow a lifecycle FSM:

```
proposed → quarantined → proposed (retry)
proposed → shadow → active → deprecated
shadow → quarantined (rollback)
```

| Stage | Description |
|-------|-------------|
| **proposed** | Newly registered. Subject to daily rate limits and quarantine period |
| **quarantined** | Held for review. Can return to proposed |
| **shadow** | Available for use, tracked for usage metrics. Duration configurable |
| **active** | Fully operational. Can only transition to deprecated |
| **deprecated** | Terminal state. No outgoing transitions |

**Rate limiting:** Combined daily limit for type + predicate proposals (`governance.maxNewPerDay`). Monthly explosion budget caps total new proposals (`governance.schemaExplosionBudget`).

**Auto-promotion:** Types/predicates in shadow mode are promoted to active when usage count reaches `governance.minUsageForPromotion`.

## Truth Maintenance

Facts are asserted with temporal metadata and source provenance:

```
AssertFact(subject, predicate, object, validFrom, validTo, source, confidence)
```

**Temporal metadata keys:**

| Key | Description |
|-----|-------------|
| `_valid_from` | Fact validity start (RFC3339, empty = epoch) |
| `_valid_to` | Fact validity end (RFC3339, empty = still valid) |
| `_recorded_at` | System time when triple was first recorded |
| `_recorded_by` | Who recorded (agent ID, "human", peer DID) |
| `_source` | Origin category for source precedence |
| `_confidence` | Confidence score (0.0000 ~ 1.0000) |

**Source precedence** (higher = higher priority):

| Source | Priority |
|--------|----------|
| `manual` | 10 |
| `knowledge` | 8 |
| `correction` | 7 |
| `llm_extraction` | 4 |
| `graph_engine` | 3 |
| `memory_hook` | 2 |
| `p2p_exchange` | 1 |

Conflicting facts on the same `(subject, predicate)` are auto-resolved when source precedence differs. Otherwise, conflicts are stored for manual resolution via `ResolveConflict`.

## Entity Resolution

Entities can have multiple aliases pointing to a single canonical ID:

- **DeclareSameAs(A, B)** -- link two entity IDs as the same entity
- **MergeEntities(canonical, duplicate)** -- merge triples from duplicate into canonical
- **SplitEntity(canonical, splitOut)** -- reverse a merge
- **Resolve(rawID)** -- resolve any alias to its canonical ID

All triple queries automatically resolve aliases before searching the graph store.

## Property Store

Entities have typed properties (key-value pairs) stored alongside their graph triples:

- **SetEntityProperty** -- set a property on an entity (with entity type tracking)
- **GetEntityProperties** -- retrieve all properties for an entity
- **QueryEntities** -- search entities by type and property filters
- **GetEntity** -- full entity view (properties + relationships)

## Access Control

When ACL is enabled, operations are gated by permission level:

| Permission | Operations |
|------------|-----------|
| **read** | Query, list, describe, validate |
| **write** | Register types/predicates, assert/retract facts, set properties, import data |
| **admin** | Deprecate, merge, split, resolve conflicts, delete properties |

Roles are assigned per principal in configuration. P2P peers default to the `p2pPermission` level (default: `write`).

## P2P Exchange

When both P2P and ontology are enabled, schemas and facts can be exchanged between peers:

- **Schema exchange** -- export/import type and predicate definitions as bundles
- **Fact exchange** -- assert facts from remote peers with source `p2p_exchange`
- **Trust thresholds** -- configurable minimum peer trust scores for schema vs. fact exchange
- **Auto-import modes** -- `shadow` (default), `governed` (subject to full lifecycle), or `disabled`

## Agent Tools

The ontologist agent has access to 19+ tools:

| Tool | Description |
|------|-------------|
| `ontology_list_types` | List registered object types |
| `ontology_describe_type` | Describe type with properties and predicates |
| `ontology_query_entities` | Search entities by type + property filters |
| `ontology_get_entity` | Get entity details (properties + relationships) |
| `ontology_assert_fact` | Assert a fact with temporal metadata |
| `ontology_retract_fact` | Retract a fact (soft delete) |
| `ontology_list_conflicts` | List open conflicts |
| `ontology_resolve_conflict` | Resolve a conflict by choosing winner |
| `ontology_merge_entities` | Merge duplicate into canonical entity |
| `ontology_facts_at` | Query facts valid at a specific time |
| `ontology_import_json` | Import entities from JSON |
| `ontology_import_csv` | Import entities from CSV |
| `ontology_from_mcp` | Convert MCP tool result to ontology entity |
| `ontology_list_actions` | List registered ontology actions |
| `ontology_promote_type` | Promote a type through lifecycle stages |
| `ontology_promote_predicate` | Promote a predicate through lifecycle stages |
| `ontology_schema_health` | Get schema health report |
| `ontology_type_usage` | Get usage information for a specific type |

Custom actions registered via `ActionRegistry` generate dynamic tools (`ontology_action_*`).

## Configuration

See [Configuration Reference](../configuration.md#ontology) for all configuration keys.

```json
{
  "ontology": {
    "enabled": true,
    "governance": { "enabled": true },
    "acl": { "enabled": true, "roles": { "operator": "write", "librarian": "read" } },
    "exchange": { "enabled": true }
  }
}
```

!!! note "Dependencies"
    The ontology subsystem requires `graph.enabled: true` for the underlying triple store.
