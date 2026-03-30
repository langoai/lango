package ontology

import (
	"time"

	"github.com/google/uuid"
)

// Reserved metadata keys for temporal and provenance fields.
// These live in graph.Triple.Metadata (prefix "_" to avoid collision with user properties).
const (
	MetaValidFrom  = "_valid_from"   // RFC3339 — fact validity start (empty = epoch)
	MetaValidTo    = "_valid_to"     // RFC3339 — fact validity end (empty = still valid)
	MetaRecordedAt = "_recorded_at"  // RFC3339 — system time when triple was first recorded
	MetaRecordedBy = "_recorded_by"  // who recorded (agent ID, "human", peer DID)
	MetaSource     = "_source"       // origin category for source precedence
	MetaConfidence = "_confidence"   // "0.0000" ~ "1.0000"
)

// SourcePrecedence defines priority ordering for source-of-truth auto-resolution.
// Higher value = higher priority. Used by TruthMaintainer.canAutoResolve.
var SourcePrecedence = map[string]int{
	"manual":         10,
	"knowledge":      8,
	"correction":     7,
	"llm_extraction": 4,
	"graph_engine":   3,
	"memory_hook":    2,
}

// SchemaStatus represents the lifecycle state of an ontology schema element.
type SchemaStatus string

const (
	SchemaActive     SchemaStatus = "active"
	SchemaDeprecated SchemaStatus = "deprecated"
)

// Cardinality defines the relationship multiplicity between subject and object types.
type Cardinality string

const (
	OneToOne   Cardinality = "one_to_one"
	OneToMany  Cardinality = "one_to_many"
	ManyToOne  Cardinality = "many_to_one"
	ManyToMany Cardinality = "many_to_many"
)

// PropertyType defines the data type of an ObjectType property.
type PropertyType string

const (
	TypeString    PropertyType = "string"
	TypeInt       PropertyType = "int"
	TypeFloat     PropertyType = "float"
	TypeBool      PropertyType = "bool"
	TypeDateTime  PropertyType = "datetime"
	TypeReference PropertyType = "reference"
)

// ConstraintKind identifies the type of validation constraint.
type ConstraintKind string

const (
	ConstraintMin   ConstraintKind = "min"
	ConstraintMax   ConstraintKind = "max"
	ConstraintEnum  ConstraintKind = "enum"
	ConstraintRegex ConstraintKind = "regex"
)

// Constraint defines a validation rule for a property.
type Constraint struct {
	Kind  ConstraintKind `json:"kind"`
	Value string         `json:"value"`
}

// PropertyDef defines a single property on an ObjectType.
type PropertyDef struct {
	Name        string       `json:"name"`
	Type        PropertyType `json:"type"`
	Required    bool         `json:"required"`
	Indexed     bool         `json:"indexed"`
	Constraints []Constraint `json:"constraints,omitempty"`
}

// ObjectType represents a formal entity type in the ontology.
type ObjectType struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Properties  []PropertyDef `json:"properties"`
	Extends     string        `json:"extends,omitempty"`
	Status      SchemaStatus  `json:"status"`
	Version     int           `json:"version"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
}

// PredicateDefinition represents a formal relationship type in the ontology.
type PredicateDefinition struct {
	ID          uuid.UUID    `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	SourceTypes []string     `json:"sourceTypes"`
	TargetTypes []string     `json:"targetTypes"`
	Cardinality Cardinality  `json:"cardinality"`
	Inverse     string       `json:"inverse,omitempty"`
	Status      SchemaStatus `json:"status"`
	Version     int          `json:"version"`
	CreatedAt   time.Time    `json:"createdAt"`
	UpdatedAt   time.Time    `json:"updatedAt"`
}
