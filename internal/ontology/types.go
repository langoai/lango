package ontology

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Permission represents an ordered access level for ontology operations.
// Higher values include all lower permissions: Admin > Write > Read.
type Permission int

const (
	PermRead  Permission = iota + 1 // query, list, validate
	PermWrite                        // register, assert, retract, set property
	PermAdmin                        // deprecate, merge, split, resolve conflict, delete
)

// ErrPermissionDenied is returned when a principal lacks the required permission.
var ErrPermissionDenied = errors.New("ontology: permission denied")

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
	"p2p_exchange":   1,
}

// SchemaStatus represents the lifecycle state of an ontology schema element.
type SchemaStatus string

const (
	SchemaProposed    SchemaStatus = "proposed"
	SchemaQuarantined SchemaStatus = "quarantined"
	SchemaShadow      SchemaStatus = "shadow"
	SchemaActive      SchemaStatus = "active"
	SchemaDeprecated  SchemaStatus = "deprecated"
)

// GovernancePolicy configures schema lifecycle governance.
type GovernancePolicy struct {
	MaxNewPerDay          int // combined daily limit for type + predicate proposals
	QuarantinePeriodHrs   int
	ShadowModeDurationHrs int
	MinUsageForPromotion  int
	SchemaExplosionBudget int // per month, type + predicate combined
}

// SchemaHealthReport provides status counts for types and predicates.
type SchemaHealthReport struct {
	Types      map[SchemaStatus]int `json:"types"`
	Predicates map[SchemaStatus]int `json:"predicates"`
}

// TypeUsageInfo provides basic information about a type's status and age.
type TypeUsageInfo struct {
	TypeName  string       `json:"typeName"`
	Status    SchemaStatus `json:"status"`
	Version   int          `json:"version"`
	CreatedAt time.Time    `json:"createdAt"`
}

// --- Slim Wire Types (P2P-portable, no UUID/timestamps/status/version) ---

// SchemaPropertySlim is a wire-format representation of PropertyDef.
type SchemaPropertySlim struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

// SchemaTypeSlim is a wire-format representation of ObjectType.
type SchemaTypeSlim struct {
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Properties  []SchemaPropertySlim `json:"properties,omitempty"`
	Extends     string               `json:"extends,omitempty"`
}

// SchemaPredicateSlim is a wire-format representation of PredicateDefinition.
type SchemaPredicateSlim struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	SourceTypes []string `json:"sourceTypes,omitempty"`
	TargetTypes []string `json:"targetTypes,omitempty"`
	Cardinality string   `json:"cardinality"`
	Inverse     string   `json:"inverse,omitempty"`
}

// SchemaBundle is the P2P-serializable representation of an ontology schema.
// Uses slim wire types only — no UUID, timestamps, status, or version.
type SchemaBundle struct {
	Version       int                   `json:"version"`       // bundle format version (1)
	SchemaVersion int                   `json:"schemaVersion"` // ontology version at export time
	ExportedAt    time.Time             `json:"exportedAt"`
	ExportedBy    string                `json:"exportedBy"`    // DID or "local"
	Types         []SchemaTypeSlim      `json:"types"`
	Predicates    []SchemaPredicateSlim `json:"predicates"`
	Digest        string                `json:"digest"`        // SHA256(canonical JSON of Types+Predicates)
}

// ImportMode determines how imported schema elements are treated.
type ImportMode string

const (
	ImportShadow   ImportMode = "shadow"   // default: import as shadow status
	ImportGoverned ImportMode = "governed" // governance FSM: import as proposed
	ImportDryRun   ImportMode = "dry_run"  // report diff only, no mutations
)

// ImportOptions controls import behavior.
type ImportOptions struct {
	Mode      ImportMode `json:"mode"`
	SourceDID string     `json:"sourceDID,omitempty"` // DID of the exporting peer
}

// ImportResult reports what happened during schema import.
type ImportResult struct {
	TypesAdded       int      `json:"typesAdded"`
	TypesSkipped     int      `json:"typesSkipped"`
	TypesConflicting []string `json:"typesConflicting,omitempty"`
	PredsAdded       int      `json:"predsAdded"`
	PredsSkipped     int      `json:"predsSkipped"`
	PredsConflicting []string `json:"predsConflicting,omitempty"`
}

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

// FilterOp defines comparison operators for property queries.
type FilterOp string

const (
	FilterEq       FilterOp = "eq"       // exact match
	FilterNeq      FilterOp = "neq"      // not equal
	FilterContains FilterOp = "contains" // substring match
)

// PropertyFilter is a single condition in a PropertyQuery.
type PropertyFilter struct {
	Property string   `json:"property"`
	Op       FilterOp `json:"op"`
	Value    string   `json:"value"`
}

// PropertyQuery selects entities by type and property filters (AND semantics).
type PropertyQuery struct {
	EntityType string           `json:"entityType"` // required
	Filters    []PropertyFilter `json:"filters"`    // AND
	Limit      int              `json:"limit"`      // default 100
	Offset     int              `json:"offset"`
}

// EntityResult combines an entity's properties with its graph relationships.
type EntityResult struct {
	EntityID   string            `json:"entityId"`
	EntityType string            `json:"entityType"`
	Properties map[string]string `json:"properties"`
	Outgoing   []ResultTriple    `json:"outgoing,omitempty"` // subject=entityID
	Incoming   []ResultTriple    `json:"incoming,omitempty"` // object=entityID
}

// ResultTriple is a serializable triple for EntityResult.
// Avoids importing graph in types.go; populated by the service layer from graph.Triple.
type ResultTriple struct {
	Subject     string            `json:"subject"`
	Predicate   string            `json:"predicate"`
	Object      string            `json:"object"`
	SubjectType string            `json:"subjectType,omitempty"`
	ObjectType  string            `json:"objectType,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ActionStatus represents the execution state of an action.
type ActionStatus string

const (
	ActionStarted     ActionStatus = "started"
	ActionCompleted   ActionStatus = "completed"
	ActionFailed      ActionStatus = "failed"
	ActionCompensated ActionStatus = "compensated"
)

// FactEffect records a fact that was asserted by an action.
type FactEffect struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
}

// FactRetraction records a fact that was retracted by an action.
type FactRetraction struct {
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	Reason    string `json:"reason"`
}

// PropertyEffect records a property change made by an action.
type PropertyEffect struct {
	EntityID string `json:"entityID"`
	Property string `json:"property"`
	OldValue string `json:"oldValue,omitempty"`
	NewValue string `json:"newValue"`
}

// ActionEffects captures the side effects produced by an action's Execute phase.
type ActionEffects struct {
	FactsAsserted  []FactEffect     `json:"factsAsserted,omitempty"`
	FactsRetracted []FactRetraction `json:"factsRetracted,omitempty"`
	PropertiesSet  []PropertyEffect `json:"propertiesSet,omitempty"`
}

// ActionResult is the outcome of executing an action.
type ActionResult struct {
	LogID   uuid.UUID    `json:"logID"`
	Status  ActionStatus `json:"status"`
	Effects *ActionEffects `json:"effects,omitempty"`
	Error   string       `json:"error,omitempty"`
}

// ActionSummary provides metadata about a registered action (for listing).
type ActionSummary struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	RequiredPerm Permission        `json:"requiredPerm"`
	ParamSchema  map[string]string `json:"paramSchema"`
}

// ActionLogEntry represents a persisted action execution record.
type ActionLogEntry struct {
	ID         uuid.UUID         `json:"id"`
	ActionName string            `json:"actionName"`
	Principal  string            `json:"principal"`
	Params     map[string]string `json:"params"`
	Status     ActionStatus      `json:"status"`
	Effects    *ActionEffects    `json:"effects,omitempty"`
	Error      string            `json:"error,omitempty"`
	StartedAt  time.Time         `json:"startedAt"`
	CompletedAt *time.Time       `json:"completedAt,omitempty"`
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
